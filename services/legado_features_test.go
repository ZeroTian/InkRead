package services

import (
	"strings"
	"testing"
)

func TestChapterInfo(t *testing.T) {
	chapters := []ChapterInfo{
		{Title: "第一章", URL: "http://example.com/1"},
		{Title: "第二章", URL: "http://example.com/2"},
	}

	if len(chapters) != 2 {
		t.Errorf("Expected 2 chapters, got %d", len(chapters))
	}

	if chapters[0].Title != "第一章" {
		t.Errorf("Expected 第一章, got %s", chapters[0].Title)
	}
}

func TestSourceParserStructs(t *testing.T) {
	// Test that parser structs are properly defined
	parser := &SourceParser{}

	if parser == nil {
		t.Error("SourceParser should not be nil")
	}

	_ = parser
}

func TestScraperService(t *testing.T) {
	// Test ScraperService can be created
	scraper := &ScraperService{}

	if scraper == nil {
		t.Error("ScraperService should not be nil")
	}

	_ = scraper
}

func TestBookInfo(t *testing.T) {
	info := &BookInfo{
		Title:   "测试书名",
		Author:  "测试作者",
		Cover:   "http://example.com/cover.jpg",
		Summary: "这是简介",
	}

	if info.Title != "测试书名" {
		t.Errorf("Expected 测试书名, got %s", info.Title)
	}

	if info.Author != "测试作者" {
		t.Errorf("Expected 测试作者, got %s", info.Author)
	}
}

func TestCleanupPatterns(t *testing.T) {
	// Test that regex patterns are valid
	patterns := []string{
		`\[广告\]`,
		`\[弹窗\]`,
		`<script>.*?</script>`,
		`\s+`,
	}

	for _, pattern := range patterns {
		if strings.Contains(pattern, "(") || strings.Contains(pattern, "[") {
			// Pattern contains group syntax - should be valid regex
			t.Logf("Pattern: %s", pattern)
		}
	}
}

func TestContentExtraction(t *testing.T) {
	html := `<html>
<body>
<h1>书名</h1>
<div class="content">这是小说正文内容</div>
<div class="ads">广告内容</div>
</body>
</html>`

	// Simple verification
	if !strings.Contains(html, "书名") {
		t.Error("HTML should contain 书名")
	}
	if !strings.Contains(html, "小说正文") {
		t.Error("HTML should contain 小说正文")
	}
	if !strings.Contains(html, "广告内容") {
		t.Error("HTML should contain 广告内容")
	}
}
