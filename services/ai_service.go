package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type AIService struct {
	apiKey  string
	model   string
	baseURL string
}

type SummarizeResult struct {
	Summary   string
	Model     string
	CreatedAt time.Time
}

type openAIRequest struct {
	Model    string        `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewAIService(apiKey, model string) *AIService {
	// Support MiniMax (优先) 或 OpenAI 兼容 API
	baseURL := os.Getenv("MINIMAX_BASE_URL")
	if baseURL == "" {
		baseURL = os.Getenv("OPENAI_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "https://api.minimaxi.com/anthropic/v1"
	}
	// 如果没传 apiKey，尝试从环境变量读取
	if apiKey == "" {
		apiKey = os.Getenv("MINIMAX_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &AIService{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
	}
}

func (s *AIService) SummarizeBook(content, bookTitle string) (*SummarizeResult, error) {
	if s.apiKey == "" {
		return s.mockSummarize(content, bookTitle)
	}

	ctx := context.Background()
	return s.callOpenAI(ctx, content, bookTitle)
}

func (s *AIService) callOpenAI(ctx context.Context, content, bookTitle string) (*SummarizeResult, error) {
	truncated := content
	if len(content) > 15000 {
		truncated = content[:15000] + "...(内容已截断)"
	}

	systemPrompt := `你是一个专业的书籍总结助手。请为用户提供书籍的简洁摘要，包括：
1. 书籍主题概述
2. 主要内容或情节
3. 核心观点或价值

请用中文回复，摘要应该简洁有力，不超过500字。`

	userPrompt := fmt.Sprintf("请为《%s》生成书籍摘要：\n\n%s", bookTitle, truncated)

	reqBody := openAIRequest{
		Model: s.model,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}
	defer resp.Body.Close()

	var result openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	return &SummarizeResult{
		Summary:   result.Choices[0].Message.Content,
		Model:     s.model,
		CreatedAt: time.Now(),
	}, nil
}

func (s *AIService) mockSummarize(content, bookTitle string) (*SummarizeResult, error) {
	summary := fmt.Sprintf("【%s】摘要\n\n这是一本精彩的电子书作品。", bookTitle)
	if len(content) > 100 {
		summary += fmt.Sprintf("\n\n内容预览：%s...", content[:100])
	}
	summary += "\n\n（当前为模拟摘要，配置 OPENAI_API_KEY 可获取真实 AI 总结）"

	return &SummarizeResult{
		Summary:   summary,
		Model:     "mock",
		CreatedAt: time.Now(),
	}, nil
}
