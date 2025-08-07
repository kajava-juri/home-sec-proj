package main

import (
	"backend/database/services"
	"backend/pkg/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int64       `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}

// getSensorReadings handles GET /api/sensor-readings with pagination
func getSensorReadings(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Only allow GET requests
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	// Set defaults
	page := 1
	pageSize := 100

	// Parse page parameter
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse page_size parameter
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// Limit maximum page size to prevent abuse
	if pageSize > 1000 {
		pageSize = 1000
	}

	// Get paginated data from service
	readings, totalCount, err := services.SensorReading.GetPaginated(page, pageSize)
	if err != nil {
		log.Printf("Error fetching sensor readings: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		return
	}

	// Calculate total pages
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	// Create response
	response := PaginatedResponse{
		Data:       readings,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}

	// Send response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// StartAPIServer starts the HTTP API server
func StartAPIServer() {
	port := utils.GetEnv("API_PORT", "8081")

	// Register routes
	http.HandleFunc("/api/sensor-readings", getSensorReadings)

	// Health check endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	log.Printf("API server starting on port %s", port)
	log.Println("Available endpoints:")
	log.Println("  GET /api/sensor-readings?page=1&page_size=100")
	log.Println("  GET /api/health")

	// Start server in goroutine
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()
}
