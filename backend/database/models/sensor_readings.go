package models

import "time"

type SensorReading struct {
    ID        int       `json:"id" db:"id"`
    SensorID  string    `json:"sensor_id" db:"sensor_id"`
    Value     float64   `json:"value" db:"value"`
    Message   string    `json:"message" db:"message"`
    Timestamp time.Time `json:"timestamp" db:"timestamp"`
}