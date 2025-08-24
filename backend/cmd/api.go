package main

import (
	handlers "backend/api"
	"backend/pkg/api"
	"backend/pkg/utils"
	"encoding/json"
	"log"
	"net/http"
)

// StartAPIServer starts the HTTP API server
func StartAPIServer() {
	port := utils.GetEnv("API_PORT", "8081")

	// Register routes
	router := api.NewRouter()
	router.RegisterRoute("GET", "/api/sensor-readings", handlers.Sensor.GetSensorReadings)

	// Health check endpoint
	router.RegisterRoute("GET", "/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	router.RegisterRoute("GET", "/api/devices", handlers.Device.GetDevices)

	log.Printf("API server starting on port %s", port)
	log.Println("Test API:")
	log.Println("  GET /api/health")

	// Start server in goroutine
	go func() {
		if err := http.ListenAndServe(":"+port, router); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()
}
