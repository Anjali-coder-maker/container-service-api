package config

import (
	"embed"
	"encoding/json"
	"log"
	"sync"
)

//go:embed default-config.json registry-services.json
var configFiles embed.FS

// ServiceConfig represents the configuration for a service
type ServiceConfig struct {
	Enabled    bool   `json:"enabled"`
	UnitFile   string `json:"unit_file"`
	Privileged bool   `json:"privileged"`
}

// Config represents the configuration file structure
type Config struct {
	Services map[string]ServiceConfig `json:"services"`
}

// ServiceTemplate represents the template configuration for a registry service
type ServiceTemplate struct {
	Enabled    bool   `json:"enabled"`
	UnitFile   string `json:"unit_file"`
	Privileged bool   `json:"privileged"`
}

// RegistryTemplates represents the registry services structure
type RegistryTemplates struct {
	Services map[string]ServiceTemplate `json:"services"`
}

var (
	// configData stores the configuration data
	configData   Config
	registryData RegistryTemplates

	// once ensures the configuration is loaded only once
	once sync.Once
)

// init is called automatically to load the configuration files
func init() {
	loadConfiguration()
}

// loadConfiguration reads the embedded configuration files and stores the values in configData and registryData
func loadConfiguration() {
	once.Do(func() {
		// Load default configuration
		defaultConfigFile, err := configFiles.ReadFile("default-config.json")
		if err != nil {
			log.Fatalf("Error reading default configuration file: %v", err)
		}
		if err := json.Unmarshal(defaultConfigFile, &configData); err != nil {
			log.Fatalf("Error unmarshalling default configuration file: %v", err)
		}

		// Load registry services
		registryConfigFile, err := configFiles.ReadFile("registry-services.json")
		if err != nil {
			log.Fatalf("Error reading registry services configuration file: %v", err)
		}
		if err := json.Unmarshal(registryConfigFile, &registryData); err != nil {
			log.Fatalf("Error unmarshalling registry services configuration file: %v", err)
		}
	})
}

// GetConfig returns the loaded configuration
func GetConfig() Config {
	return configData
}

// GetRegistryTemplates returns the loaded registry services templates
func GetRegistryTemplates() RegistryTemplates {
	return registryData
}
