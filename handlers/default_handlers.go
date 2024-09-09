//go:build default

package handlers

import (
	"fmt"
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
func CreateUnitFile(imageName string) utils.CommandResponse {
	serviceConfig, exists := cfg.Services[imageName]
	if !exists {
		return utils.CommandResponse{Error: fmt.Sprintf("Service %s not found in configuration", imageName)}
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
	username := "ahaosv1"
	image := fmt.Sprintf("docker.io/%s/%s:%s", username, imageName, tag)

	// Update ExecStart command with the correct image
	execStartCommand := fmt.Sprintf("%s %s", serviceConfig.ExecStart, image)

	unitFileContent := fmt.Sprintf(`[Unit]
Description=Podman container-%s-backend.service
Documentation=man:podman-generate-systemd(1)
Wants=network-online.target
After=network-online.target
RequiresMountsFor=%%t/containers

[Service]
Environment=PODMAN_SYSTEMD_UNIT=%%n
Restart=on-failure
ExecStart=%s
ExecStop=%s
ExecStopPost=%s
TimeoutStopSec=70
Type=simple
NotifyAccess=all

[Install]
WantedBy=multi-user.target
`, imageName, execStartCommand, serviceConfig.ExecStop, serviceConfig.ExecStopPost)

	unitFilePath := fmt.Sprintf("/etc/systemd/system/%s-backend.service", imageName)

	err := os.WriteFile(unitFilePath, []byte(unitFileContent), 0644)
	if err != nil {
		return utils.CommandResponse{Error: err.Error()}
	}
	return utils.CommandResponse{Output: "Unit file created successfully"}
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

	daemonReloadResult := utils.ExecuteCommand("systemctl", "daemon-reload")
	if daemonReloadResult.Error != "" {
		return daemonReloadResult
	}

	return utils.ExecuteCommand("systemctl", "start", serviceFileName)
}
