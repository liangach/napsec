package commands

import (
	"fmt"
	"github.com/liangach/napsec/internal/audit"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出已保护的文件",
	RunE:  runList,
}

func init() {
	listCmd.Flags().IntP("limit", "n", 20, "显示最近 N 条记录")
}

func runList(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".napsec", "audit")

	logger, err := audit.NewLogger(logDir)
	if err != nil {
		fmt.Println("暂无保护记录")
		return nil
	}

	limit, _ := cmd.Flags().GetInt("limit")
	records, err := logger.GetRecords(limit)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Println("暂无保护记录")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "时间\t操作\t原始路径\t")
	fmt.Fprintln(w, "────────────────────\t────────\t──────────────────────\t")

	for _, r := range records {
		fmt.Fprintf(w, "%s\t%s\t%s\t\n",
			r.Timestamp.Format(time.DateTime),
			r.Operation,
			r.OriginalPath,
		)
	}
	w.Flush()

	return nil
}
