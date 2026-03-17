package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "napsec",
	Short: "NapSec - 隐私数据管家",
	Long: `NapSec 是一个本地隐私数据保护工具。
它实时监控文件系统，自动检测并保护敏感数据。

功能特性:
  • 实时文件监控
  • 敏感数据自动检测
  • AES-256 加密保护
  • Git 审计日志
  • Web 仪表盘`,
	Version: "0.1.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(recoverCmd)
	rootCmd.AddCommand(webCmd)

	// 全局标志
	rootCmd.PersistentFlags().StringP(
		"config", "c", "",
		"配置文件路径 (默认: ~/.napsec/config.yaml)",
	)

	// 自定义 help 模板
	rootCmd.SetHelpTemplate(helpTemplate)
}

var helpTemplate = `{{.Long}}

用法:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [命令]{{end}}{{if gt (len .Aliases) 0}}

别名:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

示例:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

可用命令:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

标志:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

全局标志:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

使用 "{{.CommandPath}} [命令] --help" 获取命令帮助。
`

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
