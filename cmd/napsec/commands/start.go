package commands

import (
	"fmt"
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
	napsec start . --dry-run`,
	Args: cobra.MaximumNArgs(1),
	RunE: RunStart,
}

func init() {
	startCmd.Flags().StringP("password", "p", "", "加密密码")
	startCmd.Flags().StringP("vault", "v", "", "加密保险箱路径（默认：~/.napsec/vault）")
	startCmd.Flags().BoolP("dry-run", "d", false, "演习模式，只检测不执行保护操作")
	startCmd.Flags().IntP("workers", "w", 4, "并发线程数")
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
	}

	// 调试输出
	fmt.Printf("\n配置信息:\n")
	fmt.Printf("  WatchDir: %s\n", cfg.WatchDir)
	fmt.Printf("  VaultPath: %s\n", cfg.VaultPath)
	fmt.Printf("  AuditDir: %s\n", cfg.AuditDir)
	fmt.Printf("  Workers: %d\n", cfg.Workers)
	fmt.Printf("  DryRun: %v\n", cfg.DryRun)

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
