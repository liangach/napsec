package provider

import (
	"context"
)

// Provider AI提供商接口
type Provider interface {
	// Name 返回提供商名称
	Name() string

	// Detect 检测敏感内容
	Detect(ctx context.Context, req *Request) (*Response, error)

	// GetModel 获取当前使用的模型
	GetModel() string
}

// Request 检测请求
type Request struct {
	FileName    string  `json:"file_name"`
	Content     string  `json:"content"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// Response 检测响应
type Response struct {
	IsSensitive    bool     `json:"is_sensitive"`
	Confidence     int      `json:"confidence"` // 0-100
	Category       string   `json:"category"`
	Reason         string   `json:"reason"`
	SensitiveParts []string `json:"sensitive_parts"`
	TokensUsed     int      `json:"tokens_used"`
}

// Config 提供商通用配置
type Config struct {
	Endpoint    string
	APIKey      string
	Model       string
	Timeout     int
	MaxTokens   int
	Temperature float64
	Headers     map[string]string
}
