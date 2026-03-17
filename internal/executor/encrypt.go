package executor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	saltSize   = 32     // 盐值大小,为每个文件生成唯一的随机盐值
	keySize    = 32     // AES - 256,AES 加密算法使用的密钥长度
	iterations = 100000 // 密钥派生函数迭代次数,增加暴力破解的难度
)

// Encryptor AES-256-GCM 加密器
type Encryptor struct {
	password []byte
}

// NewEncryptor 创建加密器
func NewEncryptor(password []byte) (*Encryptor, error) {
	if len(password) == 0 {
		return nil, fmt.Errorf("密码不能为空")
	}
	return &Encryptor{password: password}, nil
}

// EncryptFile 加密文件并写入目标路径
func (e *Encryptor) EncryptFile(src, destDir string) (string, error) {
	// 读取源文件
	plaintext, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("读取源文件失败：%w", err)
	}
	// 生成随机盐值
	salt := make([]byte, saltSize)
	_, err = io.ReadFull(rand.Reader, salt)
	if err != nil {
		return "", fmt.Errorf("生成盐值失败：%w", err)
	}

	// 使用pbkdf2 派生密钥
	key, err := pbkdf2.Key(sha256.New, string(e.password), salt, iterations, keySize)
	if err != nil {
		return "", fmt.Errorf("生成密钥失败：%w", err)
	}

	// 创建 AES-GCM 加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建加密器失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建 GCM 失败: %w", err)
	}

	// 生成随机 Nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成 Nonce 失败: %w", err)
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 构造输出：salt(32) + nonce+ciphertext
	output := append(salt, ciphertext...)

	// 目标路径：destDir/<原文件名>.napsec
	destPath := filepath.Join(destDir, filepath.Base(src)+".napsec")

	// 确保目标目录存在
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return "", fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 写入加密文件（权限 0600）
	if err := os.WriteFile(destPath, output, 0600); err != nil {
		return "", fmt.Errorf("写入加密文件失败: %w", err)
	}

	return destPath, nil
}

// DecryptFile 解密文件到目标路径
func (e *Encryptor) DecryptFile(src, destPath string) error {
	// 读取加密文件
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("读取加密文件失败: %w", err)
	}

	if len(data) < saltSize {
		return fmt.Errorf("文件格式无效")
	}

	// 提取盐
	salt := data[:saltSize]
	ciphertext := data[saltSize:]

	// 派生密钥
	key, err := pbkdf2.Key(sha256.New, string(e.password), salt, iterations, keySize)
	if err != nil {
		return fmt.Errorf("生成密钥失败：%w", err)
	}

	// 创建解密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("创建解密器失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("创建 GCM 失败: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("密文太短")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("解密失败（密码错误？）: %w", err)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 写入解密文件
	return os.WriteFile(destPath, plaintext, 0644)
}
