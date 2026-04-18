package storage

import (
	"os"
	"testing"
	"time"

	"inkread/models"
)

func setupTestDB(t *testing.T) (*SQLiteStore, func()) {
	tmpFile, err := os.CreateTemp("", "test_inkread_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	store, err := NewSQLiteStore(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to create store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.Remove(tmpFile.Name())
	}

	return store, cleanup
}

func TestNewSQLiteStore(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	if store == nil {
		t.Fatal("store should not be nil")
	}
}

func TestCreateAndGetBook(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	book := &models.Book{
		ID:        "test-book-001",
		Title:     "Test Book",
		Author:    "Test Author",
		FilePath:  "/uploads/test.epub",
		FileSize:  1024000,
		FileType:  "epub",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.CreateBook(book); err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	retrieved, err := store.GetBook("test-book-001")
	if err != nil {
		t.Fatalf("failed to get book: %v", err)
	}

	if retrieved.Title != book.Title {
		t.Errorf("expected title %q, got %q", book.Title, retrieved.Title)
	}
	if retrieved.Author != book.Author {
		t.Errorf("expected author %q, got %q", book.Author, retrieved.Author)
	}
	if retrieved.FileSize != book.FileSize {
		t.Errorf("expected file size %d, got %d", book.FileSize, retrieved.FileSize)
	}
}

func TestGetBookNotFound(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := store.GetBook("non-existent")
	if err == nil {
		t.Error("expected error for non-existent book")
	}
}

func TestListBooks(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create multiple books
	for i := 1; i <= 5; i++ {
		book := &models.Book{
			ID:        "test-book-" + string(rune('0'+i)),
			Title:     "Book " + string(rune('0'+i)),
			Author:    "Author " + string(rune('0'+i)),
			FilePath:  "/uploads/test.epub",
			FileSize:  int64(i * 1000),
			FileType:  "epub",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := store.CreateBook(book); err != nil {
			t.Fatalf("failed to create book: %v", err)
		}
	}

	// Test pagination
	books, total, err := store.ListBooks(1, 3)
	if err != nil {
		t.Fatalf("failed to list books: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(books) != 3 {
		t.Errorf("expected 3 books on page 1, got %d", len(books))
	}

	// Test page 2
	books, _, err = store.ListBooks(2, 3)
	if err != nil {
		t.Fatalf("failed to list books page 2: %v", err)
	}
	if len(books) != 2 {
		t.Errorf("expected 2 books on page 2, got %d", len(books))
	}
}

func TestDeleteBook(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	book := &models.Book{
		ID:        "test-book-delete",
		Title:     "Delete Me",
		Author:    "Test Author",
		FilePath:  "/uploads/test.epub",
		FileSize:  1024000,
		FileType:  "epub",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.CreateBook(book); err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	if err := store.DeleteBook("test-book-delete"); err != nil {
		t.Fatalf("failed to delete book: %v", err)
	}

	_, err := store.GetBook("test-book-delete")
	if err == nil {
		t.Error("expected error after deleting book")
	}
}

func TestUpdateBook(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	book := &models.Book{
		ID:        "test-book-update",
		Title:     "Original Title",
		Author:    "Original Author",
		FilePath:  "/uploads/test.epub",
		FileSize:  1024000,
		FileType:  "epub",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.CreateBook(book); err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	book.Title = "Updated Title"
	book.Author = "Updated Author"
	if err := store.UpdateBook(book); err != nil {
		t.Fatalf("failed to update book: %v", err)
	}

	retrieved, err := store.GetBook("test-book-update")
	if err != nil {
		t.Fatalf("failed to get updated book: %v", err)
	}

	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", retrieved.Title)
	}
}

func TestSaveAndGetProgress(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	progress := &models.ReadingProgress{
		ID:              "progress-001",
		BookID:          "book-001",
		UserID:          "default",
		CurrentChapter: 5,
		ScrollPosition:  0.75,
		Percentage:      75.0,
		UpdatedAt:       time.Now(),
	}

	if err := store.SaveProgress(progress); err != nil {
		t.Fatalf("failed to save progress: %v", err)
	}

	retrieved, err := store.GetProgress("book-001")
	if err != nil {
		t.Fatalf("failed to get progress: %v", err)
	}

	if retrieved.CurrentChapter != 5 {
		t.Errorf("expected chapter 5, got %d", retrieved.CurrentChapter)
	}
	if retrieved.Percentage != 75.0 {
		t.Errorf("expected percentage 75.0, got %f", retrieved.Percentage)
	}
}

func TestUpdateProgress(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	progress := &models.ReadingProgress{
		ID:              "progress-002",
		BookID:          "book-002",
		UserID:          "default",
		CurrentChapter: 1,
		ScrollPosition:  0.0,
		Percentage:      0.0,
		UpdatedAt:       time.Now(),
	}

	if err := store.SaveProgress(progress); err != nil {
		t.Fatalf("failed to save progress: %v", err)
	}

	// Update the progress
	progress.CurrentChapter = 10
	progress.Percentage = 100.0
	if err := store.SaveProgress(progress); err != nil {
		t.Fatalf("failed to update progress: %v", err)
	}

	retrieved, err := store.GetProgress("book-002")
	if err != nil {
		t.Fatalf("failed to get updated progress: %v", err)
	}

	if retrieved.CurrentChapter != 10 {
		t.Errorf("expected chapter 10, got %d", retrieved.CurrentChapter)
	}
	if retrieved.Percentage != 100.0 {
		t.Errorf("expected percentage 100.0, got %f", retrieved.Percentage)
	}
}

func TestGetProgressNotFound(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := store.GetProgress("non-existent-book")
	if err == nil {
		t.Error("expected error for non-existent progress")
	}
}
