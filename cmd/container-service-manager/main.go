package main

import (
	"bufio"
	"flag"
	"fmt"
	"go-podman-api/utils" // assuming utils is the package name for your command execution code
	"os"
	"strings"
)

const configDirPath = "/etc/service-manager"
const defaultConfigFilePath = "/etc/service-manager/configuration.conf"

func main() {
	loadFlag := flag.String("load", "", "Load configuration from the specified file")
	//resetFlag := flag.String("reset", "", "Reset configuration using the specified file")
	flag.Parse()

	err := createConfigFileIfNotExists()
	if err != nil {
		fmt.Println("Error ensuring configuration file:", err)
		return
	}

	// if *resetFlag != "" {
	// 	err := resetConfigFile(*resetFlag)
	// 	if err != nil {
	// 		fmt.Println("Error resetting configuration:", err)
	// 	}
	// }

	if *loadFlag != "" {
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
			err := enableService(fullServiceName)
			if err != nil {
				fmt.Printf("Error enabling service %s: %v\n", fullServiceName, err)
			}
			fmt.Printf("Enabled %s successfully\n", fullServiceName)
		} else {
			err := disableService(fullServiceName)
			if err != nil {
				fmt.Printf("Error disabling service %s: %v\n", fullServiceName, err)
			}
			fmt.Printf("disabled %s successfully\n", fullServiceName)
		}
	}
	return nil
}

// func resetConfigFile(filePath string) error {
// 	configurations, err := readConfigurations(filePath)
// 	if err != nil {
// 		return err
// 	}

// 	for service := range configurations {
// 		fullServiceName := service + "-bcknd.service"
// 		err := disableService(fullServiceName)
// 		if err != nil {
// 			fmt.Printf("disabling service %s: %v\n", fullServiceName, err)
// 		}
// 	}
// 	return nil
// }

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

func enableService(service string) error {
	if isServiceRunning(service) {
		resp := utils.ExecuteCommand("systemctl", "restart", service)
		if resp.Error != "" {
			return fmt.Errorf("error restarting service %s: %v", service, resp.Error)
		}
		resp = utils.ExecuteCommand("systemctl", "daemon-reload")
		if resp.Error != "" {
			return fmt.Errorf("error reloading daemon after restarting service %s: %v", service, resp.Error)
		}
	} else {
		if !isServiceAvailable(service) {
			return fmt.Errorf("service %s can't be started as containerized manner", service)
		}

		resp := utils.ExecuteCommand("systemctl", "enable", service)
		if resp.Error != "" {
			return fmt.Errorf("error enabling service %s: %v", service, resp.Error)
		}
		resp = utils.ExecuteCommand("systemctl", "start", service)
		if resp.Error != "" {
			return fmt.Errorf("error starting service %s: %v", service, resp.Error)
		}
		resp = utils.ExecuteCommand("systemctl", "daemon-reload")
		if resp.Error != "" {
			return fmt.Errorf("error reloading daemon after starting service %s: %v", service, resp.Error)
		}
	}
	return nil

}

func disableService(service string) error {
	// if !isServiceRunning(service) {
	// 	return nil
	// }

	// resp := utils.ExecuteCommand("systemctl", "stop", service)
	// if resp.Error != "" {
	// 	return fmt.Errorf("error stopping service %s: %v", service, resp.Error)
	// }
	// resp = utils.ExecuteCommand("systemctl", "disable", service)
	// if resp.Error != "" {
	// 	return fmt.Errorf("error disabling service %s: %v", service, resp.Error)
	// }
	// resp = utils.ExecuteCommand("systemctl", "reset-failed", service)
	// if resp.Error != "" {
	// 	return fmt.Errorf("error resetting service %s: %v", service, resp.Error)
	// }
	// resp = utils.ExecuteCommand("systemctl", "daemon-reload")
	// if resp.Error != "" {
	// 	return fmt.Errorf("error reloading daemon after disabling service %s: %v", service, resp.Error)
	// }
	// return nil

	if !isServiceRunning(service) {
		return nil
	}

	resp := utils.ExecuteCommand("systemctl", "stop", service)
	if resp.Error != "" {
		return fmt.Errorf("error stopping service %s: %v", service, resp.Error)
	}
	resp = utils.ExecuteCommand("systemctl", "disable", service)
	if resp.Error != "" {
		return fmt.Errorf("error disabling service %s: %v", service, resp.Error)
	}
	resp = utils.ExecuteCommand("systemctl", "reset-failed")
	if resp.Error != "" {
		fmt.Printf("Warning: error resetting service %s: %v\n", service, resp.Error)
	}
	resp = utils.ExecuteCommand("systemctl", "daemon-reload")
	if resp.Error != "" {
		return fmt.Errorf("error reloading daemon after disabling service %s: %v", service, resp.Error)
	}
	return nil

}

func isServiceRunning(service string) bool {
	resp := utils.ExecuteCommand("systemctl", "is-active", service)
	return strings.TrimSpace(resp.Output) == "active"
}

func isServiceAvailable(service string) bool {
	resp := utils.ExecuteCommand("systemctl", "status", service)
	return strings.Contains(resp.Output, ".service")
}
