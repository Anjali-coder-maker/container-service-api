package handlers

import (
	"fmt"
	"go-podman-api/utils"
)

func CheckAndDisableExistingService(imageName string) error {
	serviceFileName := fmt.Sprintf("%s.service", imageName)

	// Check if the service is active
	checkResult := utils.ExecuteCommand("systemctl", "is-active", "--quiet", serviceFileName)
	if checkResult.Error != "" {
		return fmt.Errorf("error checking status of service %s: %s", serviceFileName, checkResult.Error)
	}

	// Service is active, attempt to stop it
	stopResult := utils.ExecuteCommand("systemctl", "stop", serviceFileName)
	if stopResult.Error != "" {
		return fmt.Errorf("failed to stop service %s: %s", serviceFileName, stopResult.Error)
	}

	// Disable the service
	disableResult := utils.ExecuteCommand("systemctl", "disable", serviceFileName)
	if disableResult.Error != "" {
		return fmt.Errorf("failed to disable service %s: %s", serviceFileName, disableResult.Error)
	}

	// Mask the service
	maskResult := utils.ExecuteCommand("systemctl", "mask", serviceFileName)
	if maskResult.Error != "" {
		return fmt.Errorf("failed to mask service %s: %s", serviceFileName, maskResult.Error)
	}

	// Special handling for avahi-daemon.service
	if serviceFileName == "avahi-daemon.service" {
		maskSocketResult := utils.ExecuteCommand("systemctl", "mask", "avahi-daemon.socket")
		if maskSocketResult.Error != "" {
			return fmt.Errorf("failed to mask avahi-daemon.socket: %s", maskSocketResult.Error)
		}
	}

	// Reload the daemon to apply changes
	daemonReloadResult := utils.ExecuteCommand("systemctl", "daemon-reload")
	if daemonReloadResult.Error != "" {
		return fmt.Errorf("failed to reload daemon after disabling service %s: %s", serviceFileName, daemonReloadResult.Error)
	}

	return nil
}
