package utils

import (
	"embed"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var (
	//go:embed .env
	envFile embed.FS

	// envMap stores the environment variables
	envMap map[string]string

	// once ensures the .env file is loaded only once
	once sync.Once
)

// init is called automatically to load the .env file and perform login
func init() {
	loadEnvFromEmbed()
	performLogin()
}

// loadEnvFromEmbed reads the embedded .env file and stores the values in envMap
func loadEnvFromEmbed() {
	once.Do(func() {
		data, err := envFile.ReadFile(".env")
		if err != nil {
			log.Fatalf("Error reading .env file from embed: %v", err)
		}

		envMap, err = godotenv.Unmarshal(string(data))
		if err != nil {
			log.Fatalf("Error unmarshalling .env file: %v", err)
		}

		for key, value := range envMap {
			os.Setenv(key, value)
		}
	})
}

// performLogin contains the core logic for logging into a registry
func performLogin() {
	// Check if the user is already logged in
	if isLoggedIn() {
		log.Println("Already logged in")
		return
	}

	var username, password string

	// Determine the architecture
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		username = envMap["DOCKER_USERNAME_AMD"]
		password = envMap["DOCKER_PASSWORD_AMD"]
	case "arm64":
		username = envMap["DOCKER_USERNAME_ARM"]
		password = envMap["DOCKER_PASSWORD_ARM"]
	default:
		log.Fatalf("Unsupported architecture: %s", arch)
	}

	registry := "docker.io"
	result := ExecuteCommand("podman", "login", registry, "-u", username, "-p", password)
	if result.Error != "" {
		log.Fatalf("Login failed: %v", result.Error)
	}
	log.Println("Login succeeded")
}

// isLoggedIn checks if the user is already logged in
func isLoggedIn() bool {
	cmd := exec.Command("podman", "login", "--get-login", "docker.io")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If there's an error, check if it's due to not being logged in
		if strings.Contains(string(output), "Error: not logged into docker.io") {
			return false
		}
		// Handle other errors (e.g., network issues)
		return false
	}

	// If the output contains a username, the user is logged in
	return strings.TrimSpace(string(output)) != ""
}

func Getenvmap() map[string]string {
	return envMap
}
