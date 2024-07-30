package config

import (
	"embed"
	"encoding/json"
	"log"
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
		// log.Println("Loading configuration from embedded default-config.json")
		defaultConfigFile, err := defaultConfig.ReadFile("default-config.json")
		if err != nil {
			log.Fatalf("Error reading default configuration file: %v", err)
		}

		if err := json.Unmarshal(defaultConfigFile, &configData); err != nil {
			log.Fatalf("Error unmarshalling default configuration file: %v", err)
		}

	})
}

// GetConfig returns the loaded configuration
func GetConfig() Config {
	return configData
}
