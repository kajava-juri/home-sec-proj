package handlers

import (
	"backend/database/services"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type SensorReadingHandler struct{}

var Sensor = SensorReadingHandler{}

// getSensorReadings handles GET /api/sensor-readings with pagination
func (h SensorReadingHandler) GetSensorReadings(w http.ResponseWriter, r *http.Request) {
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
