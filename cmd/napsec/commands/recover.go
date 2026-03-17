package commands

import (
	"fmt"
	"github.com/liangach/napsec/internal/executor"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var recoverCmd = &cobra.Command{
	Use:   "recover [加密文件路径]",
	Short: "从保险箱恢复文件",
	Args:  cobra.ExactArgs(1),
	RunE:  runRecover,
}

func init() {
	recoverCmd.Flags().StringP("output", "o", "", "恢复到指定路径")
	recoverCmd.Flags().StringP("password", "p", "", "解密密码")
}

func runRecover(cmd *cobra.Command, args []string) error {
	encryptedPath := args[0]

	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		var err error
		password, err = promptPassword("请输入解密密码: ")
		if err != nil {
			return err
		}
	}

	outputPath, _ := cmd.Flags().GetString("output")
	if outputPath == "" {
		home, _ := os.UserHomeDir()
		outputPath = filepath.Join(home, "Desktop", "recovered")
	}

	enc, err := executor.NewEncryptor([]byte(password))
	if err != nil {
		return fmt.Errorf("初始化解密器失败: %w", err)
	}

	if err := enc.DecryptFile(encryptedPath, outputPath); err != nil {
		return fmt.Errorf("解密失败: %w", err)
	}

	fmt.Printf("文件已恢复到: %s\n", outputPath)
	return nil
}
