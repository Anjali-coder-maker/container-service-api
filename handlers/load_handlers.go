//go:build load

package handlers

import (
	"bufio"
	"fmt"
	"go-podman-api/config"
	"go-podman-api/utils"
	"os"
	"runtime"
	"strings"
)

// GetServiceState parses the output of `systemctl list-unit-files` and returns "enabled" or "disabled"
func GetServiceState(service string) (string, error) {
	// Run the systemctl list-unit-files command for the service
	response := utils.ExecuteCommand("systemctl", "list-unit-files", service)
	if response.Error != "" {
		return "", fmt.Errorf("error checking service state: %v", response.Error)
	}

	// Split the output into lines and parse the state
	lines := strings.Split(response.Output, "\n")
	for _, line := range lines {
		// The line should start with the service name, followed by its state (enabled/disabled)
		if strings.HasPrefix(line, service) {
			// Split the line and extract the state (second column)
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1], nil // Return the state (enabled/disabled)
			}
		}
	}

	return "", fmt.Errorf("could not find state for service %s", service)
}

// DisableService stops and disables a service
func DisableService(service string) error {
	state, err := GetServiceState(service)
	if err != nil {
		return fmt.Errorf("error retrieving service state: %v", err)
	}

	if state == "disabled" {
		fmt.Printf("Service %s is already disabled.\n", service)
		return nil
	}

	// Stop the service if it's running
	response := utils.ExecuteCommand("systemctl", "stop", service)
	if response.Error != "" {
		fmt.Printf("Error stopping service %s: %v\n", service, response.Error)
	}

	// Disable the service
	response = utils.ExecuteCommand("systemctl", "disable", service)
	if response.Error != "" {
		return fmt.Errorf("error disabling service %s: %v", service, response.Error)
	}

	fmt.Printf("Disabled %s successfully.\n", service)
	return nil
}

// EnableService enables and starts a service
func EnableService(service string) error {
	state, err := GetServiceState(service)
	if err != nil {
		return fmt.Errorf("error retrieving service state: %v", err)
	}

	if state == "enabled" {
		fmt.Printf("Service %s is already enabled.\n", service)
		return nil
	}

	// Enable the service
	response := utils.ExecuteCommand("systemctl", "enable", service)
	if response.Error != "" {
		return fmt.Errorf("error enabling service %s: %v", service, response.Error)
	}

	// Start the service
	response = utils.ExecuteCommand("systemctl", "start", service)
	if response.Error != "" {
		// Retrieve the journal logs for the service
		journalResponse := utils.ExecuteCommand("journalctl", "-u", service, "--no-pager", "--lines=50")
		if journalResponse.Error != "" {
			return fmt.Errorf("error starting service %s: %v. Additionally, failed to retrieve journal logs: %v", service, response.Error, journalResponse.Error)
		}
		return fmt.Errorf("error starting service %s: %v. Journal logs:\n%s", service, response.Error, journalResponse.Output)
	}

	fmt.Printf("Enabled and started %s successfully.\n", service)
	return nil
}

// PullImageChroot pulls a container image inside a chroot environment
func PullImageChroot(serviceName string, chrootpath string) (string, error) {
	arch := runtime.GOARCH
	var tag string

	// Set the tag based on the architecture
	if arch == "arm" || arch == "arm64" {
		tag = "latest-arm"
	} else {
		tag = "latest-amd"
	}

	// Construct the full image name
	username := "ahaosv1"
	image := fmt.Sprintf("docker.io/%s/%s:%s", username, serviceName, tag)

	response := utils.ExecuteCommand("chroot", chrootpath, "podman", "pull", image)
	if response.Error != "" {
		return "", fmt.Errorf("error pulling image %s: %v", image, response.Error)
	}

	fmt.Printf("Pulled image %s successfully\n", image)
	return image, nil
}

// CreateAndPlaceUnitFile creates a systemd unit file and places it in the chroot directory
func CreateAndPlaceUnitFile(serviceName, chrootDir string, serviceConfig config.ServiceConfig) error {
	unitFileContent := fmt.Sprintf(`[Unit]
Description=%s service
After=network.target

[Service]
ExecStart=%s
ExecStop=%s
ExecStopPost=%s

[Install]
WantedBy=multi-user.target
`, serviceName, serviceConfig.ExecStart, serviceConfig.ExecStop, serviceConfig.ExecStopPost)

	// Create a temporary file to hold the unit file content
	tempUnitFile, err := os.CreateTemp("", fmt.Sprintf("%s-backend.service", serviceName))
	if err != nil {
		return fmt.Errorf("failed to create temporary unit file: %v", err)
	}
	defer os.Remove(tempUnitFile.Name()) // Clean up the temp file
	defer tempUnitFile.Close()

	// Write the unit file content to the temp file
	_, err = tempUnitFile.WriteString(unitFileContent)
	if err != nil {
		return fmt.Errorf("failed to write to temporary unit file: %v", err)
	}

	fmt.Printf("Created temporary unit file for %s at %s\n", serviceName, tempUnitFile.Name())

	// Use the privileged service to move the unit file to /etc/systemd/system/
	response := utils.ExecuteCommand("mv", tempUnitFile.Name(), fmt.Sprintf("/etc/systemd/system/%s-backend.service", serviceName))
	if response.Error != "" {
		return fmt.Errorf("failed to move unit file to /etc/systemd/system/: %v", response.Error)
	}

	fmt.Printf("Moved unit file for %s to /etc/systemd/system/%s-backend.service\n", serviceName, serviceName)
	return nil
}

// MoveOverlayUpperToRoot moves the overlay upper directory to the root directory
func MoveOverlayUpperToRoot() error {
	// Rsync with options to preserve attributes and ensure proper move
	response := utils.ExecuteCommand("rsync", "-aAXv", "/overlay/upper/", "/")
	if response.Error != "" {
		return fmt.Errorf("error moving overlay upper to root: %v", response.Error)
	}

	// Remove the remaining files in the upper directory
	response = utils.ExecuteCommand("rm", "-rf", "/overlay/upper/*")
	if response.Error != "" {
		return fmt.Errorf("error removing files from overlay upper: %v", response.Error)
	}

	fmt.Printf("Moved overlay upper directory to root successfully.\n")
	return nil
}

// Read the configurations from the file
func ReadConfigurations(filePath string) (map[string]bool, error) {
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
