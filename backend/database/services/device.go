package services

import (
	postgres "backend/database"
	"backend/database/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type DeviceService struct{}

var Device = DeviceService{}

func (s DeviceService) Create(device *models.Device) error {
	res := postgres.DB().Create(device).Error
	if res != nil {
		return fmt.Errorf("failed to create device: %w", res)
	}
	return nil
}

func (s DeviceService) GetByID(deviceID string) (*models.Device, error) {
	var device models.Device
	err := postgres.DB().First(&device, "name = ?", deviceID).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, errors.New("device not found")
		default:
			return nil, fmt.Errorf("failed to get device: %w", err)
		}
	}
	return &device, nil
}
