package models

import (
	"time"

	"gorm.io/gorm"
)

type SensorReading struct {
	gorm.Model
	SensorID         uint      `json:"sensor_id" db:"sensor_id"`
	Sensor           Sensor    `gorm:"foreignKey:SensorID" json:"sensor"`
	DeviceID         uint      `json:"device_id" db:"device_id"`
	Device           Device    `gorm:"foreignKey:DeviceID" json:"device"`
	Value            float64   `json:"value" db:"value"`
	Message          string    `json:"message" db:"message"`
	Timestamp        time.Time `json:"timestamp" db:"timestamp"`
	MessageTimestamp time.Time `json:"message_timestamp" db:"message_timestamp"`
}
