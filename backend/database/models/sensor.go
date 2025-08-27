package models

import (
	"gorm.io/gorm"
)

type Sensor struct {
	gorm.Model
	DeviceID    uint   `json:"device_id" db:"device_id"`
	Device      Device `gorm:"foreignKey:DeviceID" json:"device"`
	Name        string `json:"name" db:"name"`
	Type        string `json:"type" db:"type"`
	Description string `json:"description" db:"description"`
	Location    string `json:"location" db:"location"`
}
