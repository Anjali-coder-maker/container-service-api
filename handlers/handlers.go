package handlers

import (
	"encoding/json"
	"fmt"
	"go-podman-api/config"
	"go-podman-api/utils"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// performLogin contains the core logic for logging into a registry
func performLogin() utils.CommandResponse {
	username := "anjali0"
	password := "Anjali@123"
	registry := "docker.io"
	return utils.ExecuteCommand("podman", "login", registry, "-u", username, "-p", password)
}

// Login handles logging into a registry
func Login(w http.ResponseWriter, r *http.Request) {
	result := performLogin()
	json.NewEncoder(w).Encode(result)
}

//........

// PullImage handles pulling an image from a registry based on the configuration file
func PullImage(w http.ResponseWriter, r *http.Request) {
	// Ensure login before pulling the image
	loginResult := performLogin()
	if loginResult.Error != "" {
		json.NewEncoder(w).Encode(loginResult)
		return
	}

	config, err := config.LoadConfiguration()
	if err != nil {
		json.NewEncoder(w).Encode(utils.CommandResponse{Error: err.Error()})
		return
	}

	vars := mux.Vars(r)
	imageName := vars["image"]

	serviceConfig, exists := config.Services[imageName]
	if !exists {
		json.NewEncoder(w).Encode(utils.CommandResponse{Error: fmt.Sprintf("Service %s not found in configuration", imageName)})
		return
	}

	if !serviceConfig.Enabled {
		json.NewEncoder(w).Encode(utils.CommandResponse{Output: fmt.Sprintf("Service %s is disabled in configuration", imageName)})
		return
	}

	image := fmt.Sprintf("docker.io/anjali0/%s:latest", imageName)
	result := utils.ExecuteCommand("podman", "pull", image)
	json.NewEncoder(w).Encode(result)
}

// end of the code .......

// TagImage handles tagging a pulled image with a new name
func TagImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	imageName := vars["image"]

	originalImage := fmt.Sprintf("docker.io/anjali0/%s:latest", imageName)
	newImage := fmt.Sprintf("docker.io/anjali0/%s-bcknd:latest", imageName)

	result := utils.ExecuteCommand("podman", "tag", originalImage, newImage)
	json.NewEncoder(w).Encode(result)
}

// SaveImage handles saving the tagged image into a .tar file in the /tmp directory
func SaveImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	imageName := vars["image"]

	tarFile := fmt.Sprintf("/tmp/%s-bcknd.tar", imageName)
	image := fmt.Sprintf("docker.io/anjali0/%s-bcknd:latest", imageName)

	result := utils.ExecuteCommand("podman", "save", "-o", tarFile, image)
	log.Println("SaveImage endpoint hit")
	json.NewEncoder(w).Encode(result)
}

// LoadImage handles loading the .tar file as root
func LoadImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	imageName := vars["image"]

	tarFile := fmt.Sprintf("/tmp/%s-bcknd.tar", imageName)

	result := utils.ExecuteCommand("sudo", "podman", "load", "-i", tarFile)
	log.Println("LoadImage endpoint hit")
	json.NewEncoder(w).Encode(result)
}

// CreateUnitFile handles creating the systemd unit file for the container
func CreateUnitFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageName := vars["image"]

	config, err := config.LoadConfiguration()
	if err != nil {
		json.NewEncoder(w).Encode(utils.CommandResponse{Error: err.Error()})
		return
	}

	serviceConfig, exists := config.Services[imageName]
	if !exists {
		json.NewEncoder(w).Encode(utils.CommandResponse{Error: fmt.Sprintf("Service %s not found in configuration", imageName)})
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
		json.NewEncoder(w).Encode(utils.CommandResponse{Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(utils.CommandResponse{Output: "Unit file created successfully"})
}

// EnableAndStartService handles enabling and starting the systemd service
func EnableAndStartService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageName := vars["image"]
	serviceFileName := fmt.Sprintf("%s-bcknd.service", imageName)

	enableResult := utils.ExecuteCommand("sudo", "systemctl", "enable", serviceFileName)
	if enableResult.Error != "" {
		json.NewEncoder(w).Encode(enableResult)
		return
	}

	daemonReloadResult := utils.ExecuteCommand("sudo", "systemctl", "daemon-reload")
	if daemonReloadResult.Error != "" {
		json.NewEncoder(w).Encode(daemonReloadResult)
		return
	}

	startResult := utils.ExecuteCommand("sudo", "systemctl", "start", serviceFileName)
	json.NewEncoder(w).Encode(startResult)
}
