package models

import (
	"io"
	"time"
)

type Book struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	FileType  string    `json:"file_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BookListResponse struct {
	Books     []Book `json:"books"`
	Total     int    `json:"total"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type SummarizeRequest struct {
	BookID string `json:"book_id" binding:"required"`
}

type SummarizeResponse struct {
	Summary   string `json:"summary"`
	BookID    string `json:"book_id"`
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
}

type UploadedFile struct {
	Filename   string
	Data       io.Reader
	Size       int64
	Author     string
	UploadedAt time.Time
}

// ReadingSettings 阅读设置
type ReadingSettings struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	FontSize   int     `json:"font_size"`
	LineHeight float64 `json:"line_height"`
	Theme      string  `json:"theme"`
	Font       string  `json:"font"`
}
