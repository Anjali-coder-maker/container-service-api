//go:build default

package main

import (
	"fmt"
	"go-podman-api/config"
	"go-podman-api/handlers"
)

func Initialize() {
	cfg := config.GetConfig()
	fmt.Println("Configuration loaded successfully")

	// Process each service
	for serviceName := range cfg.Services {
		fmt.Printf("Processing service: %s\n", serviceName)

		// Pull the image
		fmt.Printf("Pulling image for service: %s\n", serviceName)
		pullResult := handlers.PullImage(serviceName)
		if pullResult.Error != "" {
			fmt.Printf("Error pulling image for service %s: %s\n", serviceName, pullResult.Error)
			continue
		} else {
			fmt.Printf("Successfully pulled image for service %s\n", serviceName)
		}

		// Check and disable any existing service
		fmt.Printf("Checking and disabling any existing service: %s\n", serviceName)
		if handlers.CheckAndDisableExistingService(serviceName) {
			fmt.Printf("Successfully disabled existing service %s\n", serviceName)
		} else {
			fmt.Printf("No existing service or failed to disable service %s\n", serviceName)
		}

		// Create the unit file
		fmt.Printf("Creating unit file for service: %s\n", serviceName)
		createResult := handlers.CreateUnitFile(serviceName)
		if createResult.Error != "" {
			fmt.Printf("Error creating unit file for service %s: %s\n", serviceName, createResult.Error)
			continue
		} else {
			fmt.Printf("Successfully created unit file for service %s\n", serviceName)
		}

		// Enable and start the service
		fmt.Printf("Enabling and starting service: %s\n", serviceName)
		startResult := handlers.EnableAndStartService(serviceName)
		if startResult.Error != "" {
			fmt.Printf("Error starting service %s: %s\n", serviceName, startResult.Error)
		} else {
			fmt.Printf("Successfully started service %s\n", serviceName)
		}
	}
}

func run() {
	fmt.Println("Starting initialization of default configurations")
	Initialize()
}
