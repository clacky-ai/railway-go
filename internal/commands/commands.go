package commands

import (
	"github.com/railwayapp/cli/internal/config"
	"github.com/spf13/cobra"
)

// AddAllCommands 添加所有命令到根命令
func AddAllCommands(rootCmd *cobra.Command, cfg *config.Config) {
	// 认证相关命令
	rootCmd.AddCommand(NewLoginCommand(cfg))
	rootCmd.AddCommand(NewLogoutCommand(cfg))
	rootCmd.AddCommand(NewWhoamiCommand(cfg))

	// 项目管理命令
	rootCmd.AddCommand(NewInitCommand(cfg))
	rootCmd.AddCommand(NewLinkCommand(cfg))
	rootCmd.AddCommand(NewUnlinkCommand(cfg))
	rootCmd.AddCommand(NewListCommand(cfg))
	rootCmd.AddCommand(NewAddCommand(cfg))
	proj := NewProjectCommand(cfg)
	proj.AddCommand(NewProjectTokenCommand(cfg))
	rootCmd.AddCommand(proj)

	// 部署相关命令
	rootCmd.AddCommand(NewUpCommand(cfg))
	rootCmd.AddCommand(NewDeployCommand(cfg))
	rootCmd.AddCommand(NewRedeployCommand(cfg))
	rootCmd.AddCommand(NewDownCommand(cfg))

	// 服务管理命令
	rootCmd.AddCommand(NewServiceCommand(cfg))
	rootCmd.AddCommand(NewStatusCommand(cfg))
	rootCmd.AddCommand(NewLogsCommand(cfg))

	// 域名管理命令
	rootCmd.AddCommand(NewDomainCommand(cfg))

	// 环境变量命令
	rootCmd.AddCommand(NewVariablesCommand(cfg))
	rootCmd.AddCommand(NewRunCommand(cfg))

	// 其他命令
	rootCmd.AddCommand(NewOpenCommand(cfg))
	rootCmd.AddCommand(NewDocsCommand(cfg))
	rootCmd.AddCommand(NewCompletionCommand())
}
