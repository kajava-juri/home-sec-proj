package models

import (
	"gorm.io/gorm"
)

type Alarm struct {
	gorm.Model
	DeviceID uint   `json:"device_id" db:"device_id"`
	Device   Device `gorm:"foreignKey:DeviceID" json:"device"`
	Payload  string `json:"payload" db:"payload"`
	Event    string `json:"event" db:"event"`
}
