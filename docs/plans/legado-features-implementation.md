# InkRead Legado 功能增强实现计划

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** 为 InkRead 添加 Web URL 导入、书源管理、内容净化、TXT 增强和阅读体验增强功能

**Architecture:** 
- 后端 Go: 新增 web scraper service、book source management、content cleanup pipeline
- 数据库: 新增 book_sources 和 cleanup_rules 表
- 前端: 增强阅读器 UI，支持主题切换、字体大小等设置

**Tech Stack:** Go 1.21+, Gin, SQLite, chromedp/goquery (HTML parsing), jmoiron/sqlx

---

## Phase 1: 基础设施增强

### Task 1: 添加 Go 依赖并创建项目结构

**Files:**
- Modify: `go.mod` - 添加 goquery, sqlx 依赖
- Create: `services/scraper_service.go` - 网页抓取服务
- Create: `services/source_service.go` - 书源管理服务
- Create: `services/cleanup_service.go` - 内容净化服务

**Step 1: 更新 go.mod**

```go
module inkread

go 1.21

require (
	github.com/PuerkitoBio/goquery v1.9.2
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/mattn/go-sqlite3 v1.14.22
)
```

**Step 2: 运行 go mod tidy**

Run: `cd /mnt/e/code/inkread && go mod tidy`
Expected: 无错误，生成 go.sum

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: 添加 goquery 和 sqlx 依赖"
```

---

### Task 2: 创建数据库迁移 - 书源表

**Files:**
- Create: `storage/migrations/001_book_sources.sql`
- Modify: `storage/sqlite.go` - 添加书源相关方法

**Step 1: 创建迁移 SQL**

```sql
-- 书源表
CREATE TABLE IF NOT EXISTS book_sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url_template TEXT NOT NULL,
    encoding TEXT DEFAULT 'utf-8',
    book_name_rule TEXT,
    author_rule TEXT,
    content_rule TEXT,
    chapter_list_rule TEXT,
    chapter_url_rule TEXT,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 内容净化规则表
CREATE TABLE IF NOT EXISTS cleanup_rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    pattern TEXT NOT NULL,
    replacement TEXT DEFAULT '',
    rule_type TEXT DEFAULT 'replace',
    enabled INTEGER DEFAULT 1,
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Step 2: 更新 sqlite.go 添加初始化**

```go
func (s *SQLiteStore) InitSchema() error {
    schema := `
    CREATE TABLE IF NOT EXISTS book_sources (...);
    CREATE TABLE IF NOT EXISTS cleanup_rules (...);
    `
    _, err := s.db.Exec(schema)
    return err
}
```

**Step 3: 测试**

Run: `go test ./storage/... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add storage/sqlite.go storage/migrations/
git commit -m "feat: 添加书源和净化规则数据库表"
```

---

## Phase 2: 书源管理功能

### Task 3: 创建 BookSource 模型和 API

**Files:**
- Create: `models/source.go`
- Modify: `api/handlers.go` - 添加书源 CRUD handler
- Create: `api/source_handlers.go`

**Step 1: 创建 models/source.go**

```go
package models

type BookSource struct {
    ID            string    `json:"id" db:"id"`
    Name          string    `json:"name" db:"name"`
    URLTemplate   string    `json:"url_template" db:"url_template"`
    Encoding      string    `json:"encoding" db:"encoding"`
    BookNameRule  string    `json:"book_name_rule" db:"book_name_rule"`
    AuthorRule    string    `json:"author_rule" db:"author_rule"`
    ContentRule   string    `json:"content_rule" db:"content_rule"`
    ChapterRule   string    `json:"chapter_list_rule" db:"chapter_list_rule"`
    ChapterURLRule string   `json:"chapter_url_rule" db:"chapter_url_rule"`
    Enabled       bool      `json:"enabled" db:"enabled"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type CleanupRule struct {
    ID          string    `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Pattern     string    `json:"pattern" db:"pattern"`
    Replacement string    `json:"replacement" db:"replacement"`
    RuleType    string    `json:"rule_type" db:"rule_type"`
    Enabled     bool      `json:"enabled" db:"enabled"`
    Priority    int       `json:"priority" db:"priority"`
}
```

**Step 2: 添加 API handlers**

在 handlers.go 中添加:
```go
func (h *Handlers) ListSources(c *gin.Context) {...}
func (h *Handlers) CreateSource(c *gin.Context) {...}
func (h *Handlers) UpdateSource(c *gin.Context) {...}
func (h *Handlers) DeleteSource(c *gin.Context) {...}
func (h *Handlers) TestSource(c *gin.Context) {...}
```

**Step 3: 测试**

Run: `go test ./api/... -v -run Source`
Expected: PASS

**Step 4: Commit**

```bash
git add models/source.go api/handlers.go
git commit -m "feat: 添加书源管理 CRUD API"
```

---

### Task 4: 实现书源解析引擎

**Files:**
- Modify: `services/source_service.go`

**Step 1: 实现 SourceParser**

```go
type SourceParser struct {
    source *models.BookSource
    doc    *goquery.Document
}

func (p *SourceParser) ExtractBookInfo(html string) (*BookInfo, error) {
    // 使用 CSS Selector 规则提取书名、作者
}

func (p *SourceParser) ExtractChapters(html string) ([]Chapter, error) {
    // 提取章节列表
}

func (p *SourceParser) ExtractContent(html string) (string, error) {
    // 提取正文内容
}
```

**Step 2: Commit**

```bash
git add services/source_service.go
git commit -m "feat: 实现书源解析引擎"
```

---

## Phase 3: Web URL 导入

### Task 5: 创建 Web 导入 API

**Files:**
- Create: `services/web_import_service.go`
- Modify: `api/handlers.go` - 添加 /api/import/url

**Step 1: 实现 WebImportService**

```go
type WebImportService struct {
    sourceService *SourceService
    cleanupService *CleanupService
    store *storage.SQLiteStore
}

func (s *WebImportService) ImportFromURL(url string, sourceID string) (*Book, error) {
    // 1. 获取书源
    // 2. 抓取网页
    // 3. 解析内容
    // 4. 净化处理
    // 5. 保存到数据库
}
```

**Step 2: 添加 API 路由**

```go
api.POST("/import/url", h.ImportFromURL)
```

**Step 3: 测试**

Run: `go test ./services/... -v -run WebImport`

**Step 4: Commit**

```bash
git add services/web_import_service.go api/handlers.go
git commit -m "feat: 添加 Web URL 导入功能"
```

---

## Phase 4: 内容净化

### Task 6: 实现内容净化管道

**Files:**
- Modify: `services/cleanup_service.go`

**Step 1: 实现 CleanupPipeline**

```go
type CleanupPipeline struct {
    rules []CleanupRule
}

func (p *CleanupPipeline) Execute(content string) string {
    result := content
    for _, rule := range p.rules {
        result = rule.Apply(result)
    }
    return result
}

func (r *CleanupRule) Apply(content string) string {
    switch r.RuleType {
    case "replace":
        return regexp.MustCompile(r.Pattern).ReplaceAllString(content, r.Replacement)
    case "remove":
        return regexp.MustCompile(r.Pattern).ReplaceAllString(content, "")
    }
    return content
}
```

**Step 2: 预设净化规则**

```go
var DefaultCleanupRules = []CleanupRule{
    {Name: "remove_script", Pattern: `<script[^>]*>.*?</script>`, RuleType: "remove"},
    {Name: "remove_style", Pattern: `<style[^>]*>.*?</style>`, RuleType: "remove"},
    {Name: "remove_nav", Pattern: `<nav[^>]*>.*?</nav>`, RuleType: "remove"},
    {Name: "remove_footer", Pattern: `<footer[^>]*>.*?</footer>`, RuleType: "remove"},
    {Name: "remove_comments", Pattern: `<!--.*?-->`, RuleType: "remove"},
    {Name: "remove_ads", Pattern: `<[^>]+class="[^"]*ad[^"]*"[^>]*>.*?</[^>]+>`, RuleType: "remove"},
}
```

**Step 3: 测试**

Run: `go test ./services/... -v -run Cleanup`

**Step 4: Commit**

```bash
git add services/cleanup_service.go
git commit -m "feat: 实现内容净化管道"
```

---

## Phase 5: TXT 增强

### Task 7: 增强 TXT 阅读支持

**Files:**
- Modify: `services/book_service.go` - 添加 TXT 分章
- Create: `services/txt_parser.go`

**Step 1: 实现 TXTParser**

```go
type TXTParser struct{}

func (p *TXTParser) Parse(content []byte) (string, error) {
    // 自动检测编码
    encoding := detectEncoding(content)
    // 转换编码
    reader := transform.NewReader(bytes.NewReader(content), encoding.NewDecoder())
    text, _ := io.ReadAll(reader)
    return string(text), nil
}

func (p *TXTParser) SplitChapters(content string) []string {
    // 按章节分割
    // 常见模式: 第X章、第X回、第X节
    chapters := splitByPattern(content, `(第[一二三四五六七八九十百千零\d]+[章回节篇部])\s*`)
    return chapters
}
```

**Step 2: 测试**

Run: `go test ./services/... -v -run TXT`

**Step 3: Commit**

```bash
git add services/txt_parser.go services/book_service.go
git commit -m "feat: 增强 TXT 支持，添加编码检测和自动分章"
```

---

## Phase 6: 前端增强

### Task 8: 阅读器主题和设置

**Files:**
- Modify: `reader.html` - 添加主题切换
- Modify: `css/style.css` - 添加主题样式
- Modify: `js/app.js` - 添加设置保存

**Step 1: 添加主题切换 HTML**

```html
<div id="reader-settings" class="settings-panel">
    <div class="setting-item">
        <label>主题</label>
        <select id="theme-select">
            <option value="light">白天</option>
            <option value="dark">夜间</option>
            <option value="sepia">护眼</option>
        </select>
    </div>
    <div class="setting-item">
        <label>字体大小</label>
        <input type="range" id="font-size" min="12" max="32" value="16">
    </div>
    <div class="setting-item">
        <label>行距</label>
        <input type="range" id="line-height" min="1" max="3" step="0.1" value="1.5">
    </div>
</div>
```

**Step 2: 添加 CSS 主题**

```css
.theme-light { --bg-color: #fff; --text-color: #333; }
.theme-dark { --bg-color: #1a1a1a; --text-color: #ddd; }
.theme-sepia { --bg-color: #f4ecd8; --text-color: #5b4636; }
```

**Step 3: 保存设置到 localStorage**

```javascript
function saveSettings() {
    const settings = {
        theme: document.getElementById('theme-select').value,
        fontSize: document.getElementById('font-size').value,
        lineHeight: document.getElementById('line-height').value
    };
    localStorage.setItem('readerSettings', JSON.stringify(settings));
}
```

**Step 4: 测试**

手动测试: 打开 reader.html，切换主题，调整字体大小

**Step 5: Commit**

```bash
git add reader.html css/style.css js/app.js
git commit -m "feat: 阅读器添加主题切换和自定义设置"
```

---

## Phase 7: 集成测试

### Task 9: 集成测试和修复

**Files:** N/A

**Step 1: 运行所有测试**

Run: `go test ./... -v`
Expected: 全部通过

**Step 2: 手动测试流程**

1. 启动服务: `go run main.go`
2. 添加书源
3. 通过 URL 导入小说
4. 阅读测试
5. 测试主题切换

**Step 3: 修复发现的问题**

根据测试结果修复 bug

**Step 4: Commit**

```bash
git add -A
git commit -m "test: 集成测试和修复"
```

---

## Task 10: 最终验证和发布

**Files:** N/A

**Step 1: 最终测试**

Run: `go test ./... -q && go build`

**Step 2: 更新文档**

Update: `README.md` - 添加新功能说明

**Step 3: 推送 GitHub**

```bash
git push origin master
```

---

## 验证命令

```bash
# 运行所有测试
go test ./... -v

# 编译
go build -o inkread

# 启动服务
./inkread

# API 测试
curl http://localhost:8080/api/sources  # 列出书源
curl http://localhost:8080/api/books     # 列出书籍
```

---

## 风险和注意事项

1. **网络抓取**: 需要处理网站反爬、编码问题、超时
2. **XSS 安全**: 用户输入的 HTML 内容需要净化
3. **大文件**: TXT 文件可能很大，需要流式处理
4. **并发**: 多个用户同时抓取同一个网站可能被封

---

*Plan created: 2026-04-18*
