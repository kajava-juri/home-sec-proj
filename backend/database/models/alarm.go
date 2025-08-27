package models

import (
	"gorm.io/gorm"
)

type Alarm struct {
    gorm.Model
    SensorID         uint      `json:"sensor_id" db:"sensor_id"`
}