package models

import (
	"gorm.io/gorm"
)

type User struct {
    gorm.Model
    Username  string    `json:"username" db:"username"`
    Password  string    `json:"-" db:"password"`          // Don't expose in JSON
}