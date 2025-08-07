package services

import (
	postgres "backend/database"
	"backend/database/models"
	"backend/pkg/utils"
	"fmt"
)

type SensorReadingService struct{}

var SensorReading = SensorReadingService{}

func (s SensorReadingService) Create(reading *models.SensorReading) error {
	res := postgres.DB().Create(reading).Error
	utils.DebugLog("Creating sensor reading: %v", reading)
	if res != nil {
		fmt.Printf("Error creating sensor reading: %v\n", res)
		return res
	}
	return nil
}


func (s SensorReadingService) GetPaginated(page, pageSize int) ([]models.SensorReading, int64, error) {
	var readings []models.SensorReading
	var totalCount int64

	db := postgres.DB()

	// Get total count
	if err := db.Model(&models.SensorReading{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count sensor readings: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get paginated results ordered by timestamp (newest first)
	if err := db.Order("timestamp DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&readings).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch sensor readings: %w", err)
	}

	return readings, totalCount, nil
}
