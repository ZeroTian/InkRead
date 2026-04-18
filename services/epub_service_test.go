package services

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func getTestEPUBPath(t *testing.T) string {
	// Try multiple possible locations for test EPUB
	paths := []string{
		"../Pride_and_Prejudice.epub",
		"../../Pride_and_Prejudice.epub",
		"/mnt/e/code/inkread/Pride_and_Prejudice.epub",
		"Pride_and_Prejudice.epub",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Create a minimal valid EPUB for testing
	tmpDir := t.TempDir()
	epubPath := filepath.Join(tmpDir, "test.epub")
	createMinimalEPUB(t, epubPath)
	return epubPath
}

func createMinimalEPUB(t *testing.T, path string) {
	content := `PK     ! µ  META-INF/container.xml   <?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>PK     ! µ  OEBPS/content.opf   <?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:creator>Test Author</dc:creator>
  </metadata>
  <manifest>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter1"/>
  </spine>
</package>PK     ! µ  OEBPS/chapter1.xhtml   <?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html>
<head><title>Chapter 1</title></head>
<body>
<h1>Chapter 1: The Beginning</h1>
<p>This is the first chapter of the test book.</p>
<p>It contains multiple paragraphs.</p>
</body>
</html>`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test EPUB: %v", err)
	}
}

func TestParseEPUB(t *testing.T) {
	epubPath := getTestEPUBPath(t)

	book, err := ParseEPUB(epubPath)
	if err != nil {
		t.Fatalf("failed to parse EPUB: %v", err)
	}

	if book.Title == "" {
		t.Error("book title should not be empty")
	}
	if book.Author == "" {
		t.Error("book author should not be empty")
	}
	if len(book.Chapters) == 0 {
		t.Error("book should have at least one chapter")
	}
}

func TestParseEPUBChapters(t *testing.T) {
	epubPath := getTestEPUBPath(t)

	book, err := ParseEPUB(epubPath)
	if err != nil {
		t.Fatalf("failed to parse EPUB: %v", err)
	}

	for _, ch := range book.Chapters {
		if ch.Content == "" {
			t.Error("chapter content should not be empty")
		}
	}
}

func TestParseEPUBNotFound(t *testing.T) {
	_, err := ParseEPUB("/nonexistent/path.epub")
	if err == nil {
		t.Error("expected error for non-existent EPUB")
	}
}

func TestStripHTML(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "<p>Hello</p>",
			expected: "Hello",
		},
		{
			input:    "<h1>Title</h1><p>Paragraph</p>",
			expected: "Title\nParagraph",
		},
		{
			input:    "<script>alert('xss')</script><p>Safe</p>",
			expected: "Safe",
		},
		{
			input:    "<style>.class{color:red;}</style><p>Content</p>",
			expected: "Content",
		},
	}

	for _, tc := range testCases {
		result := stripHTML(tc.input)
		if result != tc.expected {
			t.Errorf("stripHTML(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestExtractTitle(t *testing.T) {
	testCases := []struct {
		html    string
		fallback string
		want    string
	}{
		{
			html:    "<html><head><title>Page Title</title></head></html>",
			fallback: "file.xhtml",
			want:    "Page Title",
		},
		{
			html:    "<body><h1>Heading 1</h1></body>",
			fallback: "file.xhtml",
			want:    "Heading 1",
		},
		{
			html:    "<body><h2>Heading 2</h2></body>",
			fallback: "file.xhtml",
			want:    "Heading 2",
		},
		{
			html:    "<body><p>No title here</p></body>",
			fallback: "chapter_one.xhtml",
			want:    "chapter_one",
		},
	}

	for _, tc := range testCases {
		got := extractTitle(tc.html, tc.fallback)
		if got != tc.want {
			t.Errorf("extractTitle(%q, %q) = %q, want %q", tc.html, tc.fallback, got, tc.want)
		}
	}
}

func TestReadFile(t *testing.T) {
	// This test requires a valid EPUB structure
	epubPath := getTestEPUBPath(t)

	r, err := zip.OpenReader(epubPath)
	if err != nil {
		t.Fatalf("failed to open EPUB: %v", err)
	}
	defer r.Close()

	data, err := readFile(r, "META-INF/container.xml")
	if err != nil {
		t.Fatalf("failed to read container.xml: %v", err)
	}

	if len(data) == 0 {
		t.Error("container.xml content should not be empty")
	}
}

func TestReadFileNotFound(t *testing.T) {
	epubPath := getTestEPUBPath(t)

	r, err := zip.OpenReader(epubPath)
	if err != nil {
		t.Fatalf("failed to open EPUB: %v", err)
	}
	defer r.Close()

	_, err = readFile(r, "nonexistent/file.xml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
