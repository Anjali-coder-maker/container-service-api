//go:build load

package handlers

import (
	"fmt"
	"go-podman-api/utils"
	"runtime"
	"strings"
)

func getCurrentContainerImageID(service string) (string, error) {
	containerName := fmt.Sprintf("%s-service-backend", service)
	resp := utils.ExecuteCommand("podman", "inspect", "--format", "{{.Image}}", containerName)

	if strings.Contains(resp.Output, "no such object") {
		// Container does not exist
		return "", nil
	}

	if resp.Error != "" {
		return "", fmt.Errorf("error inspecting container image for service %s: %v", service, resp.Error)
	}

	return strings.TrimSpace(resp.Output), nil
}

func getLatestImageID(service string) (string, error) {
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
		return "", fmt.Errorf("Not able to pull the latest image for service %s: %v, Please check your internet connection and try again", service, resp.Error)
	}

	// Get the image ID
	resp = utils.ExecuteCommand("podman", "inspect", "--format", "{{.Id}}", imageName)
	if resp.Error != "" {
		return "", fmt.Errorf("error inspecting latest image for service %s: %v", service, resp.Error)
	}

	return strings.TrimSpace(resp.Output), nil
}

func checkAndUpdateService(service string) error {
	currentImageID, err := getCurrentContainerImageID(service)
	if err != nil {
		return fmt.Errorf("error getting current image ID for service %s: %v", service, err)
	}

	latestImageID, err := getLatestImageID(service)
	if err != nil {
		return err
	}

	if currentImageID == "" {
		fmt.Printf("No running container found for service %s. The %s service is disabled. Container image will be automatically updated in the next run.\n", service, service)
		return err
	}

	if currentImageID != latestImageID {
		fmt.Printf("Updating service %s to the latest image\n", service)
		if err := restartService(service); err != nil {
			return fmt.Errorf("error updating and restarting service %s: %v", service, err)
		}
		return nil // Update happened
	}

	return fmt.Errorf("Service %s is already up-to-date no update required", service) // No update happened
}

func restartService(service string) error {
	fullServiceName := fmt.Sprintf("%s-backend.service", service)
	err := DisableService(fullServiceName)
	if err != nil {
		return err
	}

	err = EnableService(fullServiceName)
	if err != nil {
		return err
	}

	fmt.Printf("Service %s updated and restarted successfully\n", service)
	return nil
}

func UpdateServices(filePath string) (bool, error) {
	fmt.Println("Checking for updates...")
	// open the file and read the configurations
	userConfigurations, err := ReadConfigurations(filePath)
	if err != nil {
		return false, fmt.Errorf("Error reading configurations from file %s: %v\n", filePath, err)
	}

	anyUpdates := false

	// check and update the services
	for service := range userConfigurations {
		if err := checkAndUpdateService(service); err != nil {
			fmt.Printf("%s: %v\n", service, err)
		} else {
			anyUpdates = true
		}
	}

	fmt.Println("Update completed")
	return anyUpdates, nil
}
