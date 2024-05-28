package main

import (
	"go-podman-api/handlers"
	"go-podman-api/services"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// If you want to run code with one file paste the code from codesnippet.txt and run it

func main() {
	// Initialize services on startup
	go func() {
		services.InitializeServices()
	}()

	router := mux.NewRouter()
	router.HandleFunc("/login", handlers.Login).Methods("POST")
	router.HandleFunc("/pull/{image}", handlers.PullImage).Methods("GET")
	router.HandleFunc("/tag/{image}", handlers.TagImage).Methods("GET")
	router.HandleFunc("/save/{image}", handlers.SaveImage).Methods("GET")
	router.HandleFunc("/load/{image}", handlers.LoadImage).Methods("GET")
	router.HandleFunc("/create-unit-file/{image}", handlers.CreateUnitFile).Methods("GET")
	router.HandleFunc("/enable-and-start-service/{image}", handlers.EnableAndStartService).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
}
