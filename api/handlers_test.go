package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"inkread/models"
	"inkread/services"
	"inkread/storage"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestHandler(t *testing.T) (*Handlers, *storage.SQLiteStore, string, func()) {
	tmpDir, err := os.MkdirTemp("", "inkread_api_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	tmpDB, err := os.CreateTemp("", "inkread_api_test_*.db")
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

	bookService := services.NewBookService(store, tmpDir)
	aiService := services.NewAIService("", "test-model")
	handlers := NewHandlers(bookService, aiService)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
		os.Remove(tmpDB.Name())
	}

	return handlers, store, tmpDir, cleanup
}

func TestHealthEndpoint(t *testing.T) {
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestListBooksHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	// Test empty list
	req, _ := http.NewRequest("GET", "/api/books", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Code != 200 {
		t.Errorf("expected code 200, got %d", resp.Code)
	}
}

func TestUploadBookHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.epub")
	io.WriteString(part, "test content")
	writer.WriteField("author", "Test Author")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Code != 200 {
		t.Errorf("expected code 200, got %d", resp.Code)
	}
}

func TestUploadBookNoFile(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetBookHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	// First upload a book
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.epub")
	io.WriteString(part, "test content")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var uploadResp Response
	json.Unmarshal(w.Body.Bytes(), &uploadResp)
	data := uploadResp.Data.(map[string]interface{})
	bookID := data["id"].(string)

	// Then get it
	req, _ = http.NewRequest("GET", "/api/books/"+bookID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetBookNotFoundHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	req, _ := http.NewRequest("GET", "/api/books/non-existent-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestDeleteBookHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	// First upload a book
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.epub")
	io.WriteString(part, "test content")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var uploadResp Response
	json.Unmarshal(w.Body.Bytes(), &uploadResp)
	data := uploadResp.Data.(map[string]interface{})
	bookID := data["id"].(string)

	// Then delete it
	req, _ = http.NewRequest("DELETE", "/api/books/"+bookID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify it's gone
	req, _ = http.NewRequest("GET", "/api/books/"+bookID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404 after delete, got %d", w.Code)
	}
}

func TestSaveProgressHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	// First upload a book
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.epub")
	io.WriteString(part, "test content")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var uploadResp Response
	json.Unmarshal(w.Body.Bytes(), &uploadResp)
	data := uploadResp.Data.(map[string]interface{})
	bookID := data["id"].(string)

	// Save progress
	progressReq := models.ReadingProgress{
		BookID:          bookID,
		UserID:          "default",
		CurrentChapter: 5,
		ScrollPosition: 0.75,
		Percentage:     75.0,
	}
	progressJSON, _ := json.Marshal(progressReq)

	req, _ = http.NewRequest("POST", "/api/progress/"+bookID, strings.NewReader(string(progressJSON)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestGetProgressHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	// First upload a book
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.epub")
	io.WriteString(part, "test content")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var uploadResp Response
	json.Unmarshal(w.Body.Bytes(), &uploadResp)
	data := uploadResp.Data.(map[string]interface{})
	bookID := data["id"].(string)

	// Save progress first
	progressReq := models.ReadingProgress{
		BookID:          bookID,
		UserID:          "default",
		CurrentChapter: 3,
		ScrollPosition: 0.5,
		Percentage:     50.0,
	}
	progressJSON, _ := json.Marshal(progressReq)

	req, _ = http.NewRequest("POST", "/api/progress/"+bookID, strings.NewReader(string(progressJSON)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Get progress
	req, _ = http.NewRequest("GET", "/api/progress/"+bookID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	progressData := resp.Data.(map[string]interface{})

	if int(progressData["current_chapter"].(float64)) != 3 {
		t.Errorf("expected chapter 3, got %v", progressData["current_chapter"])
	}
}

func TestGetProgressNotFoundHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	req, _ := http.NewRequest("GET", "/api/progress/non-existent-book", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestSummarizeBookHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	// First upload a book
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.epub")
	io.WriteString(part, "This is a test book content for summarization")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var uploadResp Response
	json.Unmarshal(w.Body.Bytes(), &uploadResp)
	data := uploadResp.Data.(map[string]interface{})
	bookID := data["id"].(string)

	// Summarize book
	summarizeReq := models.SummarizeRequest{BookID: bookID}
	summarizeJSON, _ := json.Marshal(summarizeReq)

	req, _ = http.NewRequest("POST", "/api/ai/summarize", strings.NewReader(string(summarizeJSON)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestSummarizeBookNotFoundHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	summarizeReq := models.SummarizeRequest{BookID: "non-existent"}
	summarizeJSON, _ := json.Marshal(summarizeReq)

	req, _ := http.NewRequest("POST", "/api/ai/summarize", strings.NewReader(string(summarizeJSON)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestSummarizeBookMissingBookID(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	req, _ := http.NewRequest("POST", "/api/ai/summarize", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetBookContentHandler(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	// First upload a book
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.epub")
	io.WriteString(part, "This is test book content")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/books", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var uploadResp Response
	json.Unmarshal(w.Body.Bytes(), &uploadResp)
	data := uploadResp.Data.(map[string]interface{})
	bookID := data["id"].(string)

	// Get content
	req, _ = http.NewRequest("GET", "/api/books/"+bookID+"/content", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestListBooksPagination(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	// Upload multiple books
	for i := 0; i < 5; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.epub")
		io.WriteString(part, "content")
		writer.Close()

		req, _ := http.NewRequest("POST", "/api/books", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	// Test pagination
	req, _ := http.NewRequest("GET", "/api/books?page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})

	if data["total"] != 5.0 {
		t.Errorf("expected total 5, got %v", data["total"])
	}
	if len(data["books"].([]interface{})) != 2 {
		t.Errorf("expected 2 books, got %d", len(data["books"].([]interface{})))
	}
}

func TestSaveProgressInvalidRequest(t *testing.T) {
	h, _, _, cleanup := setupTestHandler(t)
	defer cleanup()

	r := gin.New()
	h.RegisterRoutes(r)

	req, _ := http.NewRequest("POST", "/api/progress/test-book", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// Helper to suppress time check
func TestResponseStructure(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, Response{
			Code:    200,
			Message: "success",
			Data:    map[string]interface{}{"test": "value"},
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != 200 {
		t.Errorf("expected code 200, got %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("expected message 'success', got %q", resp.Message)
	}
	if resp.Data == nil {
		t.Error("data should not be nil")
	}
}

var _ = time.Time{} // suppress unused import
