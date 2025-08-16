package services

import (
	postgres "backend/database"
	"backend/database/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type SensorService struct{}

var Sensor = SensorService{}

func (s SensorService) Create(sensor *models.Sensor) error {
	res := postgres.DB().Create(sensor).Error
	if res != nil {
		return fmt.Errorf("failed to create sensor: %w", res)
	}
	return nil
}

func (s SensorService) GetByID(sensorID string) (*models.Sensor, error) {
	var sensor models.Sensor
	err := postgres.DB().First(&sensor, "id = ?", sensorID).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, errors.New("sensor not found")
		default:
			return nil, fmt.Errorf("failed to get sensor: %w", err)
		}
	}
	return &sensor, nil
}

func (s SensorService) GetOrCreate(sensor *models.Sensor) (*models.Sensor, error) {
	result := postgres.DB().FirstOrCreate(sensor)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get or create sensor: %w", result.Error)
	}
	return sensor, nil
}
