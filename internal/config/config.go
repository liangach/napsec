package config

import (
	"fmt"
	"github.com/liangach/napsec/internal/ai"
	"os"
	"path/filepath"
	"runtime"
)

// Config 全局配置
type Config struct {
	// 监控配置
	WatchDir string
	Workers  int
	DryRun   bool

	// 安全配置
	Password  string
	VaultPath string

	// 审计配置
	AuditDir string

	// Web 配置
	WebPort int

	// AI 配置
	AI ai.AIConfig `yaml:"ai"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	// Windows 特殊处理
	if runtime.GOOS == "windows" {
		baseDir := filepath.Join(homeDir, ".napsec")
		return &Config{
			WatchDir:  ".",
			Workers:   4,
			DryRun:    false,
			VaultPath: filepath.Join(baseDir, "vault"),
			AuditDir:  filepath.Join(baseDir, "audit"),
			WebPort:   8080,
			AI: ai.AIConfig{
				Enabled:       false,
				Provider:      "openai",
				Endpoint:      "",
				APIKey:        "",
				Model:         "gpt-3.5-turbo",
				MaxTokens:     500,
				Timeout:       30,
				SampleLines:   50,
				MaxFileSize:   1024 * 1024, // 1MB
				Temperature:   0.3,
				CustomHeaders: make(map[string]string),
			},
		}
	}

	// Unix/Linux/macOS
	return &Config{
		WatchDir:  ".",
		Workers:   4,
		DryRun:    false,
		VaultPath: filepath.Join(homeDir, ".napsec", "vault"),
		AuditDir:  filepath.Join(homeDir, ".napsec", "audit"),
		WebPort:   8080,
	}
}

// EnsureDirs 确保必要目录存在
func (c *Config) EnsureDirs() error {
	// 收集所有需要创建的目录
	dirs := []string{}

	// 只添加非空路径
	if c.VaultPath != "" {
		dirs = append(dirs, c.VaultPath)
	}
	if c.AuditDir != "" {
		dirs = append(dirs, c.AuditDir)
	}

	// 如果没有需要创建的目录，直接返回
	if len(dirs) == 0 {
		return fmt.Errorf("没有指定任何目录路径")
	}

	// 创建所有目录
	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		// 清理路径（处理 Windows 反斜杠）
		dir = filepath.Clean(dir)

		// 确保父目录存在
		parentDir := filepath.Dir(dir)
		if parentDir != "" && parentDir != "." && parentDir != dir {
			if err := os.MkdirAll(parentDir, 0700); err != nil {
				return fmt.Errorf("创建父目录失败 %s: %w", parentDir, err)
			}
		}

		// 创建目标目录
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}

		// 验证目录创建成功
		info, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("验证目录失败 %s: %w", dir, err)
		}

		// Windows 不检查权限位
		if runtime.GOOS != "windows" {
			if info.Mode().Perm() != 0700 {
				os.Chmod(dir, 0700) // 修正权限
			}
		}

		fmt.Printf("目录已确认: %s\n", dir)
	}

	return nil
}
