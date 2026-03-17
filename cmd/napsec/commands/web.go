package commands

import (
	"context"
	"fmt"
	"github.com/liangach/napsec/internal/core"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "启动 Web 仪表盘",
	Long: `启动 NapSec 的 Web 仪表盘，通过浏览器查看实时状态。

默认地址: http://localhost:8080`,
	RunE: runWeb,
}

func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Web 服务端口")
	webCmd.Flags().Bool("dev", false, "开发模式")
}

func runWeb(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetInt("port")

	cfg := core.DefaultConfig()
	engine, err := core.NewEngine(cfg)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	engine.RegisterRoutes(mux)

	// 静态文件服务
	mux.Handle("/", http.FileServer(http.Dir("web")))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	fmt.Printf("Web 仪表盘已启动: http://localhost:%d\n", port)
	fmt.Println("按 Ctrl+C 停止")

	go func() {
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "服务器错误: %v\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}
