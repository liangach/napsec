package ai

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Sampler 内容采样器
type Sampler struct {
	maxLines    int
	maxFileSize int64
}

// NewSampler 创建采样器
func NewSampler(maxLines int, maxFileSize int64) *Sampler {
	if maxLines <= 0 {
		maxLines = 50
	}
	if maxFileSize <= 0 {
		maxFileSize = 1024 * 1024 // 默认1MB
	}
	return &Sampler{
		maxLines:    maxLines,
		maxFileSize: maxFileSize,
	}
}

// Sample 采样文件内容
func (s *Sampler) Sample(content []byte, fileName string) (string, error) {
	// 1. 检查文件大小
	if int64(len(content)) > s.maxFileSize {
		return fmt.Sprintf("[文件太大] %s (大小: %d bytes，超过限制 %d bytes)",
			fileName, len(content), s.maxFileSize), nil
	}

	// 2. 检查是否为文本文件
	if !s.isText(content) {
		// 非文本文件只返回文件名和基本信息
		return fmt.Sprintf("[二进制文件] %s (大小: %d bytes)", fileName, len(content)), nil
	}

	// 3. 转换为UTF-8字符串
	text := string(content)
	if !utf8.Valid(content) {
		text = strings.ToValidUTF8(text, "�")
	}

	// 4. 按行分割
	lines := strings.Split(text, "\n")

	// 5. 如果行数少于等于最大行数，直接返回
	if len(lines) <= s.maxLines {
		return text, nil
	}

	// 6. 智能采样
	return s.smartSample(lines, fileName), nil
}

// smartSample 智能采样
func (s *Sampler) smartSample(lines []string, fileName string) string {
	var result []string

	// 添加文件头信息
	result = append(result, fmt.Sprintf("【文件信息】%s (共 %d 行，显示前%d行采样)",
		fileName, len(lines), s.maxLines))
	result = append(result, "【文件内容采样】")

	// 关键词列表（可能包含敏感信息的行）
	keywords := []string{
		"password", "passwd", "pwd", "secret", "key", "token", "api",
		"private", "ssh", "rsa", "dsa", "ecdsa", "ed25519",
		"credential", "auth", "access", "id", "account",
		"PASSWORD", "PASSWD", "SECRET", "KEY", "TOKEN", "API",
	}

	// 第一遍：找出包含关键词的行
	var keywordLines []int
	for i, line := range lines {
		lowerLine := strings.ToLower(line)
		for _, kw := range keywords {
			if strings.Contains(lowerLine, kw) {
				keywordLines = append(keywordLines, i)
				break
			}
		}
	}

	// 计算采样数量
	keywordCount := len(keywordLines)
	normalCount := s.maxLines - keywordCount

	if normalCount < 5 {
		normalCount = 5
		keywordCount = s.maxLines - 5
	}

	// 添加关键词行
	added := 0
	for i := 0; i < keywordCount && i < len(keywordLines); i++ {
		lineNum := keywordLines[i]
		if lineNum < len(lines) {
			result = append(result, fmt.Sprintf("行 %d: %s", lineNum+1, s.truncateLine(lines[lineNum])))
			added++
		}
	}

	// 如果关键词行不够，补充普通行
	if added < keywordCount {
		// 从头开始采样普通行
		step := len(lines) / (keywordCount - added + 1)
		if step < 1 {
			step = 1
		}
		for i := 0; i < len(lines) && added < keywordCount; i += step {
			// 跳过已经添加的关键词行
			isKeyword := false
			for _, kl := range keywordLines {
				if i == kl {
					isKeyword = true
					break
				}
			}
			if !isKeyword {
				result = append(result, fmt.Sprintf("行 %d: %s", i+1, s.truncateLine(lines[i])))
				added++
			}
		}
	}

	// 添加普通采样行
	if normalCount > 0 {
		result = append(result, "【其他行采样】")
		step := len(lines) / (normalCount + 1)
		if step < 1 {
			step = 1
		}
		added = 0
		for i := 0; i < len(lines) && added < normalCount; i += step {
			// 跳过已经添加的关键词行
			isKeyword := false
			for _, kl := range keywordLines {
				if i == kl {
					isKeyword = true
					break
				}
			}
			if !isKeyword {
				result = append(result, fmt.Sprintf("行 %d: %s", i+1, s.truncateLine(lines[i])))
				added++
			}
		}
	}

	return strings.Join(result, "\n")
}

// truncateLine 截断过长的行
func (s *Sampler) truncateLine(line string) string {
	if len(line) > 200 {
		return line[:197] + "..."
	}
	return line
}

// isText 判断是否为文本文件
func (s *Sampler) isText(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	// 检查前1024个字节
	sample := data
	if len(sample) > 1024 {
		sample = sample[:1024]
	}

	// 如果包含太多控制字符或空字节，可能是二进制文件
	controlChars := 0
	for _, b := range sample {
		if b == 0 || (b < 32 && b != 9 && b != 10 && b != 13) {
			controlChars++
		}
	}

	// 如果超过10%是控制字符，视为二进制文件
	return float64(controlChars)/float64(len(sample)) < 0.1
}

// Compress 压缩内容以适应token限制
func (s *Sampler) Compress(content string, maxTokens int) string {
	// 估算token数（粗略估计：1 token ≈ 4个英文字符）
	estimatedTokens := len(content) / 4
	if estimatedTokens <= maxTokens {
		return content
	}

	// 需要压缩的比例
	ratio := float64(maxTokens) / float64(estimatedTokens)
	targetLen := int(float64(len(content)) * ratio)

	if targetLen < 100 {
		targetLen = 100
	}

	// 截取开头和结尾
	if len(content) <= targetLen {
		return content
	}

	half := targetLen / 2
	start := content[:half]
	end := content[len(content)-half:]

	return start + "\n...[内容压缩]...\n" + end
}
