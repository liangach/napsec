package config

// AIConfig AI配置
type AIConfig struct {
	Enabled       bool              `yaml:"enabled"`        // 是否启用AI判断
	Provider      string            `yaml:"provider"`       // 提供商: openai, azure, anthropic, ollama, gemini, deepseek
	Endpoint      string            `yaml:"endpoint"`       // API地址
	APIKey        string            `yaml:"api_key"`        // API密钥
	Model         string            `yaml:"model"`          // 模型名称
	MaxTokens     int               `yaml:"max_tokens"`     // 最大token数
	Timeout       int               `yaml:"timeout"`        // 超时时间(秒)
	SampleLines   int               `yaml:"sample_lines"`   // 采样行数
	MaxFileSize   int64             `yaml:"max_file_size"`  // 最大文件大小(bytes)
	Temperature   float64           `yaml:"temperature"`    // 温度参数
	CustomHeaders map[string]string `yaml:"custom_headers"` // 自定义请求头
}
