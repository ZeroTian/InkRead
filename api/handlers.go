package api

import (
	"net/http"
	"inkread/models"
	"inkread/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	bookService *services.BookService
	aiService   *services.AIService
}

func NewHandlers(bookService *services.BookService, aiService *services.AIService) *Handlers {
	return &Handlers{
		bookService: bookService,
		aiService:   aiService,
	}
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (h *Handlers) success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

func (h *Handlers) error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func (h *Handlers) ListBooks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.bookService.ListBooks(page, pageSize)
	if err != nil {
		h.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(c, result)
}

func (h *Handlers) UploadBook(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.error(c, http.StatusBadRequest, "请选择要上传的文件")
		return
	}
	defer file.Close()

	author := c.PostForm("author")

	uploadedFile := &models.UploadedFile{
		Filename:   header.Filename,
		Data:       file,
		Size:       header.Size,
		Author:     author,
		UploadedAt: time.Now(),
	}

	book, err := h.bookService.UploadBook(uploadedFile)
	if err != nil {
		h.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(c, book)
}

func (h *Handlers) GetBook(c *gin.Context) {
	id := c.Param("id")

	book, err := h.bookService.GetBook(id)
	if err != nil {
		h.error(c, http.StatusNotFound, "书籍不存在")
		return
	}

	if book.FileType == "epub" {
		epubBook, err := h.bookService.GetEPUBContent(id)
		if err == nil {
			h.success(c, epubBook)
			return
		}
	}

	h.success(c, book)
}

func (h *Handlers) DeleteBook(c *gin.Context) {
	id := c.Param("id")

	if err := h.bookService.DeleteBook(id); err != nil {
		h.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(c, gin.H{"deleted": id})
}

func (h *Handlers) GetProgress(c *gin.Context) {
	bookID := c.Param("book_id")

	progress, err := h.bookService.GetReadingProgress(bookID)
	if err != nil {
		h.error(c, http.StatusNotFound, "未找到阅读进度")
		return
	}

	h.success(c, progress)
}

func (h *Handlers) SaveProgress(c *gin.Context) {
	bookID := c.Param("book_id")

	var req struct {
		CurrentChapter int     `json:"current_chapter"`
		ScrollPosition float64 `json:"scroll_position"`
		Percentage      float64 `json:"percentage"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.error(c, http.StatusBadRequest, "请求格式错误")
		return
	}

	progress := &models.ReadingProgress{
		BookID:          bookID,
		UserID:          "default",
		CurrentChapter: req.CurrentChapter,
		ScrollPosition:  req.ScrollPosition,
		Percentage:      req.Percentage,
	}

	if err := h.bookService.SaveReadingProgress(progress); err != nil {
		h.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(c, progress)
}

func (h *Handlers) GetBookContent(c *gin.Context) {
	id := c.Param("id")

	content, err := h.bookService.GetBookContent(id)
	if err != nil {
		h.error(c, http.StatusNotFound, "书籍不存在或读取失败")
		return
	}

	h.success(c, gin.H{
		"content": content,
	})
}

func (h *Handlers) SummarizeBook(c *gin.Context) {
	var req models.SummarizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.error(c, http.StatusBadRequest, "请提供 book_id")
		return
	}

	book, err := h.bookService.GetBook(req.BookID)
	if err != nil {
		h.error(c, http.StatusNotFound, "书籍不存在")
		return
	}

	content, err := h.bookService.GetBookContent(req.BookID)
	if err != nil {
		h.error(c, http.StatusInternalServerError, "无法读取书籍内容")
		return
	}

	result, err := h.aiService.SummarizeBook(content, book.Title)
	if err != nil {
		h.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.success(c, models.SummarizeResponse{
		Summary:   result.Summary,
		BookID:    req.BookID,
		Model:     result.Model,
		CreatedAt: result.CreatedAt.Format(time.RFC3339),
	})
}

func (h *Handlers) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/books", h.ListBooks)
		api.POST("/books", h.UploadBook)
		api.GET("/books/:id", h.GetBook)
		api.DELETE("/books/:id", h.DeleteBook)
		api.GET("/books/:id/content", h.GetBookContent)
		api.POST("/ai/summarize", h.SummarizeBook)
		api.GET("/progress/:book_id", h.GetProgress)
		api.POST("/progress/:book_id", h.SaveProgress)
	}
}
