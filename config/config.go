package config

import (
	"embed"
	"encoding/json"
	"log"
)

//go:embed default-config.json
var defaultConfig embed.FS

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

func LoadConfiguration() (Config, error) {
	var config Config

	// Read the embedded default configuration
	defaultConfigFile, err := defaultConfig.ReadFile("default-config.json")
	if err != nil {
		return config, err
	}

	// Unmarshal the JSON data into the config structure
	if err := json.Unmarshal(defaultConfigFile, &config); err != nil {
		return config, err
	}
	log.Printf("read your file")
	return config, nil
}

//.....

// // LoadConfiguration loads the default and user configurations
// func LoadConfiguration() (Config, error) {
// 	var config Config

// 	// // // Copy the default configuration to the /etc/ folder

// 	// copyResult := utils.ExecuteCommand("sudo", "cp", "default-config.json", "/etc/default-config.json")
// 	// if copyResult.Error != "" {
// 	// 	return config, fmt.Errorf(copyResult.Error)
// 	// }

// 	defaultConfigFile, err := os.ReadFile("default-config.json")

// 	if err != nil {
// 		return config, err
// 	}
// 	if err := json.Unmarshal(defaultConfigFile, &config); err != nil {

// 		return config, err
// 	}

// 	log.Printf("read your file")

// 	// User configuration loading and merging can be implemented here if needed

// 	return config, nil
// }
