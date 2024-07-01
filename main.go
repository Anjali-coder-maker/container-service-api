// package main

// import (
// 	"embed"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/exec"

// 	"github.com/gorilla/mux"
// )

// // CommandResponse is the structure of the response for command execution
// type CommandResponse struct {
// 	Output string `json:"output"`
// 	Error  string `json:"error,omitempty"`
// }

// // ServiceConfig represents the configuration for a service
// type ServiceConfig struct {
// 	Enabled      bool   `json:"enabled"`
// 	ExecStart    string `json:"exec_start"`
// 	ExecStop     string `json:"exec_stop"`
// 	ExecStopPost string `json:"exec_stop_post"`
// }

// // Config represents the configuration file structure
// type Config struct {
// 	Services map[string]ServiceConfig `json:"services"`
// }

// // ExecuteCommand runs a command and returns the output
// func ExecuteCommand(command string, args ...string) CommandResponse {
// 	cmd := exec.Command(command, args...)
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return CommandResponse{Output: string(output), Error: err.Error()}
// 	}
// 	return CommandResponse{Output: string(output)}
// }

// //go:embed default-config.json
// var defaultConfig embed.FS

// // LoadConfiguration loads the default and user configurations
// func LoadConfiguration() (Config, error) {

// 	var config Config

// 	// Read the embedded default configuration
// 	defaultConfigFile, err := defaultConfig.ReadFile("default-config.json")
// 	if err != nil {
// 		return config, err
// 	}

// 	// Unmarshal the JSON data into the config structure
// 	if err := json.Unmarshal(defaultConfigFile, &config); err != nil {
// 		return config, err
// 	}
// 	log.Printf("read your file")
// 	return config, nil
// }

// // performLogin contains the core logic for logging into a registry
// func performLogin() CommandResponse {
// 	username := "anjali0"
// 	password := "Anjali@123"
// 	registry := "docker.io"
// 	return ExecuteCommand("podman", "login", registry, "-u", username, "-p", password)
// }

// // Login handles logging into a registry
// func Login(w http.ResponseWriter, r *http.Request) {
// 	result := performLogin()
// 	json.NewEncoder(w).Encode(result)
// }

// //........

// // PullImage handles pulling an image from a registry based on the configuration file
// func PullImage(w http.ResponseWriter, r *http.Request) {
// 	// Ensure login before pulling the image
// 	loginResult := performLogin()
// 	if loginResult.Error != "" {
// 		json.NewEncoder(w).Encode(loginResult)
// 		return
// 	}

// 	config, err := LoadConfiguration()
// 	if err != nil {
// 		json.NewEncoder(w).Encode(CommandResponse{Error: err.Error()})
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	imageName := vars["image"]

// 	serviceConfig, exists := config.Services[imageName]
// 	if !exists {
// 		json.NewEncoder(w).Encode(CommandResponse{Error: fmt.Sprintf("Service %s not found in configuration", imageName)})
// 		return
// 	}

// 	if !serviceConfig.Enabled {
// 		json.NewEncoder(w).Encode(CommandResponse{Output: fmt.Sprintf("Service %s is disabled in configuration", imageName)})
// 		return
// 	}

// 	image := fmt.Sprintf("docker.io/anjali0/%s:latest", imageName)
// 	result := ExecuteCommand("podman", "pull", image)
// 	json.NewEncoder(w).Encode(result)
// }

// // end of the code .......

// // TagImage handles tagging a pulled image with a new name
// func TagImage(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)
// 	imageName := vars["image"]

// 	originalImage := fmt.Sprintf("docker.io/anjali0/%s:latest", imageName)
// 	newImage := fmt.Sprintf("docker.io/anjali0/%s-bcknd:latest", imageName)

// 	result := ExecuteCommand("podman", "tag", originalImage, newImage)
// 	json.NewEncoder(w).Encode(result)
// }

// // SaveImage handles saving the tagged image into a .tar file in the /tmp directory
// func SaveImage(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)
// 	imageName := vars["image"]

// 	tarFile := fmt.Sprintf("/tmp/%s-bcknd.tar", imageName)
// 	image := fmt.Sprintf("docker.io/anjali0/%s-bcknd:latest", imageName)

// 	result := ExecuteCommand("podman", "save", "-o", tarFile, image)
// 	log.Println("SaveImage endpoint hit")
// 	json.NewEncoder(w).Encode(result)
// }

// // LoadImage handles loading the .tar file as root
// func LoadImage(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)
// 	imageName := vars["image"]

// 	tarFile := fmt.Sprintf("/tmp/%s-bcknd.tar", imageName)

// 	result := ExecuteCommand("sudo", "podman", "load", "-i", tarFile)
// 	log.Println("LoadImage endpoint hit")
// 	json.NewEncoder(w).Encode(result)
// }

// // CreateUnitFile handles creating the systemd unit file for the container
// func CreateUnitFile(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	imageName := vars["image"]

// 	config, err := LoadConfiguration()
// 	if err != nil {
// 		json.NewEncoder(w).Encode(CommandResponse{Error: err.Error()})
// 		return
// 	}

// 	serviceConfig, exists := config.Services[imageName]
// 	if !exists {
// 		json.NewEncoder(w).Encode(CommandResponse{Error: fmt.Sprintf("Service %s not found in configuration", imageName)})
// 		return
// 	}

// 	unitFileContent := fmt.Sprintf(`[Unit]
//  Description=Podman container-%s-bcknd.service
//  Documentation=man:podman-generate-systemd(1)
//  Wants=network-online.target
//  After=network-online.target
//  RequiresMountsFor=%%t/containers

//  [Service]
//  Environment=PODMAN_SYSTEMD_UNIT=%%n
//  Restart=on-failure
//  ExecStart=%s
//  ExecStop=%s
//  ExecStopPost=%s
//  TimeoutStopSec=70
//  Type=simple
//  NotifyAccess=all

//  [Install]
//  WantedBy=multi-user.target
//  `, imageName, serviceConfig.ExecStart, serviceConfig.ExecStop, serviceConfig.ExecStopPost)

// 	unitFilePath := fmt.Sprintf("/etc/systemd/system/%s-bcknd.service", imageName)

// 	err = os.WriteFile(unitFilePath, []byte(unitFileContent), 0644)
// 	if err != nil {
// 		json.NewEncoder(w).Encode(CommandResponse{Error: err.Error()})
// 		return
// 	}
// 	json.NewEncoder(w).Encode(CommandResponse{Output: "Unit file created successfully"})
// }

// // EnableAndStartService handles enabling and starting the systemd service
// func EnableAndStartService(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	imageName := vars["image"]
// 	serviceFileName := fmt.Sprintf("%s-bcknd.service", imageName)

// 	enableResult := ExecuteCommand("sudo", "systemctl", "enable", serviceFileName)
// 	if enableResult.Error != "" {
// 		json.NewEncoder(w).Encode(enableResult)
// 		return
// 	}

// 	daemonReloadResult := ExecuteCommand("sudo", "systemctl", "daemon-reload")
// 	if daemonReloadResult.Error != "" {
// 		json.NewEncoder(w).Encode(daemonReloadResult)
// 		return
// 	}

// 	startResult := ExecuteCommand("sudo", "systemctl", "start", serviceFileName)
// 	json.NewEncoder(w).Encode(startResult)
// }

// // InitializeServices initializes all services defined in the configuration file
// func InitializeServices() {
// 	loginResult := performLogin()
// 	if loginResult.Error != "" {
// 		log.Println("Login failed:", loginResult.Error)
// 		return
// 	}

// 	config, err := LoadConfiguration()
// 	if err != nil {
// 		log.Println("Failed to load configuration:", err)
// 		return
// 	}

// 	for imageName, serviceConfig := range config.Services {
// 		if !serviceConfig.Enabled {
// 			log.Printf("Service %s is disabled in configuration\n", imageName)
// 			continue
// 		}

// 		log.Printf("Processing service: %s\n", imageName)

// 		// Perform each step for the service
// 		pullImage(imageName)
// 		tagImage(imageName)
// 		saveImage(imageName)
// 		loadImage(imageName)
// 		createUnitFile(imageName)
// 		enableAndStartService(imageName)
// 	}
// }

// func pullImage(imageName string) {
// 	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/pull/%s", imageName), nil)
// 	if err != nil {
// 		log.Println("Failed to create request:", err)
// 		return
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Println("Failed to pull image:", err)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	var result CommandResponse
// 	json.NewDecoder(resp.Body).Decode(&result)
// 	if result.Error != "" {
// 		log.Printf("Failed to pull image %s: %s\n", imageName, result.Error)
// 	} else {
// 		log.Printf("Successfully pulled image %s\n", imageName)
// 	}
// }

// func tagImage(imageName string) {
// 	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/tag/%s", imageName), nil)
// 	if err != nil {
// 		log.Println("Failed to create request:", err)
// 		return
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Println("Failed to tag image:", err)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	var result CommandResponse
// 	json.NewDecoder(resp.Body).Decode(&result)
// 	if result.Error != "" {
// 		log.Printf("Failed to tag image %s: %s\n", imageName, result.Error)
// 	} else {
// 		log.Printf("Successfully tagged image %s\n", imageName)
// 	}
// }

// func saveImage(imageName string) {
// 	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/save/%s", imageName), nil)
// 	if err != nil {
// 		log.Println("Failed to create request:", err)
// 		return
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Println("Failed to save image:", err)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	var result CommandResponse
// 	json.NewDecoder(resp.Body).Decode(&result)
// 	if result.Error != "" {
// 		log.Printf("Failed to save image %s: %s\n", imageName, result.Error)
// 	} else {
// 		log.Printf("Successfully saved image %s\n", imageName)
// 	}
// }

// func loadImage(imageName string) {
// 	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/load/%s", imageName), nil)
// 	if err != nil {
// 		log.Println("Failed to create request:", err)
// 		return
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Println("Failed to load image:", err)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	var result CommandResponse
// 	json.NewDecoder(resp.Body).Decode(&result)
// 	if result.Error != "" {
// 		log.Printf("Failed to load image %s: %s\n", imageName, result.Error)
// 	} else {
// 		log.Printf("Successfully loaded image %s\n", imageName)
// 	}
// }

// func createUnitFile(imageName string) {
// 	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/create-unit-file/%s", imageName), nil)
// 	if err != nil {
// 		log.Println("Failed to create request:", err)
// 		return
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Println("Failed to create unit file:", err)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	var result CommandResponse
// 	json.NewDecoder(resp.Body).Decode(&result)
// 	if result.Error != "" {
// 		log.Printf("Failed to create unit file for %s: %s\n", imageName, result.Error)
// 	} else {
// 		log.Printf("Successfully created unit file for %s\n", imageName)
// 	}
// }

// func enableAndStartService(imageName string) {
// 	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/enable-and-start-service/%s", imageName), nil)
// 	if err != nil {
// 		log.Println("Failed to create request:", err)
// 		return
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Println("Failed to enable and start service:", err)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	var result CommandResponse
// 	json.NewDecoder(resp.Body).Decode(&result)
// 	if result.Error != "" {
// 		log.Printf("Failed to enable and start service %s: %s\n", imageName, result.Error)
// 	} else {
// 		log.Printf("Successfully enabled and started service %s\n", imageName)
// 	}
// }

// func main() {
// 	// Initialize services on startup
// 	log.Println("Starting configuration loader...")

// 	go func() {
// 		InitializeServices()
// 	}()

// 	log.Println("Configuration applied successfully.")

// 	router := mux.NewRouter()
// 	router.HandleFunc("/login", Login).Methods("POST")
// 	router.HandleFunc("/pull/{image}", PullImage).Methods("GET")
// 	router.HandleFunc("/tag/{image}", TagImage).Methods("GET")
// 	router.HandleFunc("/save/{image}", SaveImage).Methods("GET")
// 	router.HandleFunc("/load/{image}", LoadImage).Methods("GET")
// 	router.HandleFunc("/create-unit-file/{image}", CreateUnitFile).Methods("GET")
// 	router.HandleFunc("/enable-and-start-service/{image}", EnableAndStartService).Methods("GET")

// 	log.Fatal(http.ListenAndServe(":8000", router))
// }

// //...

package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

// CommandResponse is the structure of the response for command execution
type CommandResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

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

// ExecuteCommand runs a command and returns the output
func ExecuteCommand(command string, args ...string) CommandResponse {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CommandResponse{Output: string(output), Error: err.Error()}
	}
	return CommandResponse{Output: string(output)}
}

//go:embed default-config.json
var defaultConfig embed.FS

// LoadConfiguration loads the default and user configurations
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
	log.Printf("Configuration file loaded successfully")
	return config, nil
}

// performLogin contains the core logic for logging into a registry
func performLogin() CommandResponse {
	username := "anjali0"
	password := "Anjali@123"
	registry := "docker.io"
	return ExecuteCommand("podman", "login", registry, "-u", username, "-p", password)
}

func InitializeServices() {
	loginResult := performLogin()
	if loginResult.Error != "" {
		log.Println("Login failed:", loginResult.Error)
		return
	}

	config, err := LoadConfiguration()
	if err != nil {
		log.Println("Failed to load configuration:", err)
		return
	}

	for imageName, serviceConfig := range config.Services {
		if !serviceConfig.Enabled {
			log.Printf("Service %s is disabled in configuration\n", imageName)
			continue
		}

		log.Printf("Processing service: %s\n", imageName)
		// Check and disable existing service if running
		if checkAndDisableExistingService(imageName) {
			log.Printf("Existing service %s found and disabled\n", imageName)
		}

		pullImage(imageName)
		tagImage(imageName)
		// saveImage(imageName)
		// loadImage(imageName)
		createUnitFile(imageName)
		enableAndStartService(imageName)
	}
}

func checkAndDisableExistingService(imageName string) bool {
	serviceFileName := fmt.Sprintf("%s.service", imageName)

	// Check if the service is active
	checkResult := ExecuteCommand("sudo", "systemctl", "is-active", "--quiet", serviceFileName)
	if checkResult.Error == "" {
		// Service is active, disable it
		stopResult := ExecuteCommand("sudo", "systemctl", "stop", serviceFileName)
		if stopResult.Error != "" {
			log.Printf("Failed to stop service %s: %s\n", serviceFileName, stopResult.Error)
			return false
		}

		disableResult := ExecuteCommand("sudo", "systemctl", "disable", serviceFileName)
		if disableResult.Error != "" {
			log.Printf("Failed to disable service %s: %s\n", serviceFileName, disableResult.Error)
			return false
		}

		maskResult := ExecuteCommand("sudo", "systemctl", "mask", serviceFileName)
		if maskResult.Error != "" {
			log.Printf("Failed to mask service %s: %s\n", serviceFileName, disableResult.Error)
			return false
		}
		daemonReloadResult := ExecuteCommand("sudo", "systemctl", "daemon-reload")
		if daemonReloadResult.Error != "" {
			log.Printf("Failed to reload daemon after disabling service %s: %s\n", serviceFileName, daemonReloadResult.Error)
			return false
		}

		return true
	}
	// Service is not active or does not exist
	return false
}

func pullImage(imageName string) {
	image := fmt.Sprintf("docker.io/anjali0/%s:latest", imageName)
	result := ExecuteCommand("podman", "pull", image)
	logResult("pull", imageName, result)
}

func tagImage(imageName string) {
	originalImage := fmt.Sprintf("docker.io/anjali0/%s:latest", imageName)
	newImage := fmt.Sprintf("docker.io/anjali0/%s-bcknd:latest", imageName)
	result := ExecuteCommand("podman", "tag", originalImage, newImage)
	logResult("tag", imageName, result)
}

/*For now I am commenting this as this code with run with sudo privileges
and is responsible for pulling and starting the container services so we not
need to addtional code to save the container images into tar file and load them
as a root image separately*/

// func saveImage(imageName string) {
// 	tarFile := fmt.Sprintf("/tmp/%s-bcknd.tar", imageName)
// 	image := fmt.Sprintf("docker.io/anjali0/%s-bcknd:latest", imageName)
// 	result := ExecuteCommand("podman", "save", "-o", tarFile, image)
// 	logResult("save", imageName, result)
// }

// func loadImage(imageName string) {
// 	tarFile := fmt.Sprintf("/tmp/%s-bcknd.tar", imageName)
// 	result := ExecuteCommand("sudo", "podman", "load", "-i", tarFile)
// 	logResult("load", imageName, result)
// }

func createUnitFile(imageName string) {
	config, err := LoadConfiguration()
	if err != nil {
		log.Println("Failed to load configuration:", err)
		return
	}

	serviceConfig, exists := config.Services[imageName]
	if !exists {
		log.Printf("Service %s not found in configuration\n", imageName)
		return
	}

	unitFileContent := fmt.Sprintf(`[Unit]
Description=Podman container-%s-bcknd.service
Documentation=man:podman-generate-systemd(1)
Wants=network-online.target
After=network-online.target
RequiresMountsFor=%%t/containers


[Service]
Environment=PODMAN_SYSTEMD_UNIT=%%n
Restart=on-failure
ExecStart=%s
ExecStop=%s
ExecStopPost=%s
TimeoutStopSec=70
Type=simple
NotifyAccess=all


[Install]
WantedBy=multi-user.target
`, imageName, serviceConfig.ExecStart, serviceConfig.ExecStop, serviceConfig.ExecStopPost)

	unitFilePath := fmt.Sprintf("/etc/systemd/system/%s-bcknd.service", imageName)

	err = os.WriteFile(unitFilePath, []byte(unitFileContent), 0644)
	if err != nil {
		log.Printf("Failed to create unit file for %s: %s\n", imageName, err)
		return
	}
	log.Printf("Successfully created unit file for %s\n", imageName)
}

func enableAndStartService(imageName string) {
	serviceFileName := fmt.Sprintf("%s-bcknd.service", imageName)

	enableResult := ExecuteCommand("sudo", "systemctl", "enable", serviceFileName)
	if enableResult.Error != "" {
		logResult("enable", imageName, enableResult)
		return
	}

	daemonReloadResult := ExecuteCommand("sudo", "systemctl", "daemon-reload")
	if daemonReloadResult.Error != "" {
		logResult("daemon-reload", imageName, daemonReloadResult)
		return
	}

	startResult := ExecuteCommand("sudo", "systemctl", "start", serviceFileName)
	logResult("start", imageName, startResult)
}

func logResult(action, imageName string, result CommandResponse) {
	if result.Error != "" {
		log.Printf("Failed to %s image %s: %s\n", action, imageName, result.Error)
	} else {
		log.Printf("Successfully %sed image %s: %s\n", action, imageName, result.Output)
	}
}

func main() {
	log.Println("Starting configuration loader...")

	InitializeServices()

	log.Println("Configuration applied successfully.")
}
