//go:build load

package main

import (
	"bufio"
	"flag"
	"fmt"
	"go-podman-api/config"
	"go-podman-api/handlers"
	"os"
	"strings"
)

const configDirPath = "/etc/service-manager"
const defaultConfigFilePath = "/etc/service-manager/configuration.conf"
const mergeDirPath = "/overlay/merged"

func run() {
	loadFlag := flag.String("load", defaultConfigFilePath, "Load configuration from /etc/service-manager/configuration.conf. You can specify your own file path also.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  --load <file-path> or -load <file-path>\n")
		fmt.Fprintf(flag.CommandLine.Output(), "      Load configuration from the specified file.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "      If not specified, the default configuration file is %s.\n", defaultConfigFilePath)
		fmt.Fprintf(flag.CommandLine.Output(), "	  Examples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  		service-manager --load /path/to/custom/config/file\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  		service-manager (uses the default configuration file)\n")
	}

	flag.Parse()

	err := createConfigFileIfNotExists()
	if err != nil {
		fmt.Println("Error ensuring configuration file:", err)
		return
	}

	if *loadFlag != "" {
		fmt.Println("Loading configuration from", *loadFlag)
		err := applyConfigFile(*loadFlag)
		if err != nil {
			fmt.Println("Error applying configuration:", err)
		}
	}
}

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

func applyConfigFile(filePath string) error {
	configurations, err := readConfigurations(filePath)
	if err != nil {
		return err
	}

	// get the default services
	cfg := config.GetConfig().Services

	for service, enable := range configurations {
		// check if the service in the configuration file is there in the default services or not
		if _, ok := cfg[service]; !ok {
			fmt.Printf("Service %s is not available in the default services\n", service)
			fmt.Println("Pulling the container in chroot environment")
			// pull the container in chroot environment
			res, image_name := handlers.PullImageChroot(service, mergeDirPath)

			if res.Error != "" {
				fmt.Printf("Error pulling image %s: %v\n", image_name, res.Error)
				continue
			}

			// prepare the serviceConfig
			serviceConfig := config.ServiceConfig{
				Enabled:      enable,
				ExecStart:    fmt.Sprintf("/usr/bin/podman run --name %s-service-backend %s", service, image_name),
				ExecStop:     fmt.Sprintf("/usr/bin/podman stop -t 10 %s-service-backend", service),
				ExecStopPost: fmt.Sprintf("/usr/bin/podman rm -t 10 %s-service-backend", service),
			}

			err = handlers.CreateAndPlaceUnitFile(service, mergeDirPath, serviceConfig)
			if err != nil {
				fmt.Printf("Error creating unit file for service %s: %v\n", service, err)
				continue
			}

			continue
		}

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

	// All the configurations are done now move everything from /overlay/upper to /
	res := handlers.MoveOverlayUpperToRoot()
	if res.Error != "" {
		return fmt.Errorf("Error moving overlay upper to root: %v", res.Error)
	}

	// Now everything is fine
	return nil
}

func readConfigurations(filePath string) (map[string]bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	configurations := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid configuration line: %s", line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1]) == "true"

		if strings.HasPrefix(key, "service.") && strings.HasSuffix(key, ".enable") {
			service := strings.TrimPrefix(key, "service.")
			service = strings.TrimSuffix(service, ".enable")
			configurations[service] = value
		} else {
			return nil, fmt.Errorf("invalid configuration key: %s", key)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return configurations, nil
}
