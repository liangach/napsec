package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	AllowedReadDirs  []string
	AllowedWriteDirs []string
}

// Sandbox 执行沙箱（限制操作权限）
type Sandbox struct {
	config SandboxConfig
}

// NewSandbox 创建沙箱
func NewSandbox(vaultPath string) *Sandbox {
	home, _ := os.UserHomeDir()
	return &Sandbox{
		config: SandboxConfig{
			AllowedReadDirs: []string{
				home,
			},
			AllowedWriteDirs: []string{
				vaultPath,
				filepath.Join(home, ".guardian"),
			},
		},
	}
}

// ValidateRead 验证读取权限
func (s *Sandbox) ValidateRead(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for _, allowed := range s.config.AllowedReadDirs {
		allowedAbs, _ := filepath.Abs(allowed)
		if isSubPath(abs, allowedAbs) {
			return nil
		}
	}

	return fmt.Errorf("沙箱拒绝: 不允许读取路径 %s", path)
}

// ValidateWrite 验证写入权限
func (s *Sandbox) ValidateWrite(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for _, allowed := range s.config.AllowedWriteDirs {
		allowedAbs, _ := filepath.Abs(allowed)
		if isSubPath(abs, allowedAbs) {
			return nil
		}
	}

	return fmt.Errorf("沙箱拒绝: 不允许写入路径 %s", path)
}

// PlatformInfo 返回当前平台信息
func PlatformInfo() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

func isSubPath(path, parent string) bool {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	return len(rel) > 0 && rel[0] != '.'
}
