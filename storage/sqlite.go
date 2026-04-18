package storage

import (
	"database/sql"
	"inkread/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS books (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT DEFAULT '',
		file_path TEXT NOT NULL,
		file_size INTEGER DEFAULT 0,
		file_type TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reading_progress (
		id TEXT PRIMARY KEY,
		book_id TEXT NOT NULL,
		user_id TEXT DEFAULT 'default',
		current_chapter INTEGER DEFAULT 0,
		scroll_position REAL DEFAULT 0,
		percentage REAL DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(book_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS book_sources (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		url_template TEXT NOT NULL,
		encoding TEXT DEFAULT 'utf-8',
		book_name_rule TEXT DEFAULT '',
		author_rule TEXT DEFAULT '',
		content_rule TEXT DEFAULT '',
		chapter_list_rule TEXT DEFAULT '',
		chapter_url_rule TEXT DEFAULT '',
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS cleanup_rules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		pattern TEXT NOT NULL,
		replacement TEXT DEFAULT '',
		rule_type TEXT DEFAULT 'replace',
		enabled INTEGER DEFAULT 1,
		priority INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reading_settings (
		id TEXT PRIMARY KEY,
		user_id TEXT UNIQUE NOT NULL,
		font_size INTEGER DEFAULT 18,
		line_height REAL DEFAULT 1.8,
		theme TEXT DEFAULT 'light',
		font TEXT DEFAULT 'Georgia'
	);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *SQLiteStore) CreateBook(book *models.Book) error {
	query := `
	INSERT INTO books (id, title, author, file_path, file_size, file_type, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, book.ID, book.Title, book.Author, book.FilePath,
		book.FileSize, book.FileType, book.CreatedAt, book.UpdatedAt)
	return err
}

func (s *SQLiteStore) GetBook(id string) (*models.Book, error) {
	query := `
	SELECT id, title, author, file_path, file_size, file_type, created_at, updated_at
	FROM books WHERE id = ?
	`
	row := s.db.QueryRow(query, id)

	var book models.Book
	err := row.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath,
		&book.FileSize, &book.FileType, &book.CreatedAt, &book.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (s *SQLiteStore) ListBooks(page, pageSize int) ([]models.Book, int, error) {
	offset := (page - 1) * pageSize

	var total int
	countQuery := `SELECT COUNT(*) FROM books`
	if err := s.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
	SELECT id, title, author, file_path, file_size, file_type, created_at, updated_at
	FROM books ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	rows, err := s.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.FilePath,
			&book.FileSize, &book.FileType, &book.CreatedAt, &book.UpdatedAt); err != nil {
			return nil, 0, err
		}
		books = append(books, book)
	}

	return books, total, nil
}

func (s *SQLiteStore) DeleteBook(id string) error {
	query := `DELETE FROM books WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *SQLiteStore) UpdateBook(book *models.Book) error {
	book.UpdatedAt = time.Now()
	query := `
	UPDATE books SET title = ?, author = ?, updated_at = ?
	WHERE id = ?
	`
	_, err := s.db.Exec(query, book.Title, book.Author, book.UpdatedAt, book.ID)
	return err
}

func (s *SQLiteStore) GetProgress(bookID string) (*models.ReadingProgress, error) {
	query := `
	SELECT id, book_id, user_id, current_chapter, scroll_position, percentage, updated_at
	FROM reading_progress WHERE book_id = ? AND user_id = 'default'
	`
	row := s.db.QueryRow(query, bookID)

	var progress models.ReadingProgress
	err := row.Scan(&progress.ID, &progress.BookID, &progress.UserID,
		&progress.CurrentChapter, &progress.ScrollPosition, &progress.Percentage, &progress.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (s *SQLiteStore) SaveProgress(progress *models.ReadingProgress) error {
	progress.UpdatedAt = time.Now()
	query := `
	INSERT INTO reading_progress (id, book_id, user_id, current_chapter, scroll_position, percentage, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(book_id, user_id) DO UPDATE SET
		current_chapter = excluded.current_chapter,
		scroll_position = excluded.scroll_position,
		percentage = excluded.percentage,
		updated_at = excluded.updated_at
	`
	_, err := s.db.Exec(query, progress.ID, progress.BookID, progress.UserID,
		progress.CurrentChapter, progress.ScrollPosition, progress.Percentage, progress.UpdatedAt)
	return err
}

// BookSource CRUD

func (s *SQLiteStore) CreateBookSource(source *models.BookSource) error {
	query := `
	INSERT INTO book_sources (id, name, url_template, encoding, book_name_rule, author_rule, content_rule, chapter_list_rule, chapter_url_rule, enabled, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, source.ID, source.Name, source.URLTemplate, source.Encoding,
		source.BookNameRule, source.AuthorRule, source.ContentRule, source.ChapterRule,
		source.ChapterURLRule, source.Enabled, source.CreatedAt, source.UpdatedAt)
	return err
}

func (s *SQLiteStore) GetBookSource(id string) (*models.BookSource, error) {
	query := `
	SELECT id, name, url_template, encoding, book_name_rule, author_rule, content_rule, chapter_list_rule, chapter_url_rule, enabled, created_at, updated_at
	FROM book_sources WHERE id = ?
	`
	row := s.db.QueryRow(query, id)

	var source models.BookSource
	var enabled int
	err := row.Scan(&source.ID, &source.Name, &source.URLTemplate, &source.Encoding,
		&source.BookNameRule, &source.AuthorRule, &source.ContentRule, &source.ChapterRule,
		&source.ChapterURLRule, &enabled, &source.CreatedAt, &source.UpdatedAt)
	if err != nil {
		return nil, err
	}
	source.Enabled = enabled == 1
	return &source, nil
}

func (s *SQLiteStore) ListBookSources() ([]models.BookSource, error) {
	query := `
	SELECT id, name, url_template, encoding, book_name_rule, author_rule, content_rule, chapter_list_rule, chapter_url_rule, enabled, created_at, updated_at
	FROM book_sources ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []models.BookSource
	for rows.Next() {
		var source models.BookSource
		var enabled int
		if err := rows.Scan(&source.ID, &source.Name, &source.URLTemplate, &source.Encoding,
			&source.BookNameRule, &source.AuthorRule, &source.ContentRule, &source.ChapterRule,
			&source.ChapterURLRule, &enabled, &source.CreatedAt, &source.UpdatedAt); err != nil {
			return nil, err
		}
		source.Enabled = enabled == 1
		sources = append(sources, source)
	}
	return sources, nil
}

func (s *SQLiteStore) UpdateBookSource(source *models.BookSource) error {
	source.UpdatedAt = time.Now()
	query := `
	UPDATE book_sources SET name = ?, url_template = ?, encoding = ?, book_name_rule = ?,
		author_rule = ?, content_rule = ?, chapter_list_rule = ?, chapter_url_rule = ?,
		enabled = ?, updated_at = ?
	WHERE id = ?
	`
	enabled := 0
	if source.Enabled {
		enabled = 1
	}
	_, err := s.db.Exec(query, source.Name, source.URLTemplate, source.Encoding,
		source.BookNameRule, source.AuthorRule, source.ContentRule, source.ChapterRule,
		source.ChapterURLRule, enabled, source.UpdatedAt, source.ID)
	return err
}

func (s *SQLiteStore) DeleteBookSource(id string) error {
	query := `DELETE FROM book_sources WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// CleanupRule CRUD

func (s *SQLiteStore) CreateCleanupRule(rule *models.CleanupRule) error {
	query := `
	INSERT INTO cleanup_rules (id, name, pattern, replacement, rule_type, enabled, priority, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, rule.ID, rule.Name, rule.Pattern, rule.Replacement,
		rule.RuleType, rule.Enabled, rule.Priority, rule.CreatedAt)
	return err
}

func (s *SQLiteStore) ListCleanupRules() ([]models.CleanupRule, error) {
	query := `
	SELECT id, name, pattern, replacement, rule_type, enabled, priority, created_at
	FROM cleanup_rules ORDER BY priority DESC, created_at ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.CleanupRule
	for rows.Next() {
		var rule models.CleanupRule
		var enabled int
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Pattern, &rule.Replacement,
			&rule.RuleType, &enabled, &rule.Priority, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rule.Enabled = enabled == 1
		rules = append(rules, rule)
	}
	return rules, nil
}

func (s *SQLiteStore) DeleteCleanupRule(id string) error {
	query := `DELETE FROM cleanup_rules WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// reading_settings 表已存在
func (s *SQLiteStore) SaveSettings(settings *models.ReadingSettings) error {
	query := `
	INSERT INTO reading_settings (id, user_id, font_size, line_height, theme, font)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(user_id) DO UPDATE SET
		font_size = excluded.font_size,
		line_height = excluded.line_height,
		theme = excluded.theme,
		font = excluded.font
	`
	_, err := s.db.Exec(query, settings.ID, settings.UserID, settings.FontSize, settings.LineHeight, settings.Theme, settings.Font)
	return err
}

func (s *SQLiteStore) GetSettings(userID string) (*models.ReadingSettings, error) {
	query := `SELECT id, user_id, font_size, line_height, theme, font FROM reading_settings WHERE user_id = ?`
	row := s.db.QueryRow(query, userID)

	var settings models.ReadingSettings
	err := row.Scan(&settings.ID, &settings.UserID, &settings.FontSize, &settings.LineHeight, &settings.Theme, &settings.Font)
	if err != nil {
		// 返回默认设置
		return &models.ReadingSettings{
			UserID:     userID,
			FontSize:   18,
			LineHeight: 1.8,
			Theme:      "light",
			Font:       "Georgia",
		}, nil
	}
	return &settings, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
