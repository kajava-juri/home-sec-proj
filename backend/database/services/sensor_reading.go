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
