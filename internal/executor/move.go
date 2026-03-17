package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MoveResult 文件移动结果
type MoveResult struct {
	OriginalPath  string
	EncryptedPath string
	DecoyPath     string
	Timestamp     time.Time
}

// Mover 文件移动器（加密 + 创建替身）
type Mover struct {
	encryptor *Encryptor
	vaultPath string
	dryRun    bool
}

// NewMover 创建文件移动器
func NewMover(password []byte, vaultPath string, dryRun bool) (*Mover, error) {
	enc, err := NewEncryptor(password)
	if err != nil {
		return nil, err
	}

	return &Mover{
		encryptor: enc,
		vaultPath: vaultPath,
		dryRun:    dryRun,
	}, nil
}

// ProtectFile 保护文件：加密移入保险箱 + 创建替身
func (m *Mover) ProtectFile(path string) (*MoveResult, error) {
	if m.dryRun {
		fmt.Printf("  [演习] 将保护文件: %s\n", path)
		simulatedPath := filepath.Join(m.vaultPath, filepath.Base(path)+".napsec")
		return &MoveResult{
			OriginalPath:  path,
			EncryptedPath: simulatedPath,
			Timestamp:     time.Now(),
		}, nil
	}

	// 1. 加密文件到保险箱
	encryptedPath, err := m.encryptor.EncryptFile(path, m.vaultPath)
	if err != nil {
		return nil, fmt.Errorf("加密文件失败: %w", err)
	}

	// 验证加密文件确实在保险箱内
	if !strings.HasPrefix(encryptedPath, m.vaultPath) {
		return nil, fmt.Errorf("加密文件路径异常: 不在保险箱内")
	}

	// 2. 创建替身文件（在原位置）
	decoyPath, err := m.createDecoy(path, encryptedPath)
	if err != nil {
		// 替身创建失败不影响主流程
		fmt.Printf("创建替身失败: %v\n", err)
	}

	// 3. 删除原始文件
	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("删除原始文件失败: %w", err)
	}

	return &MoveResult{
		OriginalPath:  path,
		EncryptedPath: encryptedPath,
		DecoyPath:     decoyPath,
		Timestamp:     time.Now(),
	}, nil
}

// createDecoy 在原位置创建替身文件
func (m *Mover) createDecoy(originalPath, encryptedPath string) (string, error) {
	decoyPath := originalPath + ".napsec_decoy"

	content := fmt.Sprintf(`NapSec 隐私保护提示
════════════════════════════════
此文件已被 NapSec 安全保护。

原始文件: %s
加密存储: %s
保护时间: %s

如需恢复，请运行:
  napsec recover "%s"
════════════════════════════════
`,
		filepath.Base(originalPath),
		encryptedPath,
		time.Now().Format(time.DateTime),
		encryptedPath,
	)

	if err := os.WriteFile(decoyPath, []byte(content), 0444); err != nil {
		return "", err
	}

	return decoyPath, nil
}
