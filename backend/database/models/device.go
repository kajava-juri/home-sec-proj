package models

import (
	"gorm.io/gorm"
)

type Device struct {
    gorm.Model
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    Location    string    `json:"location" db:"location"`
}