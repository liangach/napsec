package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/liangach/napsec/internal/ai/provider"
	"github.com/liangach/napsec/internal/ai/provider/anthropic"
	"github.com/liangach/napsec/internal/ai/provider/azure"
	"github.com/liangach/napsec/internal/ai/provider/deepseek"
	"github.com/liangach/napsec/internal/ai/provider/gemini"
	"github.com/liangach/napsec/internal/ai/provider/ollama"
	"github.com/liangach/napsec/internal/ai/provider/openai"
)

// Client AI客户端
type Client struct {
	provider provider.Provider
	sampler  *Sampler
	config   *AIConfig
}

// NewClient 创建AI客户端
func NewClient(cfg *AIConfig) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	if !cfg.Enabled {
		return nil, nil
	}

	// 创建提供商配置
	providerCfg := &provider.Config{
		Endpoint:    cfg.Endpoint,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		Timeout:     cfg.Timeout,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		Headers:     cfg.CustomHeaders,
	}

	// 根据提供商类型创建对应的provider
	var prov provider.Provider
	var err error

	switch cfg.Provider {
	case OpenAI:
		prov, err = openai.NewProvider(providerCfg)
	case Azure:
		prov, err = azure.NewProvider(providerCfg)
	case Anthropic:
		prov, err = anthropic.NewProvider(providerCfg)
	case Ollama:
		prov, err = ollama.NewProvider(providerCfg)
	case Gemini:
		prov, err = gemini.NewProvider(providerCfg)
	case DeepSeek:
		prov, err = deepseek.NewProvider(providerCfg)
	default:
		return nil, fmt.Errorf("不支持的AI提供商: %s", cfg.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("初始化AI提供商失败: %w", err)
	}

	return &Client{
		provider: prov,
		sampler:  NewSampler(cfg.SampleLines, cfg.MaxFileSize),
		config:   cfg,
	}, nil
}

// Detect 检测文件是否包含敏感信息
func (c *Client) Detect(ctx context.Context, fileName string, content []byte) (*provider.Response, error) {
	if c == nil || !c.config.Enabled {
		return nil, fmt.Errorf("AI客户端未启用")
	}

	// 1. 采样内容
	sampled, err := c.sampler.Sample(content, fileName)
	if err != nil {
		return nil, fmt.Errorf("内容采样失败: %w", err)
	}

	// 2. 构建请求
	req := &provider.Request{
		FileName:    fileName,
		Content:     sampled,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	// 3. 调用提供商API
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.config.Timeout)*time.Second)
	defer cancel()

	resp, err := c.provider.Detect(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AI检测失败: %w", err)
	}

	return resp, nil
}

// GetProvider 获取提供商名称
func (c *Client) GetProvider() string {
	if c == nil || c.provider == nil {
		return ""
	}
	return c.provider.Name()
}

// GetModel 获取模型名称
func (c *Client) GetModel() string {
	if c == nil || c.provider == nil {
		return ""
	}
	return c.provider.GetModel()
}

// IsEnabled 检查AI是否启用
func (c *Client) IsEnabled() bool {
	return c != nil && c.config != nil && c.config.Enabled
}
