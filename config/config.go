package config

import (
	"encoding/json"
	"os"
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

// LoadConfiguration loads the default and user configurations
func LoadConfiguration() (Config, error) {
	var config Config

	// Load default configuration
	defaultConfigFile, err := os.ReadFile("default-config.json")
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(defaultConfigFile, &config); err != nil {
		return config, err
	}

	// User configuration loading and merging can be implemented here if needed

	return config, nil
}
