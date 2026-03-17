package commands

import (
	"fmt"
	"github.com/liangach/napsec/internal/audit"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看 NapSec 运行状态",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".napsec", "audit")

	logger, err := audit.NewLogger(logDir)
	if err != nil {
		fmt.Println("NapSec 状态: 未初始化")
		return nil
	}

	stats, err := logger.GetStats()
	if err != nil {
		return err
	}

	fmt.Printf("NapSec 状态报告\n")
	fmt.Printf("═══════════════════════════════\n")
	fmt.Printf("总保护文件数:  %d\n", stats.TotalProtected)
	fmt.Printf("今日保护文件:  %d\n", stats.TodayProtected)
	fmt.Printf("最后操作时间:  %s\n",
		stats.LastOperation.Format(time.DateTime))
	fmt.Printf("审计日志条数:  %d\n", stats.TotalLogs)
	fmt.Printf("═══════════════════════════════\n")

	return nil
}
