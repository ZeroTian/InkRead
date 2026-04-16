package services

import (
	"fmt"
	"io"
	"inkread/models"
	"inkread/storage"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type BookService struct {
	store     *storage.SQLiteStore
	uploadDir string
}

func NewBookService(store *storage.SQLiteStore, uploadDir string) *BookService {
	return &BookService{
		store:     store,
		uploadDir: uploadDir,
	}
}

func (s *BookService) UploadBook(file *models.UploadedFile) (*models.Book, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".epub" && ext != ".txt" && ext != ".pdf" {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	id := uuid.New().String()
	filename := id + ext
	filePath := filepath.Join(s.uploadDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file.Data); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	title := strings.TrimSuffix(file.Filename, ext)
	book := &models.Book{
		ID:        id,
		Title:     title,
		Author:    file.Author,
		FilePath:  filePath,
		FileSize:  file.Size,
		FileType:  ext[1:],
		CreatedAt: file.UploadedAt,
		UpdatedAt: file.UploadedAt,
	}

	if err := s.store.CreateBook(book); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save book: %w", err)
	}

	return book, nil
}

func (s *BookService) GetBook(id string) (*models.Book, error) {
	return s.store.GetBook(id)
}

func (s *BookService) ListBooks(page, pageSize int) (*models.BookListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	books, total, err := s.store.ListBooks(page, pageSize)
	if err != nil {
		return nil, err
	}

	if books == nil {
		books = []models.Book{}
	}

	return &models.BookListResponse{
		Books:    books,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *BookService) DeleteBook(id string) error {
	book, err := s.store.GetBook(id)
	if err != nil {
		return err
	}

	if err := os.Remove(book.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return s.store.DeleteBook(id)
}

func (s *BookService) GetBookContent(id string) (string, error) {
	book, err := s.store.GetBook(id)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(book.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(data), nil
}

func (s *BookService) GetEPUBContent(id string) (*EPUBBook, error) {
	book, err := s.store.GetBook(id)
	if err != nil {
		return nil, err
	}

	if book.FileType != "epub" {
		return nil, fmt.Errorf("not an epub file")
	}

	return ParseEPUB(book.FilePath)
}

func (s *BookService) GetReadingProgress(bookID string) (*models.ReadingProgress, error) {
	return s.store.GetProgress(bookID)
}

func (s *BookService) SaveReadingProgress(progress *models.ReadingProgress) error {
	if progress.ID == "" {
		progress.ID = uuid.New().String()
	}
	if progress.UserID == "" {
		progress.UserID = "default"
	}
	return s.store.SaveProgress(progress)
}
