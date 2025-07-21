package models

import "time"

type User struct {
    ID        int       `json:"id" db:"id"`
    Username  string    `json:"username" db:"username"`
    Password  string    `json:"-" db:"password"`          // Don't expose in JSON
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}