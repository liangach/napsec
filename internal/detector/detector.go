package detector

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Match 检测匹配结果
type Match struct {
	Rule     Rule
	FilePath string
	Line     int
	Content  string // 匹配内容（脱敏）
}

// Detector 敏感信息检测器
type Detector struct {
	engine *RegexEngine
	rules  []Rule
}

// NewDetector 创建检测器
func NewDetector() (*Detector, error) {
	engine, err := NewRegexEngine(DefaultRules)
	if err != nil {
		return nil, fmt.Errorf("初始化规则引擎失败：%w", err)
	}
	return &Detector{
		engine: engine,
		rules:  DefaultRules,
	}, nil
}

// ScanFile 扫描文件：先文件名，再内容
func (d *Detector) ScanFile(path string) ([]Match, error) {
	// 1. 先检查文件名（有就直接返回）
	matches := d.checkFileName(path)
	if len(matches) > 0 {
		return matches, nil
	}

	// 2. 再检查文件内容
	contentMatches, err := d.scanFileContent(path)
	if err != nil {
		return nil, err
	}
	return contentMatches, nil
}

// checkFileName 检查敏感文件名
func (d *Detector) checkFileName(path string) []Match {
	name := strings.ToLower(filepath.Base(path))
	suspicious := []string{
		"password", "passwd", "secret", "credentials",
		"private_key", "id_rsa", "id_ed25519",
		".env", "token", "apikey", "api_key",
		"pem", "key", "db", "conf", "config",
	}

	for _, s := range suspicious {
		if strings.Contains(name, s) {
			return []Match{{
				Rule: Rule{
					Name:        "敏感文件名",
					Category:    CategoryConfig,
					MinSeverity: 6,
				},
				FilePath: path,
				Line:     0,
				Content:  "敏感文件名：" + filepath.Base(path),
			}}
		}
	}
	return nil
}

// scanFileContent 扫描内容
func (d *Detector) scanFileContent(path string) ([]Match, error) {
	// 跳过真正的二进制/媒体
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil
	}

	// 跳过超大文件
	if info.Size() > 50*1024*1024 {
		return nil, nil
	}

	// 跳过媒体/压缩包
	skipExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".mp4": true, ".mp3": true, ".zip": true, ".tar": true, ".gz": true,
		".exe": true, ".dll": true, ".bin": true,
	}
	ext := strings.ToLower(filepath.Ext(path))
	if skipExts[ext] {
		return nil, nil
	}

	// 读取整个文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 处理 UTF-8 BOM
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))

	// 转换为字符串并清理可能的非UTF8字符
	text := string(content)
	text = strings.ToValidUTF8(text, "")

	// 按行分割
	lines := strings.Split(text, "\n")

	var matches []Match

	for lineNum, line := range lines {
		// 清理行中的控制字符和乱码
		cleanLine := cleanString(line)

		if cleanLine == "" {
			continue
		}

		//// 调试输出
		//if strings.Contains(path, "test.txt") || strings.Contains(path, "aws") {
		//	fmt.Printf("[调试] 扫描行 %d: %s\n", lineNum+1, cleanLine)
		//}

		lineMatches := d.engine.MatchLine(cleanLine, path)

		for _, m := range lineMatches {
			// fmt.Printf("[调试] 发现匹配! 规则: %s, 匹配内容: %s\n", m.Rule.Name, m.Match)
			matches = append(matches, Match{
				Rule:     m.Rule,
				FilePath: path,
				Line:     lineNum + 1,
				Content:  maskSensitive(cleanLine),
			})
		}
		if len(matches) >= 20 {
			break
		}
	}

	return matches, nil
}

// cleanString 清理字符串中的非打印字符
func cleanString(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1 // 删除控制字符
		}
		return r
	}, s)
}

// maskSensitive 脱敏
func maskSensitive(content string) string {
	if len(content) <= 10 {
		return strings.Repeat("*", len(content))
	}
	return content[:8] + strings.Repeat("*", len(content)-8)
}
