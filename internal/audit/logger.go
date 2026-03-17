package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// OperationType 操作类型
type OperationType string

const (
	OpEncrypt OperationType = "ENCRYPT"
	OpDecrypt OperationType = "DECRYPT"
	OpDetect  OperationType = "DETECT"
	OpRecover OperationType = "RECOVER"
)

// LogRecord 审计日志记录
type LogRecord struct {
	ID            string        `json:"id"`
	Timestamp     time.Time     `json:"timestamp"`
	Operation     OperationType `json:"operation"`
	OriginalPath  string        `json:"original_path"`
	EncryptedPath string        `json:"encrypted_path,omitempty"`
	RuleName      string        `json:"rule_name,omitempty"`
	Severity      int           `json:"severity,omitempty"`
	Success       bool          `json:"success"`
	ErrorMsg      string        `json:"error,omitempty"`
}

// Stats 统计信息
type Stats struct {
	TotalProtected int
	TodayProtected int
	TotalLogs      int
	LastOperation  time.Time
}

// Logger 审计日志管理器
type Logger struct {
	logDir  string
	gitLog  *GitLog
	records []LogRecord // 内存缓存
}

// NewLogger 创建日志管理器
func NewLogger(logDir string) (*Logger, error) {
	// 检查路径是否为空
	if logDir == "" {
		return nil, fmt.Errorf("日志目录路径为空")
	}

	// 清理路径（处理 Windows 反斜杠）
	logDir = filepath.Clean(logDir)

	// 确保父目录存在
	parentDir := filepath.Dir(logDir)
	if parentDir != "" && parentDir != "." && parentDir != logDir {
		if err := os.MkdirAll(parentDir, 0700); err != nil {
			return nil, fmt.Errorf("创建日志父目录失败 %s: %w", parentDir, err)
		}
	}

	// 创建日志目录
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("创建日志目录失败 %s: %w", logDir, err)
	}

	// 验证目录是否可写
	testFile := filepath.Join(logDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return nil, fmt.Errorf("日志目录不可写 %s: %w", logDir, err)
	}
	os.Remove(testFile) // 清理测试文件

	// 初始化 Git 日志
	gitLog, err := NewGitLog(logDir)
	if err != nil {
		fmt.Printf("Git 日志初始化失败（不影响运行）: %v\n", err)
	}

	l := &Logger{
		logDir: logDir,
		gitLog: gitLog,
	}

	// 加载历史记录
	_ = l.loadRecords()

	return l, nil
}

// Log 记录操作
func (l *Logger) Log(record LogRecord) error {
	if record.ID == "" {
		record.ID = generateID()
	}
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	// 追加到内存
	l.records = append(l.records, record)

	// 持久化到 JSON 文件
	if err := l.appendToFile(record); err != nil {
		return err
	}

	// Git commit（异步）
	if l.gitLog != nil {
		go l.gitLog.Commit(
			fmt.Sprintf("[%s] %s", record.Operation, filepath.Base(record.OriginalPath)),
		)
	}

	return nil
}

// GetRecords 获取最近 N 条记录
func (l *Logger) GetRecords(limit int) ([]LogRecord, error) {
	records := make([]LogRecord, len(l.records))
	copy(records, l.records)

	// 按时间倒序
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	if limit > 0 && len(records) > limit {
		return records[:limit], nil
	}
	return records, nil
}

// GetStats 获取统计信息
func (l *Logger) GetStats() (*Stats, error) {
	stats := &Stats{
		TotalLogs: len(l.records),
	}

	today := time.Now().Truncate(24 * time.Hour)

	for _, r := range l.records {
		if r.Operation == OpEncrypt && r.Success {
			stats.TotalProtected++
			if r.Timestamp.After(today) {
				stats.TodayProtected++
			}
		}
		if r.Timestamp.After(stats.LastOperation) {
			stats.LastOperation = r.Timestamp
		}
	}

	return stats, nil
}

// appendToFile 追加记录到日志文件
func (l *Logger) appendToFile(record LogRecord) error {
	logFile := filepath.Join(l.logDir,
		fmt.Sprintf("audit_%s.jsonl",
			time.Now().Format("2006-01-02")))

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// loadRecords 从文件加载历史记录
func (l *Logger) loadRecords() error {
	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		filePath := filepath.Join(l.logDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var record LogRecord
			if err := json.Unmarshal([]byte(line), &record); err == nil {
				l.records = append(l.records, record)
			}
		}
	}

	return nil
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
