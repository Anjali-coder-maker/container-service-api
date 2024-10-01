//go:build default

package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"runtime"

	"go-podman-api/config" // ensure that config init function loads the config file
	"go-podman-api/utils"  // ensures the login also
)

var (
	cfg config.Config
)

// init function to load the configuration once
func init() {
	cfg = config.GetConfig()
}

// PullImage handles pulling an image from a registry based on the configuration file
func PullImage(imageName string) utils.CommandResponse {
	serviceConfig, exists := cfg.Services[imageName]
	if !exists {
		return utils.CommandResponse{Error: fmt.Sprintf("Service %s not found in configuration", imageName)}
	}

	if !serviceConfig.Enabled {
		return utils.CommandResponse{Output: fmt.Sprintf("Service %s is disabled in configuration", imageName)}
	}

	arch := runtime.GOARCH
	var tag string

	// Set the tag based on the architecture
	if arch == "arm" || arch == "arm64" {
		tag = "latest-arm"
	} else {
		tag = "latest-amd"
	}

	// Construct the full image name
	username := utils.Getenvmap()["DOCKER_USERNAME"]
	image := fmt.Sprintf("docker.io/%s/%s:%s", username, imageName, tag)
	result := utils.ExecuteCommand("podman", "pull", image)

	return result
}

// CreateUnitFile handles creating the systemd unit file for the container
func CreateUnitFile(serviceName string) utils.CommandResponse {
	serviceConfig, exists := cfg.Services[serviceName]
	if !exists {
		return utils.CommandResponse{Error: fmt.Sprintf("Service %s not found in configuration", serviceName)}
	}

	if !serviceConfig.Enabled {
		return utils.CommandResponse{Error: fmt.Sprintf("Service %s is not enabled", serviceName)}
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
		return utils.CommandResponse{Error: fmt.Sprintf("Error parsing unit file template: %v", err)}
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
		return utils.CommandResponse{Error: fmt.Sprintf("Error executing unit file template: %v", err)}
	}

	unitFileContent := unitFileBuffer.String()

	// Define the unit file path with the 'backend' suffix
	unitFilePath := fmt.Sprintf("/etc/systemd/system/%s-backend.service", serviceName)

	// Write the unit file content to the systemd unit file
	err = os.WriteFile(unitFilePath, []byte(unitFileContent), 0644)
	if err != nil {
		return utils.CommandResponse{Error: fmt.Sprintf("Error writing unit file: %v", err)}
	}

	// Reload systemd to recognize the new unit file
	response := utils.ExecuteCommand("systemctl", "daemon-reload")
	if response.Error != "" {
		return utils.CommandResponse{Error: response.Error}
	}

	return utils.CommandResponse{Output: "Unit file created, and daemon-reload successfully"}
}

// checkAndDisableExistingService checks if a service is active and disables it if necessary
func CheckAndDisableExistingService(imageName string) bool {
	serviceFileName := fmt.Sprintf("%s.service", imageName)

	// Check if the service is active
	checkResult := utils.ExecuteCommand("systemctl", "is-active", "--quiet", serviceFileName)
	// if there is any error then return for that service
	if checkResult.Error != "" {
		return false
	}

	// Service is active, disable it
	stopResult := utils.ExecuteCommand("systemctl", "stop", serviceFileName)
	if stopResult.Error != "" {
		log.Printf("Failed to stop service %s: %s\n", serviceFileName, stopResult.Error)
	}

	disableResult := utils.ExecuteCommand("systemctl", "disable", serviceFileName)
	if disableResult.Error != "" {
		log.Printf("Failed to disable service %s: %s\n", serviceFileName, disableResult.Error)
	}

	if !cfg.Services[imageName].Privileged {
		maskResult := utils.ExecuteCommand("systemctl", "mask", serviceFileName)
		if maskResult.Error != "" {
			log.Printf("Failed to mask service %s: %s\n", serviceFileName, maskResult.Error)
		}
	}

	daemonReloadResult := utils.ExecuteCommand("systemctl", "daemon-reload")
	if daemonReloadResult.Error != "" {
		log.Printf("Failed to reload daemon after disabling service %s: %s\n", serviceFileName, daemonReloadResult.Error)
	}

	return true
}

// EnableAndStartService handles enabling and starting the systemd service
func EnableAndStartService(imageName string) utils.CommandResponse {
	serviceFileName := fmt.Sprintf("%s-backend.service", imageName)

	enableResult := utils.ExecuteCommand("systemctl", "enable", serviceFileName)
	if enableResult.Error != "" {
		return enableResult
	}

	return utils.ExecuteCommand("systemctl", "start", serviceFileName)
}
