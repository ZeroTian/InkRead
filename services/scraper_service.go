package services

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ScraperService 网页抓取服务
type ScraperService struct {
	client *http.Client
}

func NewScraperService() *ScraperService {
	return &ScraperService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Fetch 获取网页内容
func (s *ScraperService) Fetch(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取失败: %w", err)
	}

	// 检测编码
	encoding := detectEncoding(body)
	return convertToString(body, encoding), nil
}

// FetchWithSelector 使用 CSS Selector 提取内容
func (s *ScraperService) FetchWithSelector(url string, selector string) (string, error) {
	html, err := s.Fetch(url)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("解析HTML失败: %w", err)
	}

	content := doc.Find(selector).Text()
	return strings.TrimSpace(content), nil
}

// ExtractLinks 提取页面中所有链接
func (s *ScraperService) ExtractLinks(url string, selector string) ([]string, error) {
	html, err := s.Fetch(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %w", err)
	}

	var links []string
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && href != "" {
			links = append(links, href)
		}
	})

	return links, nil
}

// SourceParser 书源解析器
type SourceParser struct {
	source *BookSource
}

// BookSource 书源（精简版用于解析）
type BookSource struct {
	URLTemplate    string
	Encoding       string
	BookNameRule   string
	AuthorRule     string
	ContentRule    string
	ChapterRule    string
	ChapterURLRule string
}

// BookInfo 书籍信息
type BookInfo struct {
	Title   string
	Author  string
	Cover   string
	Summary string
}

// ParseBookInfo 解析书籍信息
func (p *SourceParser) ParseBookInfo(html string) (*BookInfo, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	info := &BookInfo{}

	if p.source.BookNameRule != "" {
		info.Title = doc.Find(p.source.BookNameRule).Text()
		info.Title = strings.TrimSpace(info.Title)
	}

	if p.source.AuthorRule != "" {
		info.Author = doc.Find(p.source.AuthorRule).Text()
		info.Author = strings.TrimSpace(info.Author)
	}

	return info, nil
}

// ParseChapters 解析章节列表
func (p *SourceParser) ParseChapters(html string) ([]ChapterInfo, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var chapters []ChapterInfo

	doc.Find(p.source.ChapterRule).Each(func(i int, s *goquery.Selection) {
		title := s.Text()
		title = strings.TrimSpace(title)

		url, _ := s.Attr("href")
		if url != "" {
			// 处理相对路径
			if strings.HasPrefix(url, "/") {
				url = "https://example.com" + url // 实际应该从页面URL推导
			}
		}

		if title != "" {
			chapters = append(chapters, ChapterInfo{
				Title: title,
				URL:   url,
			})
		}
	})

	return chapters, nil
}

// ParseContent 解析章节内容
func (p *SourceParser) ParseContent(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	content := doc.Find(p.source.ContentRule).Text()
	return strings.TrimSpace(content), nil
}

// 编码检测
func detectEncoding(body []byte) string {
	// 检查 BOM
	if len(body) >= 3 && body[0] == 0xEF && body[1] == 0xBB && body[2] == 0xBF {
		return "utf-8-bom"
	}
	if len(body) >= 2 && body[0] == 0xFF && body[1] == 0xFE {
		return "utf-16-le"
	}
	if len(body) >= 2 && body[0] == 0xFE && body[1] == 0xFF {
		return "utf-16-be"
	}

	// 简单检测是否包含 GB 编码特征
	gbCount := 0
	for i := 0; i < len(body)-1; i++ {
		if body[i] >= 0x81 && body[i] <= 0xFE && body[i+1] >= 0x40 && body[i+1] <= 0xFE {
			gbCount++
		}
	}

	// 如果 GB 特征超过一定比例，认为是 GB 编码
	if gbCount > len(body)/100 {
		return "gbk"
	}

	return "utf-8"
}

// 转换编码
func convertToString(body []byte, encoding string) string {
	switch encoding {
	case "gbk":
		// 简单的 GBK -> UTF-8 转换（实际应该用专门的库）
		return string(body)
	case "utf-8-bom":
		return string(body[3:])
	default:
		return string(body)
	}
}
