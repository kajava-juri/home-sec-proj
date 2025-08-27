package models

import (
	"gorm.io/gorm"
)

type Device struct {
	gorm.Model
	Name        string  `json:"name" db:"name"`
	WifiStatus  string  `json:"wifi_status" db:"wifi_status"`
	MqttStatus  string  `json:"mqtt_status" db:"mqtt_status"`
	UptimeMs    float64 `json:"uptime_ms" db:"uptime_ms"`
	AlarmStatus string  `json:"alarm_status" db:"alarm_status"`
	Description string  `json:"description" db:"description"`
	Location    string  `json:"location" db:"location"`
}
