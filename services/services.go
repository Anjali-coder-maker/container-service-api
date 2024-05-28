package services

import (
	"encoding/json"
	"fmt"
	"go-podman-api/config"
	"go-podman-api/utils"
	"log"
	"net/http"
)

func InitializeServices() {
	loginResult := performLogin()
	if loginResult.Error != "" {
		log.Println("Login failed:", loginResult.Error)
		return
	}

	config, err := config.LoadConfiguration()
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

		// Perform each step for the service
		pullImage(imageName)
		tagImage(imageName)
		saveImage(imageName)
		loadImage(imageName)
		createUnitFile(imageName)
		enableAndStartService(imageName)
	}
}
func performLogin() utils.CommandResponse {
	username := "anjali0"
	password := "Anjali@123"
	registry := "docker.io"
	return utils.ExecuteCommand("podman", "login", registry, "-u", username, "-p", password)
}

func pullImage(imageName string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/pull/%s", imageName), nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Failed to pull image:", err)
		return
	}
	defer resp.Body.Close()
	var result utils.CommandResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		log.Printf("Failed to pull image %s: %s\n", imageName, result.Error)
	} else {
		log.Printf("Successfully pulled image %s\n", imageName)
	}
}

func tagImage(imageName string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/tag/%s", imageName), nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Failed to tag image:", err)
		return
	}
	defer resp.Body.Close()
	var result utils.CommandResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		log.Printf("Failed to tag image %s: %s\n", imageName, result.Error)
	} else {
		log.Printf("Successfully tagged image %s\n", imageName)
	}
}

func saveImage(imageName string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/save/%s", imageName), nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Failed to save image:", err)
		return
	}
	defer resp.Body.Close()
	var result utils.CommandResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		log.Printf("Failed to save image %s: %s\n", imageName, result.Error)
	} else {
		log.Printf("Successfully saved image %s\n", imageName)
	}
}

func loadImage(imageName string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/load/%s", imageName), nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Failed to load image:", err)
		return
	}
	defer resp.Body.Close()
	var result utils.CommandResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		log.Printf("Failed to load image %s: %s\n", imageName, result.Error)
	} else {
		log.Printf("Successfully loaded image %s\n", imageName)
	}
}

func createUnitFile(imageName string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/create-unit-file/%s", imageName), nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Failed to create unit file:", err)
		return
	}
	defer resp.Body.Close()
	var result utils.CommandResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		log.Printf("Failed to create unit file for %s: %s\n", imageName, result.Error)
	} else {
		log.Printf("Successfully created unit file for %s\n", imageName)
	}
}

func enableAndStartService(imageName string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/enable-and-start-service/%s", imageName), nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Failed to enable and start service:", err)
		return
	}
	defer resp.Body.Close()
	var result utils.CommandResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Error != "" {
		log.Printf("Failed to enable and start service %s: %s\n", imageName, result.Error)
	} else {
		log.Printf("Successfully enabled and started service %s\n", imageName)
	}
}

//...
