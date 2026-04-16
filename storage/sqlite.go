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

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
