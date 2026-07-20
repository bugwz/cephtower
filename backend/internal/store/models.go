package store

import "time"

type Setting struct {
	Key       string `gorm:"primaryKey;size:128"`
	Value     string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
