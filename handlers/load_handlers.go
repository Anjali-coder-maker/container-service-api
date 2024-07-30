//go:build load

package handlers

import (
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

func PullImageChroot(serviceName string, chrootpath string) (utils.CommandResponse, string) {
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
		return utils.CommandResponse{Error: resp.Error}, ""
	}
	return utils.CommandResponse{Output: "Image pulled and saved successfully"}, image
}

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

func MoveOverlayUpperToRoot() utils.CommandResponse {
	resp := utils.ExecuteCommand("rsync", "-a", "/overlay/upper/", "/")
	if resp.Error != "" {
		return utils.CommandResponse{Error: fmt.Sprintf("Error moving overlay upper to root: %v", resp.Error)}
	}

	// restart the system
	resp = utils.ExecuteCommand("reboot")
	if resp.Error != "" {
		return utils.CommandResponse{Error: fmt.Sprintf("Error restarting the system: %v", resp.Error)}
	}
	return utils.CommandResponse{Output: "System restarted successfully"}
}
