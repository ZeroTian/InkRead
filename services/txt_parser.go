package services

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"
)

// TXTParser TXT 解析器
type TXTParser struct{}

// NewTXTParser 创建 TXT 解析器
func NewTXTParser() *TXTParser {
	return &TXTParser{}
}

// ChapterInfo 章节信息
type ChapterInfo struct {
	Title   string
	Content string
	URL     string
}

// ParseContent 解析内容
func (p *TXTParser) ParseContent(data []byte) (string, error) {
	content, _ := p.detectAndConvert(data)
	return content, nil
}

// detectAndConvert 检测编码并转换为 UTF-8
func (p *TXTParser) detectAndConvert(data []byte) (string, string) {
	// 检查 BOM
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return string(data[3:]), "utf-8-bom"
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return p.utf16leToUtf8(data[2:]), "utf-16-le"
	}
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return p.utf16beToUtf8(data[2:]), "utf-16-be"
	}

	// 检测 GB2312/GBK/GB18030
	if p.isGBK(data) {
		return p.gbkToUtf8(data), "gbk"
	}

	return string(data), "utf-8"
}

// isGBK 检测是否为 GBK 编码
func (p *TXTParser) isGBK(data []byte) bool {
	gbCount := 0
	for i := 0; i < len(data)-1; i++ {
		if data[i] >= 0x81 && data[i] <= 0xFE && data[i+1] >= 0x40 && data[i+1] <= 0xFE {
			gbCount++
		}
	}
	return gbCount > len(data)/50
}

// gbkToUtf8 GBK 转 UTF-8
func (p *TXTParser) gbkToUtf8(data []byte) string {
	return string(data)
}

func (p *TXTParser) utf16leToUtf8(data []byte) string {
	result := make([]rune, 0, len(data)/2)
	for i := 0; i < len(data)-1; i += 2 {
		result = append(result, rune(data[i])|(rune(data[i+1])<<8))
	}
	return string(result)
}

func (p *TXTParser) utf16beToUtf8(data []byte) string {
	result := make([]rune, 0, len(data)/2)
	for i := 0; i < len(data)-1; i += 2 {
		result = append(result, (rune(data[i])<<8)|rune(data[i+1]))
	}
	return string(result)
}

// SplitChapters 按章节分割 TXT
func (p *TXTParser) SplitChapters(content string) []ChapterInfo {
	var chapters []ChapterInfo

	// 常见章节匹配模式
	patterns := []string{
		`第[一二三四五六七八九十百千零\d]+[章回节篇部卷]\s*[:：]?\s*(.*)`,
		`^[零一二三四五六七八九十百千\d]+[、，,]\s*(.*)$`,
		`(?:^|\n)\s*([一二三四五六七八九十]+)\s+(.*?)\s*(?:\n|$)`,
	}

	var selectedPattern string
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		if len(matches) >= 3 {
			selectedPattern = pattern
			break
		}
	}

	if selectedPattern == "" {
		// 没有匹配到章节模式，按固定长度分割
		return p.splitByLength(content)
	}

	re := regexp.MustCompile(selectedPattern)
	lines := strings.Split(content, "\n")

	var currentChapter ChapterInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if matches != nil {
			if currentChapter.Title != "" {
				chapters = append(chapters, currentChapter)
			}
			title := line
			if len(matches) > 1 && matches[1] != "" {
				title = matches[1]
			}
			currentChapter = ChapterInfo{
				Title:   title,
				Content: "",
			}
		} else {
			if currentChapter.Title != "" {
				currentChapter.Content += "\n" + line
			}
		}
	}

	if currentChapter.Title != "" {
		chapters = append(chapters, currentChapter)
	}

	return chapters
}

// splitByLength 按固定长度分割
func (p *TXTParser) splitByLength(content string) []ChapterInfo {
	const chapterSize = 5000 // 每章约 5000 字符
	runes := []rune(content)
	var chapters []ChapterInfo

	for i := 0; i < len(runes); i += chapterSize {
		end := i + chapterSize
		if end > len(runes) {
			end = len(runes)
		}
		chapters = append(chapters, ChapterInfo{
			Title:   "第" + string(rune('一'+len(chapters))) + "章",
			Content: string(runes[i:end]),
		})
	}

	return chapters
}

// ExtractTitle 提取书名
func (p *TXTParser) ExtractTitle(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Scan()
	firstLine := strings.TrimSpace(scanner.Text())

	// 检查第一行是否是书名模式
	patterns := []string{
		`书名[：:]\s*(.+)`,
		`《(.+)》`,
		`(.+)\s*$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(firstLine)
		if matches != nil && len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// 返回第一行作为标题（截断）
	if len(firstLine) > 50 {
		return firstLine[:50] + "..."
	}
	return firstLine
}

// CleanContent 清理 TXT 内容
func (p *TXTParser) CleanContent(content string) string {
	lines := strings.Split(content, "\n")
	var cleaned []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行
		if line == "" {
			continue
		}
		// 跳过纯数字行（页码等）
		if matched, _ := regexp.MatchString(`^\d+$`, line); matched {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// ReadChapter 从流中读取章节
func (p *TXTParser) ReadChapter(r io.Reader) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	if err != nil {
		return "", err
	}
	content, _ := p.ParseContent(buf.Bytes())
	chapters := p.SplitChapters(content)
	if len(chapters) > 0 {
		return chapters[0].Content, nil
	}
	return content, nil
}
