//go:build load

package main

import (
	"flag"
	"fmt"
	"go-podman-api/config"
	"go-podman-api/handlers"
	"go-podman-api/utils"
	"os"
	"runtime"
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

		// fmt.Println("Managing existing snapshots")
		// err = handlers.ManageSnapshots()
		// if err != nil {
		// 	fmt.Println("Error managing snapshots:", err)
		// 	return
		// }

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
			// fmt.Println("Managing existing snapshots")
			// err = handlers.ManageSnapshots()
			// if err != nil {
			// 	fmt.Println("Error managing snapshots:", err)
			// 	return
			// }

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
	// Read the user-provided configurations from the given file
	userConfigurations, err := handlers.ReadConfigurations(filePath)
	if err != nil {
		return fmt.Errorf("error reading configurations from file %s: %v", filePath, err)
	}

	var daemonReloadNeeded bool
	// Get the registry templates (which contains all services, including defaults)
	registryTemplates := config.GetRegistryTemplates().Services

	for service, enable := range userConfigurations {
		// Check if the service is in the registry templates
		template, exists := registryTemplates[service]
		if !exists {
			fmt.Printf("No service found in registry with name %s\n", service)
			continue
		}

		fullServiceName := service + "-backend.service"

		if enable {
			// If the service is required to be enabled, check if the image exists
			arch := runtime.GOARCH
			var tag string

			// Set the tag based on the architecture
			switch arch {
			case "arm", "arm64":
				tag = "latest-arm"
			default:
				tag = "latest-amd"
			}

			// Construct the full image name
			imageName := fmt.Sprintf("docker.io/ahaosv1/%s:%s", service, tag)

			if !isImagePresent(imageName) {
				// Image not found, pulling the container and preparing the service
				fmt.Printf("Image for service %s not found locally, pulling the image\n", service)
				fmt.Println("from applyConfigFile", mergeDirPath)

				// Pull the container in the chroot environment
				imageName, err := handlers.PullImageChroot(service, mergeDirPath)
				if err != nil {
					fmt.Printf("Error pulling image %s: %v\n", imageName, err)
					continue
				}

				// Create a serviceConfig using the registry template and dynamic imageName
				serviceConfig := config.ServiceConfig{
					Enabled:    enable,
					UnitFile:   template.UnitFile, // Keep the original unit file template
					Privileged: template.Privileged,
				}

				// Place the service unit file in the appropriate location
				err = handlers.CreateAndPlaceUnitFile(service, mergeDirPath, serviceConfig)
				if err != nil {
					fmt.Printf("Error creating unit file for service %s: %v\n", service, err)
					continue
				}

				// Move everything from /overlay/upper to /
				err = handlers.MoveOverlayUpperToRoot()
				if err != nil {
					fmt.Printf("Error moving overlay upper directory to root: %v\n", err)
					continue
				}

				daemonReloadNeeded = true
			} else {
				fmt.Printf("Image for service %s found locally, skipping pull step\n", service)
			}

			// Enable the service using the full service name with the 'backend' suffix
			err := handlers.EnableService(fullServiceName)
			if err != nil {
				fmt.Printf("Error enabling service %s: %v\n", fullServiceName, err)
			}
		} else {
			// If the service is not required to be enabled, disable it
			err := handlers.DisableService(fullServiceName)
			if err != nil {
				fmt.Printf("Error disabling %s: %v\n", fullServiceName, err)
			}
		}
	}

	// After all operations, reload the systemd daemon if needed
	if daemonReloadNeeded {
		resp := utils.ExecuteCommand("systemctl", "daemon-reload")
		if resp.Error != "" {
			return fmt.Errorf("error reloading daemon: %v", resp.Error)
		}
		fmt.Println("Systemd daemon reloaded.")
	}

	fmt.Println("Configuration applied successfully")
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
