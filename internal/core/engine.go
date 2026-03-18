package core

import (
	"encoding/json"
	"fmt"
	"github.com/liangach/napsec/internal/ai"
	"github.com/liangach/napsec/internal/audit"
	"github.com/liangach/napsec/internal/config"
	"github.com/liangach/napsec/internal/detector"
	"github.com/liangach/napsec/internal/executor"
	"github.com/liangach/napsec/internal/monitor"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Engine 核心引擎
type Engine struct {
	cfg      *config.Config
	watcher  *monitor.Watcher
	detector *detector.Detector
	mover    *executor.Mover
	logger   *audit.Logger
	aiClient *ai.Client

	// 实时统计（供 Web API 使用）
	stats EngineStats
}

// EngineStats 引擎实时统计
type EngineStats struct {
	StartTime      time.Time `json:"start_time"`
	FilesScanned   int       `json:"files_scanned"`
	FilesProtected int       `json:"files_protected"`
	LastEvent      string    `json:"last_event"`
	IsRunning      bool      `json:"is_running"`
	AIEnabled      bool      `json:"ai_enabled"`
	AICalls        int       `json:"ai_calls"`
}

// NewEngine 创建核心引擎
func NewEngine(cfg *config.Config) (*Engine, error) {
	// 验证必要配置
	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证目录配置
	if cfg.AuditDir == "" {
		return nil, fmt.Errorf("审计目录路径未配置")
	}

	if cfg.VaultPath == "" {
		return nil, fmt.Errorf("保险箱目录路径未配置")
	}

	// 初始化目录（二次确认）
	fmt.Printf("正在初始化目录结构...\n")
	fmt.Printf("  保险箱: %s\n", cfg.VaultPath)
	fmt.Printf("  审计: %s\n", cfg.AuditDir)

	if err := cfg.EnsureDirs(); err != nil {
		return nil, fmt.Errorf("初始化目录失败: %w", err)
	}

	// 初始化AI客户端
	var aiClient *ai.Client
	var err error
	if cfg.AI.Enabled {
		fmt.Printf("AI大模型判断已启用 (提供商: %s, 模型: %s)\n",
			cfg.AI.Provider, cfg.AI.Model)

		aiClient, err = ai.NewClient(&cfg.AI)
		if err != nil {
			fmt.Printf("警告: AI客户端初始化失败: %v\n", err)
			fmt.Println("将仅使用规则引擎继续运行")
			aiClient = nil
		}
	} else {
		fmt.Println("AI大模型判断未启用（仅使用规则引擎）")
	}

	// 初始化检测器
	det, err := detector.NewDetector(aiClient)
	if err != nil {
		return nil, fmt.Errorf("初始化检测器失败: %w", err)
	}

	// 初始化日志 - 使用配置中的路径
	logger, err := audit.NewLogger(cfg.AuditDir)
	if err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	engine := &Engine{
		cfg:      cfg,
		detector: det,
		logger:   logger,
		aiClient: aiClient,
		stats: EngineStats{
			AIEnabled: cfg.AI.Enabled && aiClient != nil,
		},
	}

	// 只有提供密码时才初始化 Mover
	if cfg.Password != "" {
		mover, err := executor.NewMover(
			[]byte(cfg.Password),
			cfg.VaultPath,
			cfg.DryRun,
		)
		if err != nil {
			return nil, fmt.Errorf("初始化执行器失败: %w", err)
		}
		engine.mover = mover
	}

	return engine, nil
}

// DefaultConfig 返回默认配置（供 Web 命令使用）
func DefaultConfig() *config.Config {
	return config.DefaultConfig()
}

// Start 启动引擎
func (e *Engine) Start() error {
	// 1. 先执行初始扫描，处理现有文件
	fmt.Println("正在扫描现有文件...")
	err := filepath.WalkDir(e.cfg.WatchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// 跳过无法访问的路径
			return nil
		}
		// 跳过目录本身
		if d.IsDir() {
			// 避免进入保险箱、审计等内部目录（减少无用遍历）
			if isSubPath(path, e.cfg.VaultPath) || isSubPath(path, e.cfg.AuditDir) {
				return filepath.SkipDir
			}
			return nil
		}

		// 获取文件信息
		info, err := d.Info()
		if err != nil {
			return nil
		}

		// 构造一个“创建”事件并交给处理器
		event := monitor.FileEvent{
			Path:      path,
			Type:      monitor.EventCreate, // 模拟创建事件
			Timestamp: time.Now(),
			Size:      info.Size(),
		}
		e.handleFileEvent(event)
		return nil
	})
	if err != nil {
		fmt.Printf("扫描现有文件时出错: %v\n", err)
		// 继续启动监控，不阻断
	}

	// 2. 启动实时监控
	watcher, err := monitor.NewWatcher(e.cfg.WatchDir, e.handleFileEvent)
	if err != nil {
		return err
	}
	e.watcher = watcher

	e.stats.StartTime = time.Now()
	e.stats.IsRunning = true

	return watcher.Start()
}

// Stop 停止引擎
func (e *Engine) Stop() error {
	if e.watcher != nil {
		e.stats.IsRunning = false
		return e.watcher.Stop()
	}
	return nil
}

// handleFileEvent 处理文件事件
func (e *Engine) handleFileEvent(event monitor.FileEvent) {
	fmt.Printf("\n=== 调试信息 ===\n")
	fmt.Printf("收到事件: 类型=%s, 路径=%s\n", event.Type, event.Path)
	fmt.Printf("当前时间: %s\n", time.Now().Format(time.DateTime))

	// 检查是否在保险箱内
	absPath, err := filepath.Abs(event.Path)
	if err != nil {
		fmt.Printf("无法获取绝对路径: %v\n", err)
		return
	}

	absVault, err := filepath.Abs(e.cfg.VaultPath)
	if err != nil {
		fmt.Printf("无法获取保险箱绝对路径: %v\n", err)
		return
	}

	// 检查是否在保险箱内
	if isSubPath(absPath, absVault) {
		fmt.Printf("跳过: 文件在保险箱内 (%s)\n", absVault)
		return
	}

	// 检查是否是替身文件
	if strings.HasSuffix(event.Path, ".napsec_decoy") {
		fmt.Printf("跳过: 是替身文件\n")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(event.Path); os.IsNotExist(err) {
		fmt.Printf("警告: 文件不存在: %s\n", event.Path)
		return
	}

	e.stats.FilesScanned++
	e.stats.LastEvent = filepath.Base(event.Path)

	fmt.Printf("开始扫描文件: %s\n", event.Path)

	// 运行敏感检测
	matches, err := e.detector.ScanFile(event.Path)
	if err != nil {
		fmt.Printf("扫描失败: %v\n", err)
		return
	}

	fmt.Printf("扫描完成，找到 %d 个匹配\n", len(matches))

	if len(matches) == 0 {
		fmt.Printf("文件未发现敏感内容\n")
		return
	}

	// 输出检测结果
	fmt.Printf("\n发现敏感文件!\n")
	fmt.Printf("   路径    : %s\n", event.Path)
	fmt.Printf("   检测时间: %s\n", event.Timestamp.Format(time.DateTime))
	for _, m := range matches {
		fmt.Printf("   规则    : [%s] %s (严重程度: %d)\n",
			m.Rule.Category, m.Rule.Name, m.Rule.MinSeverity)
		fmt.Printf("   内容    : %s\n", m.Content)
	}

	// 记录检测日志
	_ = e.logger.Log(audit.LogRecord{
		Operation:    audit.OpDetect,
		OriginalPath: event.Path,
		RuleName:     matches[0].Rule.Name,
		Severity:     matches[0].Rule.MinSeverity,
		Success:      true,
	})

	// 执行保护操作
	if e.mover != nil {
		result, err := e.mover.ProtectFile(event.Path)
		if err != nil {
			fmt.Printf("保护失败: %v\n", err)
			_ = e.logger.Log(audit.LogRecord{
				Operation:    audit.OpEncrypt,
				OriginalPath: event.Path,
				Success:      false,
				ErrorMsg:     err.Error(),
			})
			return
		}

		e.stats.FilesProtected++
		fmt.Printf("文件已加密保护\n")
		fmt.Printf("     加密路径: %s\n", result.EncryptedPath)
		if result.DecoyPath != "" {
			fmt.Printf("     替身文件: %s\n", result.DecoyPath)
		}

		// 记录加密日志
		_ = e.logger.Log(audit.LogRecord{
			Operation:     audit.OpEncrypt,
			OriginalPath:  event.Path,
			EncryptedPath: result.EncryptedPath,
			Success:       true,
		})
	}
}

// RegisterRoutes 注册 HTTP 路由（供 Web 仪表盘使用）
func (e *Engine) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/stats", e.handleAPIStats)
	mux.HandleFunc("/api/records", e.handleAPIRecords)
	mux.HandleFunc("/api/health", e.handleAPIHealth)
}

// handleAPIStats 返回实时统计
func (e *Engine) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	stats, err := e.logger.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"engine":  e.stats,
		"history": stats,
	}
	writeJSON(w, resp)
}

// handleAPIRecords 返回审计记录列表
func (e *Engine) handleAPIRecords(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	records, err := e.logger.GetRecords(50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, records)
}

// handleAPIHealth 健康检查
func (e *Engine) handleAPIHealth(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	writeJSON(w, map[string]interface{}{
		"status":    "ok",
		"running":   e.stats.IsRunning,
		"timestamp": time.Now(),
	})
}

// 工具函数
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func isSubPath(path, parent string) bool {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	return len(rel) > 0 && rel[0] != '.'
}
