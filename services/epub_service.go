package services

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

type EPUBChapter struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type EPUBBook struct {
	Title    string       `json:"title"`
	Author   string       `json:"author"`
	Chapters []EPUBChapter `json:"chapters"`
}

type container struct {
	Rootfiles rootfiles `xml:"rootfiles"`
}

type rootfiles struct {
	Rootfile rootfile `xml:"rootfile"`
}

type rootfile struct {
	FullPath string `xml:"full-path,attr"`
}

type packageDoc struct {
	Metadata metadata `xml:"metadata"`
	Manifest manifest `xml:"manifest"`
	Spine    spine    `xml:"spine"`
}

type metadata struct {
	XMLNS     string `xml:"xmlns,attr"`
	Title     string `xml:"title"`
	Creator   string `xml:"creator"`
}

type manifest struct {
	Items []item `xml:"item"`
}

type item struct {
	ID      string `xml:"id,attr"`
	Href    string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

type spine struct {
	Itemrefs []itemref `xml:"itemref"`
}

type itemref struct {
	IDRef string `xml:"idref,attr"`
}

func ParseEPUB(filePath string) (*EPUBBook, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open epub: %w", err)
	}
	defer r.Close()

	containerData, err := readFile(r, "META-INF/container.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to read container.xml: %w", err)
	}

	var container container
	if err := xml.Unmarshal(containerData, &container); err != nil {
		return nil, fmt.Errorf("failed to parse container.xml: %w", err)
	}

	opfPath := container.Rootfiles.Rootfile.FullPath
	opfDir := filepath.Dir(opfPath)

	opfData, err := readFile(r, opfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read opf: %w", err)
	}

	var pkg packageDoc
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse opf: %w", err)
	}

	manifest := make(map[string]item)
	for _, i := range pkg.Manifest.Items {
		manifest[i.ID] = i
	}

	var chapters []EPUBChapter
	for _, ref := range pkg.Spine.Itemrefs {
		if item, ok := manifest[ref.IDRef]; ok {
			chapterPath := opfDir + "/" + item.Href
			chapterData, err := readFile(r, chapterPath)
			if err != nil {
				continue
			}

			title := extractTitle(string(chapterData), item.Href)
			content := stripHTML(string(chapterData))

			chapters = append(chapters, EPUBChapter{
				Title:   title,
				Content: content,
			})
		}
	}

	return &EPUBBook{
		Title:    pkg.Metadata.Title,
		Author:   pkg.Metadata.Creator,
		Chapters: chapters,
	}, nil
}

func readFile(r *zip.ReadCloser, name string) ([]byte, error) {
	for _, f := range r.File {
		if filepath.ToSlash(f.Name) == filepath.ToSlash(name) {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func extractTitle(htmlContent, fallback string) string {
	content := strings.ToLower(htmlContent)

	titleStart := strings.Index(content, "<title")
	if titleStart != -1 {
		titleEnd := strings.Index(content, "</title>")
		if titleEnd != -1 && titleEnd > titleStart {
			titleContent := htmlContent[titleStart:titleEnd]
			openTagEnd := strings.Index(titleContent, ">")
			if openTagEnd != -1 {
				return strings.TrimSpace(html.UnescapeString(titleContent[openTagEnd+1:]))
			}
		}
	}

	h1Start := strings.Index(content, "<h1")
	if h1Start != -1 {
		h1End := strings.Index(content, "</h1>")
		if h1End != -1 && h1End > h1Start {
			h1Content := htmlContent[h1Start:h1End]
			openTagEnd := strings.Index(h1Content, ">")
			if openTagEnd != -1 {
				return strings.TrimSpace(html.UnescapeString(h1Content[openTagEnd+1:]))
			}
		}
	}

	h2Start := strings.Index(content, "<h2")
	if h2Start != -1 {
		h2End := strings.Index(content, "</h2>")
		if h2End != -1 && h2End > h2Start {
			h2Content := htmlContent[h2Start:h2End]
			openTagEnd := strings.Index(h2Content, ">")
			if openTagEnd != -1 {
				return strings.TrimSpace(html.UnescapeString(h2Content[openTagEnd+1:]))
			}
		}
	}

	return strings.TrimSuffix(filepath.Base(fallback), filepath.Ext(fallback))
}

func stripHTML(htmlContent string) string {
	content := htmlContent

	// Use regexp to remove script, style, head, nav tags and their content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	headRegex := regexp.MustCompile(`(?i)<head[^>]*>.*?</head>`)
	navRegex := regexp.MustCompile(`(?i)<nav[^>]*>.*?</nav>`)

	content = scriptRegex.ReplaceAllString(content, "")
	content = styleRegex.ReplaceAllString(content, "")
	content = headRegex.ReplaceAllString(content, "")
	content = navRegex.ReplaceAllString(content, "")

	// Replace specific tags with newlines
	replacements := []string{
		`<br\s*/?>`, "\n",
		`</p>`, "\n",
		`<p[^>]*>`, "\n",
		`<div[^>]*>`, "\n",
		`<h[1-6][^>]*>`, "\n",
		`</h[1-6]>`, "\n",
	}

	for i := 0; i < len(replacements); i += 2 {
		re := regexp.MustCompile(`(?i)` + replacements[i])
		content = re.ReplaceAllString(content, replacements[i+1])
	}

	// Remove all remaining HTML tags
	var result strings.Builder
	inTag := false
	for _, r := range content {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	resultStr := result.String()
	// Collapse multiple newlines into single newlines
	resultStr = regexp.MustCompile(`\n+`).ReplaceAllString(resultStr, "\n")
	resultStr = strings.TrimSpace(resultStr)
	return html.UnescapeString(resultStr)
}
