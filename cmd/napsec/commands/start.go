package commands

import (
	"fmt"
	"github.com/liangach/napsec/internal/ai"
	"github.com/liangach/napsec/internal/config"
	"github.com/liangach/napsec/internal/core"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var startCmd = &cobra.Command{
	Use:   "start [监控目录]",
	Short: "启动 NapSec 监控服务",
	Long: `启动 NapSec 文件监控服务，实时监测并保护敏感文件 

示例:
  napsec start ~/Documents
  napsec start ~/Desktop --password password
  napsec start . --dry-run
  
  # 启用AI判断
  napsec start ~/Documents --ai-enabled --ai-provider openai --ai-key sk-xxx
  
  # 使用Azure OpenAI
  napsec start ~/Documents --ai-enabled --ai-provider azure --ai-endpoint https://xxx.openai.azure.com --ai-key xxx
  
  # 使用本地Ollama
  napsec start ~/Documents --ai-enabled --ai-provider ollama --ai-endpoint http://localhost:11434/api/generate --ai-model llama2
  
  # 使用DeepSeek
  napsec start ~/Documents --ai-enabled --ai-provider deepseek --ai-key sk-xxx`,
	Args: cobra.MaximumNArgs(1),
	RunE: RunStart,
}

func init() {
	startCmd.Flags().StringP("password", "p", "", "加密密码")
	startCmd.Flags().StringP("vault", "v", "", "加密保险箱路径（默认：~/.napsec/vault）")
	startCmd.Flags().BoolP("dry-run", "d", false, "演习模式，只检测不执行保护操作")
	startCmd.Flags().IntP("workers", "w", 4, "并发线程数")
	// AI相关标志
	startCmd.Flags().Bool("ai-enabled", false, "启用AI大模型判断")
	startCmd.Flags().String("ai-provider", "openai", "AI提供商 (openai, azure, anthropic, ollama, gemini, deepseek)")
	startCmd.Flags().String("ai-endpoint", "", "AI API地址（可选）")
	startCmd.Flags().String("ai-key", "", "AI API密钥")
	startCmd.Flags().String("ai-model", "", "AI模型名称（可选）")
	startCmd.Flags().Int("ai-max-tokens", 500, "AI最大token数")
	startCmd.Flags().Int("ai-timeout", 30, "AI超时时间(秒)")
	startCmd.Flags().Int("ai-sample-lines", 50, "AI采样行数")
	startCmd.Flags().Int64("ai-max-file-size", 1024*1024, "AI处理的最大文件大小(bytes)")
	startCmd.Flags().Float64("ai-temperature", 0.3, "AI温度参数(0-1)")
}

func RunStart(cmd *cobra.Command, args []string) error {
	// 解析监控目录
	watchDir := "."
	if len(args) > 0 {
		watchDir = args[0]
	}
	absWatchDir, err := filepath.Abs(watchDir)
	if err != nil {
		return fmt.Errorf("无法解析路径: %w", err)
	}

	// 检查目录是否存在
	_, err = os.Stat(absWatchDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("目录不存在: %s", absWatchDir)
	}

	// 获取密码
	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		password, err = promptPasswordWithConfirm("请输入加密密码：")
		if err != nil {
			return fmt.Errorf("密码输入失败：%w", err)
		}
	}

	// 获取其他参数
	vaultPath, _ := cmd.Flags().GetString("vault")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	workers, _ := cmd.Flags().GetInt("workers")

	// 获取AI参数
	aiEnabled, _ := cmd.Flags().GetBool("ai-enabled")
	aiProvider, _ := cmd.Flags().GetString("ai-provider")
	aiEndpoint, _ := cmd.Flags().GetString("ai-endpoint")
	aiKey, _ := cmd.Flags().GetString("ai-key")
	aiModel, _ := cmd.Flags().GetString("ai-model")
	aiMaxTokens, _ := cmd.Flags().GetInt("ai-max-tokens")
	aiTimeout, _ := cmd.Flags().GetInt("ai-timeout")
	aiSampleLines, _ := cmd.Flags().GetInt("ai-sample-lines")
	aiMaxFileSize, _ := cmd.Flags().GetInt64("ai-max-file-size")
	aiTemperature, _ := cmd.Flags().GetFloat64("ai-temperature")

	// 获取用户主目录
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %w", err)
	}

	// 基础目录
	baseDir := filepath.Join(home, ".napsec")

	// 设置保险箱路径
	if vaultPath == "" {
		vaultPath = filepath.Join(baseDir, "vault")
	}

	// 审计目录路径（必须设置！）
	auditDir := filepath.Join(baseDir, "audit")

	// 创建所有必要的目录
	fmt.Println("正在初始化目录...")
	dirsToCreate := []string{
		baseDir,   // ~/.napsec
		vaultPath, // ~/.napsec/vault
		auditDir,  // ~/.napsec/audit
	}

	for _, dir := range dirsToCreate {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
		fmt.Printf("  ✓ %s\n", dir)
	}

	// 构建配置 - 确保所有字段都有值！
	cfg := &config.Config{
		WatchDir:  absWatchDir,
		Workers:   workers,
		DryRun:    dryRun,
		VaultPath: vaultPath,
		AuditDir:  auditDir, // 这里一定要赋值！
		Password:  password,
		WebPort:   8080, // 设置默认值
		AI: ai.AIConfig{
			Enabled:       aiEnabled,
			Provider:      ai.ProviderType(aiProvider),
			Endpoint:      aiEndpoint,
			APIKey:        aiKey,
			Model:         aiModel,
			MaxTokens:     aiMaxTokens,
			Timeout:       aiTimeout,
			SampleLines:   aiSampleLines,
			MaxFileSize:   aiMaxFileSize,
			Temperature:   aiTemperature,
			CustomHeaders: make(map[string]string),
		},
	}

	// 如果未指定模型，根据提供商设置默认值
	if aiEnabled && cfg.AI.Model == "" {
		switch cfg.AI.Provider {
		case "openai":
			cfg.AI.Model = "gpt-3.5-turbo"
		case "azure":
			cfg.AI.Model = "gpt-35-turbo"
		case "anthropic":
			cfg.AI.Model = "claude-3-haiku-20240307"
		case "ollama":
			cfg.AI.Model = "llama2"
		case "gemini":
			cfg.AI.Model = "gemini-pro"
		case "deepseek":
			cfg.AI.Model = "deepseek-chat"
		}
	}

	// 调试输出
	fmt.Printf("\n配置信息:\n")
	fmt.Printf("  WatchDir: %s\n", cfg.WatchDir)
	fmt.Printf("  VaultPath: %s\n", cfg.VaultPath)
	fmt.Printf("  AuditDir: %s\n", cfg.AuditDir)
	fmt.Printf("  Workers: %d\n", cfg.Workers)
	fmt.Printf("  DryRun: %v\n", cfg.DryRun)
	fmt.Printf("  AI Enabled: %v\n", cfg.AI.Enabled)
	if cfg.AI.Enabled {
		fmt.Printf("  AI Provider: %s\n", cfg.AI.Provider)
		fmt.Printf("  AI Model: %s\n", cfg.AI.Model)
	}

	// 初始化引擎
	engine, err := core.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("引擎初始化失败: %w", err)
	}

	// 启动引擎
	fmt.Printf("\nNapSec 已启动\n")
	fmt.Printf("监控目录: %s\n", absWatchDir)
	fmt.Printf("保险箱:   %s\n", vaultPath)
	fmt.Printf("审计目录: %s\n", auditDir)
	if dryRun {
		fmt.Printf("演习模式: 不执行实际保护操作\n")
	}
	fmt.Printf("按 Ctrl+C 停止监控\n\n")

	if err := engine.Start(); err != nil {
		return fmt.Errorf("引擎启动失败: %w", err)
	}

	// 等待信号中断
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Printf("\nNapSec 正在停止...\n")
	return engine.Stop()
}

func promptPasswordWithConfirm(prompt string) (string, error) {
	// 第一次输入
	password, err := promptPassword(prompt)
	if err != nil {
		return "", err
	}

	// 第二次确认
	confirm, err := promptPassword("请再次输入密码: ")
	if err != nil {
		return "", err
	}

	// 比较两次输入的密码是否一致
	if password != confirm {
		return "", fmt.Errorf("两次输入的密码不一致")
	}

	return password, nil
}

func promptPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("读取密码失败: %w", err)
	}
	fmt.Println()
	return string(bytePassword), nil
}
