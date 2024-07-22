//go:build default

package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"go-podman-api/utils"
	"log"
	"runtime"
	"strings"
	"sync"
)

var (
	//go:embed default-config.json
	defaultConfig embed.FS

	// config stores the configuration data
	configData Config

	// once ensures the configuration is loaded only once
	once sync.Once
)

// ServiceConfig represents the configuration for a service
type ServiceConfig struct {
	Enabled      bool   `json:"enabled"`
	ExecStart    string `json:"exec_start"`
	ExecStop     string `json:"exec_stop"`
	ExecStopPost string `json:"exec_stop_post"`
}

// Config represents the configuration file structure
type Config struct {
	Services map[string]ServiceConfig `json:"services"`
}

// init is called automatically to load the configuration file
func init() {
	loadConfiguration()
}

// loadConfiguration reads the embedded configuration file and stores the values in configData
func loadConfiguration() {
	once.Do(func() {
		log.Println("Loading configuration from embedded default-config.json")
		defaultConfigFile, err := defaultConfig.ReadFile("default-config.json")
		if err != nil {
			log.Fatalf("Error reading default configuration file: %v", err)
		}

		if err := json.Unmarshal(defaultConfigFile, &configData); err != nil {
			log.Fatalf("Error unmarshalling default configuration file: %v", err)
		}

		// Update exec_start commands based on architecture and environment variables
		updateExecStartCommands()

		log.Println("Configuration loaded successfully")
	})
}

// updateExecStartCommands updates the exec_start commands in the configuration based on architecture and environment variables
func updateExecStartCommands() {
	arch := runtime.GOARCH
	var usernameEnv, tagSuffix string

	switch arch {
	case "arm", "arm64":
		usernameEnv = "DOCKER_USERNAME_ARM"
		tagSuffix = "latest-arm"
	default:
		usernameEnv = "DOCKER_USERNAME_AMD"
		tagSuffix = "latest"
	}

	username := utils.Getenvmap()[usernameEnv]
	if username == "" {
		log.Fatalf("Environment variable %s not set", usernameEnv)
	}

	for serviceName, serviceConfig := range configData.Services {
		imageName := extractImageName(serviceConfig.ExecStart)
		newImage := fmt.Sprintf("docker.io/%s/%s:%s", username, imageName, tagSuffix)
		serviceConfig.ExecStart = strings.Replace(serviceConfig.ExecStart, fmt.Sprintf("docker.io/anjali0/%s:latest", imageName), newImage, 1)
		configData.Services[serviceName] = serviceConfig
	}
}

// extractImageName extracts the image name from the exec_start command
func extractImageName(execStart string) string {
	parts := strings.Split(execStart, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "docker.io/") {
			// Split the part by '/' and ':' to get the image name
			imageParts := strings.Split(strings.Split(part, "/")[2], ":")
			return imageParts[0]
		}
	}
	return ""
}

// GetConfig returns the loaded configuration
func GetConfig() Config {
	return configData
}
