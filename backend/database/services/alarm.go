package services

import (
	postgres "backend/database"
	"backend/database/models"
	"fmt"
)

type AlarmService struct{}

var Alarm = AlarmService{}

func (s AlarmService) Create(alarm *models.Alarm) error {
	res := postgres.DB().Create(alarm).Error
	if res != nil {
		return fmt.Errorf("failed to create alarm: %w", res)
	}
	return nil
}
