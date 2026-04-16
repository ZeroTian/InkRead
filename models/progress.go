package models

import (
	"time"
)

type ReadingProgress struct {
	ID              string    `json:"id"`
	BookID          string    `json:"book_id"`
	UserID          string    `json:"user_id"`
	CurrentChapter  int       `json:"current_chapter"`
	ScrollPosition  float64   `json:"scroll_position"`
	Percentage      float64   `json:"percentage"`
	UpdatedAt       time.Time `json:"updated_at"`
}
