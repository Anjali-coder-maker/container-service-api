//go:build load

package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"go-podman-api/config"
	"go-podman-api/utils"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetServiceState parses the output of `systemctl list-unit-files` and returns "enabled", "disabled", or "not-found"
func GetServiceState(service string) (string, error) {
	// Run the systemctl list-unit-files command for the service
	response := utils.ExecuteCommand("systemctl", "list-unit-files", service)

	// Handle cases where systemctl exits with an error but lists "0 unit files found"
	if response.Error != "" {
		if strings.Contains(response.Output, "0 unit files listed") {
			// Return "not-found" if no unit file exists for the service
			return "not-found", nil
		}
		// If another error occurred, return it
		return "", fmt.Errorf("error checking service state: %v", response.Error)
	}

	// Split the output into lines and parse the state
	lines := strings.Split(response.Output, "\n")
	for _, line := range lines {
		// The line should start with the service name, followed by its state (enabled/disabled)
		if strings.HasPrefix(line, service) {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1], nil // Return the state (enabled/disabled)
			}
		}
	}

	// Return not-found if the service is not listed
	return "not-found", nil
}

// DisableService stops and disables a service
func DisableService(service string) error {
	state, err := GetServiceState(service)
	if err != nil {
		return fmt.Errorf("error retrieving service state: %v", err)
	}

	// Handle the case where the service unit file is not found
	if state == "not-found" {
		fmt.Printf("Service %s does not exist. No action needed.\n", service)
		return nil
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

	// Handle the case where the service unit file is not found
	if state == "not-found" {
		fmt.Printf("Service %s does not exist. Please create the unit file first.\n", service)
		return nil
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
	if !serviceConfig.Enabled {
		return fmt.Errorf("service %s is not enabled", serviceName)
	}

	// Determine the system architecture
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
	username := "ahaosv1"
	image := fmt.Sprintf("docker.io/%s/%s:%s", username, serviceName, tag)

	// Prepare the template
	tmpl, err := template.New("unitFile").Parse(serviceConfig.UnitFile)
	if err != nil {
		return fmt.Errorf("error parsing unit file template: %v", err)
	}

	// Data to inject into the template
	data := struct {
		ImageName string
	}{
		ImageName: image,
	}

	// Execute the template
	var unitFileBuffer bytes.Buffer
	if err := tmpl.Execute(&unitFileBuffer, data); err != nil {
		return fmt.Errorf("error executing unit file template: %v", err)
	}

	unitFileContent := unitFileBuffer.String()

	// Define the target path within the chroot directory with the 'backend' suffix
	unitFilePath := filepath.Join(chrootDir, "etc/systemd/system", fmt.Sprintf("%s-backend.service", serviceName))

	// Ensure the target directory exists
	err = os.MkdirAll(filepath.Dir(unitFilePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directories for unit file: %v", err)
	}

	// Write the unit file content
	err = os.WriteFile(unitFilePath, []byte(unitFileContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write unit file: %v", err)
	}

	fmt.Printf("Placed unit file for %s at %s\n", serviceName, unitFilePath)
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
