package models

import "time"

type Sensor struct {
    ID          int       `json:"id" db:"id"`
    SensorID    string    `json:"sensor_id" db:"sensor_id"`
    Name        string    `json:"name" db:"name"`
    Type        string    `json:"type" db:"type"`
    Description string    `json:"description" db:"description"`
    Location    string    `json:"location" db:"location"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}