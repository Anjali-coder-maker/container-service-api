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

// IsServiceRunning checks if a service is active
func IsServiceRunning(service string) bool {
	resp := utils.ExecuteCommand("systemctl", "is-active", service)
	return strings.TrimSpace(resp.Output) == "active"
}

// IsServiceAvailable checks if a service is available
func IsServiceAvailable(service string) bool {
	resp := utils.ExecuteCommand("systemctl", "status", service)
	return strings.Contains(resp.Output, ".service")
}

// DisableService stops and disables a service
func DisableService(service string) error {
	if !IsServiceRunning(service) {
		return fmt.Errorf("service %s is not running", service)
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

// EnableService enables and starts a service
func EnableService(service string) error {
	if IsServiceRunning(service) {
		resp := utils.ExecuteCommand("systemctl", "restart", service)
		if resp.Error != "" {
			return fmt.Errorf("error restarting service %s: %v", service, resp.Error)
		}
		resp = utils.ExecuteCommand("systemctl", "daemon-reload")
		if resp.Error != "" {
			return fmt.Errorf("error reloading daemon after restarting service %s: %v", service, resp.Error)
		}
	} else {
		if !IsServiceAvailable(service) {
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

	resp := utils.ExecuteCommand("chroot", chrootpath, "podman", "pull", image)
	if resp.Error != "" {
		return "", fmt.Errorf("error pulling image %s: %v", image, resp.Error)
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

	unitFilePath := fmt.Sprintf("%s/etc/systemd/system/%s-backend.service", chrootDir, serviceName)

	file, err := os.Create(unitFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(unitFileContent)
	if err != nil {
		return err
	}

	fmt.Printf("Created unit file for %s at %s\n", serviceName, unitFilePath)
	return nil
}

// MoveOverlayUpperToRoot moves the overlay upper directory to the root directory
func MoveOverlayUpperToRoot() error {
	// Rsync with options to preserve attributes and ensure proper move
	resp := utils.ExecuteCommand("rsync", "-aAXv", "/overlay/upper/", "/")
	if resp.Error != "" {
		return fmt.Errorf("error moving overlay upper to root: %v", resp.Error)
	}

	// Remove the remaining files in the upper directory
	resp = utils.ExecuteCommand("rm", "-rf", "/overlay/upper/*")
	if resp.Error != "" {
		return fmt.Errorf("error removing files from overlay upper: %v", resp.Error)
	}

	fmt.Printf("Moved overlay upper directory to root successfully\n")
	return nil
}

// EnableTheNewService enables the new service
func EnableTheNewService(serviceName string) error {
	resp := utils.ExecuteCommand("systemctl", "enable", fmt.Sprintf("%s-backend.service", serviceName))
	if resp.Error != "" {
		return fmt.Errorf("error enabling the new service: %v", resp.Error)
	}

	return nil
}
