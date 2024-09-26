//go:build load

package handlers

import (
	"encoding/json"
	"fmt"
	"go-podman-api/utils"
	"runtime"
	"strings"
)

func getCurrentContainerImageDigest(service string) (string, error) {
	arch := runtime.GOARCH
	var tag string

	if arch == "arm" || arch == "arm64" {
		tag = "latest-arm"
	} else {
		tag = "latest-amd"
	}

	imageName := fmt.Sprintf("docker.io/ahaosv1/%s:%s", service, tag)
	resp := utils.ExecuteCommand("podman", "images", "--format", "{{.Digest}}", imageName)

	if strings.Contains(resp.Output, "no such object") {
		// Container does not exist locally
		return "", nil
	}

	if resp.Error != "" {
		return "", fmt.Errorf("error getting local image digest for service %s: %v", service, resp.Error)
	}

	return strings.TrimSpace(resp.Output), nil
}

func getRemoteContainerImageDigest(service string) (string, error) {
	arch := runtime.GOARCH
	var tag string

	if arch == "arm" || arch == "arm64" {
		tag = "latest-arm"
	} else {
		tag = "latest-amd"
	}

	imageName := fmt.Sprintf("docker.io/ahaosv1/%s:%s", service, tag)
	resp := utils.ExecuteCommand("skopeo", "inspect", fmt.Sprintf("docker://%s", imageName))

	if resp.Error != "" {
		return "", fmt.Errorf("error inspecting remote image for service %s: %v", service, resp.Error)
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Output), &result)
	if err != nil {
		return "", fmt.Errorf("error parsing skopeo output for service %s: %v", service, err)
	}

	digest, ok := result["Digest"].(string)
	if !ok {
		return "", fmt.Errorf("no digest found for remote image %s", service)
	}

	return digest, nil
}

func pullLatestImage(service string) (string, error) {
	arch := runtime.GOARCH
	var tag string

	if arch == "arm" || arch == "arm64" {
		tag = "latest-arm"
	} else {
		tag = "latest-amd"
	}

	imageName := fmt.Sprintf("docker.io/ahaosv1/%s:%s", service, tag)
	resp := utils.ExecuteCommand("podman", "pull", imageName)
	if resp.Error != "" {
		return "", fmt.Errorf("not able to pull the latest image for service %s: %v. please check your internet connection and try again", service, resp.Error)
	}

	return fmt.Sprintf("Successfully pulled image %s", imageName), nil
}

func checkAndUpdateService(service string, enabled bool) (bool, error) {
	currentDigest, err := getCurrentContainerImageDigest(service)
	if err != nil {
		return false, fmt.Errorf("error getting current image digest for service %s: %v", service, err)
	}

	remoteDigest, err := getRemoteContainerImageDigest(service)
	if err != nil {
		return false, fmt.Errorf("error getting remote image digest for service %s: %v", service, err)
	}

	// If service is disabled, there might be scenarios where the service unit file is not present so we do not do anything,
	// updates will be done for the enabled services only
	if !enabled {
		// No update required for disabled services with an existing image
		return false, nil
	}

	// If the service is enabled, check if the container needs an update and restart the service if necessary.
	if currentDigest != remoteDigest {
		fmt.Printf("Updating enabled service %s to the latest image\n", service)
		if err := restartService(service); err != nil {
			return false, fmt.Errorf("error updating and restarting service %s: %v", service, err)
		}
		return true, nil // Update occurred
	}

	// No update was required
	return false, nil
}

func restartService(service string) error {
	fullServiceName := fmt.Sprintf("%s-backend.service", service)

	// Disable the service temporarily before updating
	err := DisableService(fullServiceName)
	if err != nil {
		return fmt.Errorf("error disabling service %s: %v", fullServiceName, err)
	}

	// Pull the latest image
	_, err = pullLatestImage(service)
	if err != nil {
		return fmt.Errorf("error pulling the latest image for service %s: %v", service, err)
	}

	// Re-enable and restart the service
	err = EnableService(fullServiceName)
	if err != nil {
		return fmt.Errorf("error enabling service %s: %v", fullServiceName, err)
	}

	fmt.Printf("Service %s updated and restarted successfully.\n", service)
	return nil
}

func UpdateServices(filePath string) (bool, error) {
	fmt.Println("Checking for updates...")
	// open the file and read the configurations
	userConfigurations, err := ReadConfigurations(filePath)
	if err != nil {
		return false, fmt.Errorf("error reading configurations from file %s: %v", filePath, err)
	}

	anyUpdates := false

	// Iterate over the services in the configuration file
	for service, enabled := range userConfigurations {
		updated, err := checkAndUpdateService(service, enabled)
		if err != nil {
			fmt.Printf("%s: %v\n", service, err)
		}

		// Track if any update has occurred
		if updated {
			anyUpdates = true
		}
	}

	if anyUpdates {
		fmt.Println("Update completed, a snapshot is needed.")
	} else {
		fmt.Println("No updates were made, skipping snapshot.")
	}
	return anyUpdates, nil
}
