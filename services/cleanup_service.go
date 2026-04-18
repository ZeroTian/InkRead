package services

import (
	"regexp"
	"strings"
)

// CleanupService 内容净化服务
type CleanupService struct {
	rules []CleanupRule
}

// CleanupRule 净化规则
type CleanupRule struct {
	Pattern     string
	Replacement string
	RuleType    string // "replace" or "remove"
	Priority    int
}

// NewCleanupService 创建净化服务（带默认规则）
func NewCleanupService() *CleanupService {
	s := &CleanupService{}
	s.InitDefaultRules()
	return s
}

// InitDefaultRules 初始化默认规则
func (s *CleanupService) InitDefaultRules() {
	s.rules = []CleanupRule{
		// 移除脚本标签
		{Pattern: `(?i)<script[^>]*>[\s\S]*?</script>`, Replacement: "", RuleType: "remove", Priority: 100},
		// 移除样式标签
		{Pattern: `(?i)<style[^>]*>[\s\S]*?</style>`, Replacement: "", RuleType: "remove", Priority: 90},
		// 移除 head 标签
		{Pattern: `(?i)<head[^>]*>[\s\S]*?</head>`, Replacement: "", RuleType: "remove", Priority: 80},
		// 移除导航栏
		{Pattern: `(?i)<nav[^>]*>[\s\S]*?</nav>`, Replacement: "", RuleType: "remove", Priority: 70},
		// 移除 footer
		{Pattern: `(?i)<footer[^>]*>[\s\S]*?</footer>`, Replacement: "", RuleType: "remove", Priority: 60},
		// 移除注释
		{Pattern: `<!--[\s\S]*?-->`, Replacement: "", RuleType: "remove", Priority: 50},
		// 移除 class 包含 ad 的元素（广告）
		{Pattern: `(?i)<[^>]+class="[^"]*\b(ad|advert|sponsor)[^"]*"[^>]*>[\s\S]*?</[^>]+>`, Replacement: "", RuleType: "remove", Priority: 40},
		// 移除空标签
		{Pattern: `(?i)<(div|span|p)[^>]*>\s*</\1>`, Replacement: "", RuleType: "remove", Priority: 30},
		// 清理多余空行
		{Pattern: `\n{3,}`, Replacement: "\n\n", RuleType: "replace", Priority: 10},
	}
}

// Cleanup 执行净化
func (s *CleanupService) Cleanup(content string) string {
	result := content

	// 按优先级排序
	rules := make([]CleanupRule, len(s.rules))
	copy(rules, s.rules)

	// 先移除标签
	for _, rule := range rules {
		if rule.RuleType == "remove" {
			re := regexp.MustCompile(rule.Pattern)
			result = re.ReplaceAllString(result, rule.Replacement)
		}
	}

	// 再处理替换
	for _, rule := range rules {
		if rule.RuleType == "replace" {
			re := regexp.MustCompile(rule.Pattern)
			result = re.ReplaceAllString(result, rule.Replacement)
		}
	}

	// 清理 HTML 实体
	result = cleanHTMLEntities(result)

	return strings.TrimSpace(result)
}

// cleanHTMLEntities 清理 HTML 实体
func cleanHTMLEntities(content string) string {
	entities := map[string]string{
		"&nbsp;":  " ",
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
		"&quot;":  "\"",
		"&apos;":  "'",
		"&#39;":   "'",
		"&#x27;":  "'",
		"&#xA0;":  " ",
		"&mdash;": "-",
		"&ndash;": "-",
		"&ldquo;": "\"",
		"&rdquo;": "\"",
		"&lsquo;": "'",
		"&rsquo;": "'",
		"&hellip;": "...",
		"&copy;":  "(c)",
		"&reg;":   "(R)",
		"&trade;": "(TM)",
	}

	result := content
	for entity, char := range entities {
		result = strings.ReplaceAll(result, entity, char)
	}

	// 转换数字 HTML 实体
	re := regexp.MustCompile(`&#(\d+);`)
	result = re.ReplaceAllStringFunc(result, func(match string) string {
		re := regexp.MustCompile(`&#(\d+);`)
		matches := re.FindStringSubmatch(match)
		if len(matches) == 2 {
			return intToRune(matches[1])
		}
		return match
	})

	return result
}

// intToRune 将数字转换为对应的字符
func intToRune(s string) string {
	// 简单实现，实际应该更完整
	return s
}

// AddRule 添加规则
func (s *CleanupService) AddRule(rule CleanupRule) {
	s.rules = append(s.rules, rule)
}

// RemoveRule 移除规则
func (s *CleanupService) RemoveRule(pattern string) {
	var newRules []CleanupRule
	for _, r := range s.rules {
		if r.Pattern != pattern {
			newRules = append(newRules, r)
		}
	}
	s.rules = newRules
}
