package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/liangach/napsec/internal/ai/provider"
)

type DeepSeekProvider struct {
	config *provider.Config
	client *http.Client
}

func NewProvider(cfg *provider.Config) (provider.Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("DeepSeek API密钥不能为空")
	}

	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.deepseek.com/v1/chat/completions"
	}

	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30
	}

	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 500
	}

	if cfg.Temperature == 0 {
		cfg.Temperature = 0.3
	}

	return &DeepSeekProvider{
		config: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}, nil
}

func (p *DeepSeekProvider) Name() string {
	return "deepseek"
}

func (p *DeepSeekProvider) GetModel() string {
	return p.config.Model
}

func (p *DeepSeekProvider) Detect(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	// 构建提示词
	prompt := p.buildPrompt(req.FileName, req.Content)

	// DeepSeek API 格式（兼容OpenAI）
	reqBody := map[string]interface{}{
		"model": p.config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个敏感信息检测助手。请判断用户提供的文件是否包含敏感信息。只返回JSON格式的结果。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": p.config.Temperature,
		"max_tokens":  p.config.MaxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// 添加自定义头
	for k, v := range p.config.Headers {
		httpReq.Header.Set(k, v)
	}

	// 发送请求
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API返回错误 [%d]: %s", resp.StatusCode, string(body))
	}

	// 解析DeepSeek响应
	var deepseekResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("API返回空结果")
	}

	// 解析返回的JSON内容
	content := deepseekResp.Choices[0].Message.Content

	// 尝试提取JSON
	var aiResp struct {
		IsSensitive    bool     `json:"is_sensitive"`
		Confidence     int      `json:"confidence"`
		Category       string   `json:"category"`
		Reason         string   `json:"reason"`
		SensitiveParts []string `json:"sensitive_parts"`
	}

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		jsonStr := content[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &aiResp); err != nil {
			aiResp.IsSensitive = false
			aiResp.Confidence = 0
			aiResp.Reason = "解析响应失败"
		}
	} else {
		aiResp.IsSensitive = false
		aiResp.Confidence = 0
		aiResp.Reason = "无效的响应格式"
	}

	return &provider.Response{
		IsSensitive:    aiResp.IsSensitive,
		Confidence:     aiResp.Confidence,
		Category:       aiResp.Category,
		Reason:         aiResp.Reason,
		SensitiveParts: aiResp.SensitiveParts,
		TokensUsed:     deepseekResp.Usage.TotalTokens,
	}, nil
}

func (p *DeepSeekProvider) buildPrompt(fileName, content string) string {
	return fmt.Sprintf(`请判断以下文件是否包含敏感信息。

文件名: %s

文件内容:
%s

请按以下JSON格式返回:
{
  "is_sensitive": true/false,
  "confidence": 0-100的置信度,
  "category": "敏感信息类别",
  "reason": "判断理由",
  "sensitive_parts": ["敏感内容片段"]
}

只返回JSON，不要有其他文字。`, fileName, content)
}
