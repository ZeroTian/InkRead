package models

import "time"

type BookSource struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	URLTemplate    string    `json:"url_template" db:"url_template"`
	Encoding       string    `json:"encoding" db:"encoding"`
	BookNameRule   string    `json:"book_name_rule" db:"book_name_rule"`
	AuthorRule     string    `json:"author_rule" db:"author_rule"`
	ContentRule    string    `json:"content_rule" db:"content_rule"`
	ChapterRule    string    `json:"chapter_list_rule" db:"chapter_list_rule"`
	ChapterURLRule string    `json:"chapter_url_rule" db:"chapter_url_rule"`
	Enabled        bool      `json:"enabled" db:"enabled"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type CleanupRule struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Pattern     string    `json:"pattern" db:"pattern"`
	Replacement string    `json:"replacement" db:"replacement"`
	RuleType    string    `json:"rule_type" db:"rule_type"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	Priority    int       `json:"priority" db:"priority"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type WebImportRequest struct {
	URL     string `json:"url" binding:"required"`
	SourceID string `json:"source_id"`
}

type SourceTestRequest struct {
	URL      string `json:"url" binding:"required"`
	SourceID string `json:"source_id"`
}

type SourceTestResponse struct {
	Success bool           `json:"success"`
	Book    *Book          `json:"book,omitempty"`
	Error   string         `json:"error,omitempty"`
	Content *ChapterContent `json:"content,omitempty"`
}

type ChapterContent struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
