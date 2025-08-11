package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/railwayapp/cli/internal/commands"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

var version = "4.6.1"

func main() {
	// 设置优雅关闭
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 初始化配置
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 创建根命令
	rootCmd := &cobra.Command{
		Use:     "railway",
		Short:   "Railway CLI - 与Railway基础设施交互",
		Long:    "Railway命令行界面(CLI)允许你从命令行连接代码到Railway项目，无需担心环境变量或配置。",
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 检查更新（后台运行）
			go util.CheckForUpdates(version)
		},
	}

	// 添加全局标志
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")
	rootCmd.PersistentFlags().Bool("json", false, "JSON格式输出")

	// 添加所有命令
	commands.AddAllCommands(rootCmd, cfg)

	// 执行命令
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
