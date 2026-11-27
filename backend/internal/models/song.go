package models

import "time"

// schema de dados que vai ser montado no DB
type Song struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Path      string    `gorm:"uniqueIndex;not null" json:"path"`
	Title     string    `json:"title"`
	ImageURL  string    `json:"image_url,omitempty"`
	Duration  int       `json:"duration,omitempty"` // segundos
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}