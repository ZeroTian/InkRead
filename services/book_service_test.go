package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"inkread/models"
	"inkread/storage"
)

func setupTestBookService(t *testing.T) (*BookService, string, func()) {
	tmpDir, err := os.MkdirTemp("", "inkread_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	tmpDB, err := os.CreateTemp("", "inkread_test_*.db")
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create temp db: %v", err)
	}
	tmpDB.Close()

	store, err := storage.NewSQLiteStore(tmpDB.Name())
	if err != nil {
		os.RemoveAll(tmpDir)
		os.Remove(tmpDB.Name())
		t.Fatalf("failed to create store: %v", err)
	}

	bookService := NewBookService(store, tmpDir)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
		os.Remove(tmpDB.Name())
	}

	return bookService, tmpDir, cleanup
}

func TestNewBookService(t *testing.T) {
	svc, _, cleanup := setupTestBookService(t)
	defer cleanup()

	if svc == nil {
		t.Fatal("book service should not be nil")
	}
}

func TestUploadBook(t *testing.T) {
	svc, _, cleanup := setupTestBookService(t)
	defer cleanup()

	content := "Test book content"
	uploadedFile := &models.UploadedFile{
		Filename:   "test.epub",
		Data:       strings.NewReader(content),
		Size:       int64(len(content)),
		Author:     "Test Author",
		UploadedAt: time.Now(),
	}

	book, err := svc.UploadBook(uploadedFile)
	if err != nil {
		t.Fatalf("failed to upload book: %v", err)
	}

	if book.Title != "test" {
		t.Errorf("expected title 'test', got %q", book.Title)
	}
	if book.Author != "Test Author" {
		t.Errorf("expected author 'Test Author', got %q", book.Author)
	}
	if book.FileType != "epub" {
		t.Errorf("expected file type 'epub', got %q", book.FileType)
	}
	if book.ID == "" {
		t.Error("book ID should not be empty")
	}
}

func TestUploadBookUnsupportedType(t *testing.T) {
	svc, _, cleanup := setupTestBookService(t)
	defer cleanup()

	uploadedFile := &models.UploadedFile{
		Filename:   "test.xyz",
		Data:       strings.NewReader("content"),
		Size:       7,
		Author:     "Test Author",
		UploadedAt: time.Now(),
	}

	_, err := svc.UploadBook(uploadedFile)
	if err == nil {
		t.Error("expected error for unsupported file type")
	}
}

func TestGetBook(t *testing.T) {
	svc, _, cleanup := setupTestBookService(t)
	defer cleanup()

	uploadedFile := &models.UploadedFile{
		Filename:   "test.epub",
		Data:       strings.NewReader("content"),
		Size:       7,
		Author:     "Test Author",
		UploadedAt: time.Now(),
	}

	created, _ := svc.UploadBook(uploadedFile)

	book, err := svc.GetBook(created.ID)
	if err != nil {
		t.Fatalf("failed to get book: %v", err)
	}

	if book.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, book.ID)
	}
}

func TestListBooks(t *testing.T) {
	svc, _, cleanup := setupTestBookService(t)
	defer cleanup()

	// Create multiple books
	for i := 0; i < 3; i++ {
		uploadedFile := &models.UploadedFile{
			Filename:   "test.epub",
			Data:       strings.NewReader("content"),
			Size:       7,
			Author:     "Author",
			UploadedAt: time.Now(),
		}
		svc.UploadBook(uploadedFile)
	}

	result, err := svc.ListBooks(1, 10)
	if err != nil {
		t.Fatalf("failed to list books: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
	if len(result.Books) != 3 {
		t.Errorf("expected 3 books, got %d", len(result.Books))
	}
}

func TestListBooksPagination(t *testing.T) {
	svc, _, cleanup := setupTestBookService(t)
	defer cleanup()

	// Create 5 books
	for i := 0; i < 5; i++ {
		uploadedFile := &models.UploadedFile{
			Filename:   "test.epub",
			Data:       strings.NewReader("content"),
			Size:       7,
			Author:     "Author",
			UploadedAt: time.Now(),
		}
		svc.UploadBook(uploadedFile)
	}

	result, err := svc.ListBooks(1, 2)
	if err != nil {
		t.Fatalf("failed to list books: %v", err)
	}

	if result.Total != 5 {
		t.Errorf("expected total 5, got %d", result.Total)
	}
	if len(result.Books) != 2 {
		t.Errorf("expected 2 books, got %d", len(result.Books))
	}
	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
}

func TestDeleteBook(t *testing.T) {
	svc, tmpDir, cleanup := setupTestBookService(t)
	defer cleanup()

	uploadedFile := &models.UploadedFile{
		Filename:   "test.epub",
		Data:       strings.NewReader("content"),
		Size:       7,
		Author:     "Test Author",
		UploadedAt: time.Now(),
	}

	book, _ := svc.UploadBook(uploadedFile)

	// Verify file exists
	filePath := filepath.Join(tmpDir, book.ID+".epub")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected file to exist before deletion")
	}

	if err := svc.DeleteBook(book.ID); err != nil {
		t.Fatalf("failed to delete book: %v", err)
	}

	// Verify file no longer exists
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}

	// Verify book no longer retrievable
	_, err := svc.GetBook(book.ID)
	if err == nil {
		t.Error("expected error after deleting book")
	}
}

func TestGetBookContent(t *testing.T) {
	svc, _, cleanup := setupTestBookService(t)
	defer cleanup()

	content := "This is test book content"
	uploadedFile := &models.UploadedFile{
		Filename:   "test.epub",
		Data:       strings.NewReader(content),
		Size:       int64(len(content)),
		Author:     "Test Author",
		UploadedAt: time.Now(),
	}

	book, _ := svc.UploadBook(uploadedFile)

	retrievedContent, err := svc.GetBookContent(book.ID)
	if err != nil {
		t.Fatalf("failed to get book content: %v", err)
	}

	if retrievedContent != content {
		t.Errorf("expected content %q, got %q", content, retrievedContent)
	}
}

func TestSaveAndGetReadingProgress(t *testing.T) {
	svc, _, cleanup := setupTestBookService(t)
	defer cleanup()

	// Create a book first
	uploadedFile := &models.UploadedFile{
		Filename:   "test.epub",
		Data:       strings.NewReader("content"),
		Size:       7,
		Author:     "Test Author",
		UploadedAt: time.Now(),
	}
	book, _ := svc.UploadBook(uploadedFile)

	progress := &models.ReadingProgress{
		BookID:          book.ID,
		CurrentChapter: 3,
		ScrollPosition: 0.5,
		Percentage:     50.0,
	}

	if err := svc.SaveReadingProgress(progress); err != nil {
		t.Fatalf("failed to save progress: %v", err)
	}

	retrieved, err := svc.GetReadingProgress(book.ID)
	if err != nil {
		t.Fatalf("failed to get progress: %v", err)
	}

	if retrieved.CurrentChapter != 3 {
		t.Errorf("expected chapter 3, got %d", retrieved.CurrentChapter)
	}
	if retrieved.Percentage != 50.0 {
		t.Errorf("expected percentage 50.0, got %f", retrieved.Percentage)
	}
}

func TestUpdateReadingProgress(t *testing.T) {
	svc, _, cleanup := setupTestBookService(t)
	defer cleanup()

	// Create a book first
	uploadedFile := &models.UploadedFile{
		Filename:   "test.epub",
		Data:       strings.NewReader("content"),
		Size:       7,
		Author:     "Test Author",
		UploadedAt: time.Now(),
	}
	book, _ := svc.UploadBook(uploadedFile)

	progress := &models.ReadingProgress{
		BookID:          book.ID,
		CurrentChapter: 1,
		ScrollPosition: 0.0,
		Percentage:     0.0,
	}
	svc.SaveReadingProgress(progress)

	// Update
	progress.CurrentChapter = 10
	progress.Percentage = 100.0
	if err := svc.SaveReadingProgress(progress); err != nil {
		t.Fatalf("failed to update progress: %v", err)
	}

	retrieved, err := svc.GetReadingProgress(book.ID)
	if err != nil {
		t.Fatalf("failed to get updated progress: %v", err)
	}

	if retrieved.CurrentChapter != 10 {
		t.Errorf("expected chapter 10, got %d", retrieved.CurrentChapter)
	}
}
