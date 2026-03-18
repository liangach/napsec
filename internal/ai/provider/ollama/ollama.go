package ollama

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

type OllamaProvider struct {
	config *provider.Config
	client *http.Client
}

func NewProvider(cfg *provider.Config) (provider.Provider, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:11434/api/generate"
	}

	if cfg.Model == "" {
		cfg.Model = "llama2"
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 60 // Ollama本地可能较慢
	}

	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 500
	}

	if cfg.Temperature == 0 {
		cfg.Temperature = 0.3
	}

	return &OllamaProvider{
		config: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}, nil
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

func (p *OllamaProvider) GetModel() string {
	return p.config.Model
}

func (p *OllamaProvider) Detect(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	// 构建提示词
	prompt := p.buildPrompt(req.FileName, req.Content)

	// Ollama API 格式
	reqBody := map[string]interface{}{
		"model":  p.config.Model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": p.config.Temperature,
			"num_predict": p.config.MaxTokens,
		},
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

	// 解析Ollama响应
	var ollamaResp struct {
		Response string `json:"response"`
	}

	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 解析返回的JSON内容
	content := ollamaResp.Response

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

	// Ollama不返回token使用量，估算一下
	tokensUsed := len(strings.Fields(prompt)) + len(strings.Fields(content))

	return &provider.Response{
		IsSensitive:    aiResp.IsSensitive,
		Confidence:     aiResp.Confidence,
		Category:       aiResp.Category,
		Reason:         aiResp.Reason,
		SensitiveParts: aiResp.SensitiveParts,
		TokensUsed:     tokensUsed,
	}, nil
}

func (p *OllamaProvider) buildPrompt(fileName, content string) string {
	return fmt.Sprintf(`请判断以下文件是否包含敏感信息（如API密钥、密码、私钥、身份证号、银行卡号等）。

文件名: %s

文件内容:
%s

请严格按以下JSON格式返回，不要有其他文字：
{
  "is_sensitive": true/false,
  "confidence": 0-100的置信度,
  "category": "敏感信息类别",
  "reason": "判断理由",
  "sensitive_parts": ["具体的敏感内容片段"]
}`, fileName, content)
}
