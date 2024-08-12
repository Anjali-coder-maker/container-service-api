//go:build load

package main

import (
	"flag"
	"fmt"
	"go-podman-api/config"
	"go-podman-api/handlers"
	"go-podman-api/utils"
	"os"
	"strings"
)

const configDirPath = "/etc/service-manager"
const defaultConfigFilePath = "/etc/service-manager/configuration.conf"
const mergeDirPath = "/overlay/merged"
const snapshotDir = "/mnt/snapshots"

func run() {
	loadFlag := flag.String("load", "", "Load configuration from /etc/service-manager/configuration.conf. You can specify your own file path also.")
	updateFlag := flag.Bool("update", false, "Update the services based on the configuration file.")
	revertFlag := flag.Bool("rollback", false, "Revert to the previous state")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  --load <file-path> or -load <file-path>\n")
		fmt.Fprintf(flag.CommandLine.Output(), "      Load configuration from the specified file.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "      If not specified, the default configuration file is %s.\n", defaultConfigFilePath)
		fmt.Fprintf(flag.CommandLine.Output(), "	  Examples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  		service-manager --load /path/to/custom/config/file\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  		service-manager (uses the default configuration file)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  --update or -update\n")
		fmt.Fprintf(flag.CommandLine.Output(), "      Update the services based on the configuration file.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "      If not specified, the services are not updated.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "	  Examples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  		service-manager --update\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  		service-manager (services are not updated)\n")
	}

	flag.Parse()

	err := createConfigFileIfNotExists()
	if err != nil {
		fmt.Println("Error ensuring configuration file:", err)
		return
	}

	rootDevicePath := handlers.GetRootDevicePath()
	if rootDevicePath == "" {
		fmt.Println("Error: Unable to find root device path.")
		return
	}

	fmt.Printf("Mounting root device %s to /mnt\n", rootDevicePath)
	err = handlers.MountDisk(rootDevicePath)
	if err != nil {
		fmt.Println("Error mounting disk:", err)
		return
	}

	if *loadFlag != "" && *updateFlag {
		fmt.Println("Both --load and --update flags cannot be used together.")
		flag.Usage()
		return
	}

	if *loadFlag != "" {
		if !handlers.IsConfigurationChanged() {
			fmt.Println("No configuration changes detected. Exiting.")
			return
		}

		fmt.Println("Managing existing snapshots")
		err = handlers.ManageSnapshots()
		if err != nil {
			fmt.Println("Error managing snapshots:", err)
			return
		}

		fmt.Println("Loading and applying configuration from", *loadFlag)
		err := applyConfigFile(*loadFlag)
		if err != nil {
			fmt.Println("Error applying configuration:", err)
			return
		}

		fmt.Println("Creating a new snapshot of the current state")
		err = handlers.CreateNewSnapshot()
		if err != nil {
			fmt.Println("Error creating new snapshot:", err)
			return
		}

	} else if *updateFlag {
		fmt.Println("Updating services based on the configuration file")
		updatesMade, err := handlers.UpdateServices(defaultConfigFilePath)
		if err != nil {
			fmt.Println("Error updating services:", err)
		}

		if updatesMade {
			fmt.Println("Managing existing snapshots")
			err = handlers.ManageSnapshots()
			if err != nil {
				fmt.Println("Error managing snapshots:", err)
				return
			}

			fmt.Println("Creating a new snapshot of the current state")
			err = handlers.CreateNewSnapshot()
			if err != nil {
				fmt.Println("Error creating new snapshot:", err)
				return
			}
		} else {
			fmt.Println("No updates were made; skipping snapshot management.")
		}
		return
	} else if *revertFlag {
		fmt.Println("Reverting to the previous state")
		err = handlers.RevertToPreviousState()
		if err != nil {
			fmt.Println("Error reverting to previous state:", err)
			return
		}
	}

	fmt.Println("Unmounting the disk from /mnt")
	err = handlers.UnmountDisk()
	if err != nil {
		fmt.Println("Error unmounting disk:", err)
		return
	}
}

// Create the configuration file if it does not exist
func createConfigFileIfNotExists() error {
	err := os.MkdirAll(configDirPath, 0755)
	if err != nil {
		return err
	}

	if _, err := os.Stat(defaultConfigFilePath); os.IsNotExist(err) {
		file, err := os.Create(defaultConfigFilePath)
		if err != nil {
			return err
		}
		defer file.Close()

		defaultContent := `# Define services in the following format:		
# service.<service_name>.enable = true|false
`
		_, err = file.WriteString(defaultContent)
		return err
	}
	return nil
}

// Apply the configuration file to the services
func applyConfigFile(filePath string) error {
	userConfigurations, err := handlers.ReadConfigurations(filePath)
	if err != nil {
		return err
	}

	// Get the default services and registry templates
	defaultCfg := config.GetConfig().Services
	registryTemplates := config.GetRegistryTemplates().Services

	for service, enable := range userConfigurations {
		// Check if the service is in the default configuration
		if _, ok := defaultCfg[service]; ok {
			fullServiceName := service + "-backend.service"
			if enable {
				err := handlers.EnableService(fullServiceName)
				if err != nil {
					fmt.Printf("Error enabling service %s: %v\n", fullServiceName, err)
				} else {
					fmt.Printf("Enabled %s successfully\n", fullServiceName)
				}
			} else {
				err := handlers.DisableService(fullServiceName)
				if err != nil {
					fmt.Printf("Error disabling service %s: %v\n", fullServiceName, err)
				} else {
					fmt.Printf("Disabled %s successfully\n", fullServiceName)
				}
			}
			continue
		}

		// Service not found in default configuration, check if the image is already present
		imageName := fmt.Sprintf("docker.io/ahaosv1/%s", service)
		if !isImagePresent(imageName) {
			fmt.Printf("Service %s is not available in the default services and image not found locally\n", service)
			fmt.Println("Pulling the container in chroot environment")

			// Pull the container in chroot environment
			pulledImageName, res := handlers.PullImageChroot(service, mergeDirPath)
			if res != nil {
				fmt.Printf("Error pulling image %s: %v\n", pulledImageName, res.Error)
				continue
			}
			imageName = pulledImageName

			// Check if the service is in the registry templates
			template, exists := registryTemplates[service]
			if !exists {
				fmt.Printf("No template found for service %s\n", service)
				continue
			}

			// Prepare the serviceConfig using the registry template
			serviceConfig := config.ServiceConfig{
				Enabled:      enable,
				ExecStart:    fmt.Sprintf("%s %s", template.ExecStart, imageName),
				ExecStop:     template.ExecStop,
				ExecStopPost: template.ExecStopPost,
			}

			err = handlers.CreateAndPlaceUnitFile(service, mergeDirPath, serviceConfig)
			if err != nil {
				fmt.Printf("Error creating unit file for service %s: %v\n", service, err)
				continue
			}

			// Move everything from /overlay/upper to /
			res = handlers.MoveOverlayUpperToRoot()
			if res != nil {
				fmt.Printf("Error moving overlay upper directory to root: %v\n", res.Error)
				continue
			}
		} else {
			fmt.Printf("Image for service %s found locally, skipping pull step\n", service)
		}

		// Enable or disable the service based on the user configuration
		fullServiceName := service + "-backend.service"
		if enable {
			err := handlers.EnableService(fullServiceName)
			if err != nil {
				fmt.Printf("Error enabling service %s: %v\n", fullServiceName, err)
			} else {
				fmt.Printf("Enabled %s successfully\n", fullServiceName)
			}
		} else {
			err := handlers.DisableService(fullServiceName)
			if err != nil {
				fmt.Printf("Error disabling service %s: %v\n", fullServiceName, err)
			} else {
				fmt.Printf("Disabled %s successfully\n", fullServiceName)
			}
		}
	}
	return nil
}

// Check if the image is present locally
func isImagePresent(imageName string) bool {
	resp := utils.ExecuteCommand("podman", "images", "-q", imageName)
	outputLines := strings.Split(resp.Output, "\n")
	for _, line := range outputLines {
		trimmedLine := strings.TrimSpace(line)
		// Check if the line is a valid hexadecimal string
		if isHex(trimmedLine) && len(trimmedLine) > 0 {
			return true
		}
	}
	return false
}

func isHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
