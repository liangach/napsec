package ai

import "time"

// ProviderType 提供商类型
type ProviderType string

const (
	OpenAI    ProviderType = "openai"
	Azure     ProviderType = "azure"
	Anthropic ProviderType = "anthropic"
	Ollama    ProviderType = "ollama"
	Gemini    ProviderType = "gemini"
	DeepSeek  ProviderType = "deepseek"
)

// DetectionMode 检测模式
type DetectionMode string

const (
	ModeRuleOnly DetectionMode = "rule-only" // 仅规则
	ModeHybrid   DetectionMode = "hybrid"    // 混合模式
	ModeAIOnly   DetectionMode = "ai-only"   // 仅AI
)

// AIConfig AI配置
type AIConfig struct {
	Enabled       bool              `yaml:"enabled" json:"enabled"`
	Mode          DetectionMode     `yaml:"mode" json:"mode"`
	Provider      ProviderType      `yaml:"provider" json:"provider"`
	Endpoint      string            `yaml:"endpoint" json:"endpoint"`
	APIKey        string            `yaml:"api_key" json:"-"`
	Model         string            `yaml:"model" json:"model"`
	MaxTokens     int               `yaml:"max_tokens" json:"max_tokens"`
	Timeout       int               `yaml:"timeout" json:"timeout"`
	SampleLines   int               `yaml:"sample_lines" json:"sample_lines"`
	MaxFileSize   int64             `yaml:"max_file_size" json:"max_file_size"`
	Temperature   float64           `yaml:"temperature" json:"temperature"`
	CustomHeaders map[string]string `yaml:"custom_headers" json:"custom_headers"`
}

// AIRequest AI请求结构
type AIRequest struct {
	FileName    string  `json:"filename"`
	Content     string  `json:"content"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// AIResponse AI响应结构
type AIResponse struct {
	IsSensitive    bool     `json:"is_sensitive"`
	Confidence     int      `json:"confidence"` // 0-100
	Category       string   `json:"category"`
	Reason         string   `json:"reason"`
	SensitiveParts []string `json:"sensitive_parts"`
	TokensUsed     int      `json:"tokens_used"`
}

// TokenUsage Token使用统计
type TokenUsage struct {
	Daily    int       `json:"daily"`
	LastTime time.Time `json:"last_time"`
}
