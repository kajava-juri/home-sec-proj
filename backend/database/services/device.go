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

func (s DeviceService) GetAll() ([]models.Device, error) {
	var devices []models.Device
	err := postgres.DB().Find(&devices).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	return devices, nil
}

func (s DeviceService) GetPaginated(page, pageSize int) ([]models.Device, int64, error) {
	var devices []models.Device
	var totalCount int64

	db := postgres.DB()

	// Get total count
	if err := db.Model(&models.Device{}).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count devices: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get paginated results ordered by updated_at (newest first)
	if err := db.Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&devices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch devices: %w", err)
	}

	return devices, totalCount, nil
}

func (s DeviceService) Update(device *models.Device) error {
	res := postgres.DB().Save(device).Error
	if res != nil {
		return fmt.Errorf("failed to update device: %w", res)
	}
	return nil
}
