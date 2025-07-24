package models

import (
	"gorm.io/gorm"
)

type Sensor struct {
    gorm.Model
    SensorID    string    `json:"sensor_id" db:"sensor_id"`
    Name        string    `json:"name" db:"name"`
    Type        string    `json:"type" db:"type"`
    Description string    `json:"description" db:"description"`
    Location    string    `json:"location" db:"location"`
}