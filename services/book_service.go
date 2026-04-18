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

	// 如果是 TXT，使用 TXTParser 解析
	if book.FileType == "txt" {
		txtParser := NewTXTParser()
		content, _ := txtParser.ParseContent(data)
		return txtParser.CleanContent(content), nil
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

// GetTXTChapters 获取 TXT 章节列表
func (s *BookService) GetTXTChapters(id string) ([]ChapterInfo, error) {
	book, err := s.store.GetBook(id)
	if err != nil {
		return nil, err
	}

	if book.FileType != "txt" {
		return nil, fmt.Errorf("not a txt file")
	}

	data, err := os.ReadFile(book.FilePath)
	if err != nil {
		return nil, err
	}

	txtParser := NewTXTParser()
	content, _ := txtParser.ParseContent(data)
	cleanContent := txtParser.CleanContent(content)

	return txtParser.SplitChapters(cleanContent), nil
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

// BookSource methods

func (s *BookService) ListSources() ([]models.BookSource, error) {
	return s.store.ListBookSources()
}

func (s *BookService) CreateSource(source *models.BookSource) error {
	if source.ID == "" {
		source.ID = uuid.New().String()
	}
	if source.Encoding == "" {
		source.Encoding = "utf-8"
	}
	return s.store.CreateBookSource(source)
}

func (s *BookService) UpdateSource(source *models.BookSource) error {
	return s.store.UpdateBookSource(source)
}

func (s *BookService) DeleteSource(id string) error {
	return s.store.DeleteBookSource(id)
}

func (s *BookService) TestSource(url string) (*models.SourceTestResponse, error) {
	importService := NewWebImportService(s.store, s.uploadDir)
	return importService.TestSource(url, "")
}

func (s *BookService) ImportFromURL(url, sourceID string) (*models.Book, error) {
	importService := NewWebImportService(s.store, s.uploadDir)
	return importService.ImportFromURL(url, sourceID)
}

// CleanupRule methods

func (s *BookService) ListCleanupRules() ([]models.CleanupRule, error) {
	return s.store.ListCleanupRules()
}

func (s *BookService) CreateCleanupRule(rule *models.CleanupRule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	return s.store.CreateCleanupRule(rule)
}

func (s *BookService) DeleteCleanupRule(id string) error {
	return s.store.DeleteCleanupRule(id)
}
