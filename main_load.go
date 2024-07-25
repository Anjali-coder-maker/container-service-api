//go:build load

package main

import (
	"bufio"
	"flag"
	"fmt"
	"go-podman-api/handlers"
	"os"
	"strings"
)

const configDirPath = "/etc/service-manager"
const defaultConfigFilePath = "/etc/service-manager/configuration.conf"

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

	for service, enable := range configurations {
		fullServiceName := service + "-bcknd.service"
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
