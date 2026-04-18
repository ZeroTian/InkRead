package services

import (
	"fmt"
	"inkread/models"
	"inkread/storage"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// WebImportService Web 导入服务
type WebImportService struct {
	store          *storage.SQLiteStore
	scraperService  *ScraperService
	cleanupService  *CleanupService
	uploadDir       string
}

func NewWebImportService(store *storage.SQLiteStore, uploadDir string) *WebImportService {
	return &WebImportService{
		store:         store,
		scraperService: NewScraperService(),
		cleanupService: NewCleanupService(),
		uploadDir:     uploadDir,
	}
}

// ImportFromURL 从 URL 导入书籍
func (s *WebImportService) ImportFromURL(url string, sourceID string) (*models.Book, error) {
	// 获取书源
	var source *BookSource
	if sourceID != "" {
		src, err := s.store.GetBookSource(sourceID)
		if err != nil {
			return nil, fmt.Errorf("获取书源失败: %w", err)
		}
		source = &BookSource{
			URLTemplate:    src.URLTemplate,
			Encoding:       src.Encoding,
			BookNameRule:   src.BookNameRule,
			AuthorRule:     src.AuthorRule,
			ContentRule:    src.ContentRule,
			ChapterRule:    src.ChapterRule,
			ChapterURLRule: src.ChapterURLRule,
		}
	}

	// 如果没有书源，尝试自动检测
	if source == nil {
		source = &BookSource{
			URLTemplate: url,
			Encoding:   "utf-8",
			// 使用通用选择器
			BookNameRule: "h1",
			AuthorRule:   ".author, .info",
			ContentRule:  ".content, #content, .chapter, article",
			ChapterRule:  "a[href]",
		}
	}

	// 抓取页面
	html, err := s.scraperService.Fetch(url)
	if err != nil {
		return nil, fmt.Errorf("抓取页面失败: %w", err)
	}

	// 解析书籍信息
	parser := &SourceParser{source: source}
	bookInfo, err := parser.ParseBookInfo(html)
	if err != nil {
		return nil, fmt.Errorf("解析书籍信息失败: %w", err)
	}

	// 如果没有获取到标题，使用 URL 提取
	if bookInfo.Title == "" {
		bookInfo.Title = extractTitleFromURL(url)
	}

	// 提取正文内容并净化
	content, err := parser.ParseContent(html)
	if err != nil {
		return nil, fmt.Errorf("提取内容失败: %w", err)
	}

	// 净化内容
	cleanContent := s.cleanupService.Cleanup(content)

	// 保存为文本文件
	bookID := uuid.New().String()
	filename := bookID + ".txt"
	filePath := filepath.Join(s.uploadDir, filename)

	err = os.WriteFile(filePath, []byte(cleanContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 创建书籍记录
	book := &models.Book{
		ID:        bookID,
		Title:     bookInfo.Title,
		Author:    bookInfo.Author,
		FilePath:  filePath,
		FileSize:  int64(len(cleanContent)),
		FileType:  "txt",
	}

	if err := s.store.CreateBook(book); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("保存书籍记录失败: %w", err)
	}

	return book, nil
}

// TestSource 测试书源
func (s *WebImportService) TestSource(url string, sourceID string) (*models.SourceTestResponse, error) {
	var source *BookSource
	if sourceID != "" {
		src, err := s.store.GetBookSource(sourceID)
		if err != nil {
			return nil, fmt.Errorf("获取书源失败: %w", err)
		}
		source = &BookSource{
			URLTemplate:    src.URLTemplate,
			Encoding:       src.Encoding,
			BookNameRule:   src.BookNameRule,
			AuthorRule:     src.AuthorRule,
			ContentRule:    src.ContentRule,
			ChapterRule:    src.ChapterRule,
			ChapterURLRule: src.ChapterURLRule,
		}
	}

	if source == nil {
		source = &BookSource{
			URLTemplate: url,
			Encoding:   "utf-8",
			BookNameRule: "h1",
			AuthorRule:   ".author, .info",
			ContentRule:  ".content, #content, .chapter, article",
			ChapterRule:  "a[href]",
		}
	}

	html, err := s.scraperService.Fetch(url)
	if err != nil {
		return &models.SourceTestResponse{
			Success: false,
			Error:   fmt.Sprintf("抓取失败: %v", err),
		}, nil
	}

	parser := &SourceParser{source: source}
	bookInfo, _ := parser.ParseBookInfo(html)

	content, _ := parser.ParseContent(html)
	cleanContent := s.cleanupService.Cleanup(content)

	return &models.SourceTestResponse{
		Success: true,
		Content: &models.ChapterContent{
			Title:   bookInfo.Title,
			Content: truncateString(cleanContent, 500),
		},
	}, nil
}

// extractTitleFromURL 从 URL 提取标题
func extractTitleFromURL(url string) string {
	// 从 URL 路径提取
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// 移除扩展名
		if idx := strings.Index(lastPart, "."); idx != -1 {
			lastPart = lastPart[:idx]
		}
		// 替换下划线和连字符为空格
		lastPart = strings.ReplaceAll(lastPart, "_", " ")
		lastPart = strings.ReplaceAll(lastPart, "-", " ")
		return lastPart
	}
	return "未命名"
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
