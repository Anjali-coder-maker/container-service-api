//go:build load

package handlers

import (
	"fmt"
	"go-podman-api/utils"
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
