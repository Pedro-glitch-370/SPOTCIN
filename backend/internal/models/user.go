package models

import "time"

type User struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"` // pra não serializar
	CreatedAt time.Time `json:"created_at"`
}