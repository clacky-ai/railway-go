package commands

import (
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

// NewLogoutCommand 创建登出命令
func NewLogoutCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "登出你的Railway账户",
		Long:  "从本地配置中移除认证令牌，登出你的Railway账户。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cfg)
		},
	}

	return cmd
}

func runLogout(cfg *config.Config) error {
	// 重置配置
	if err := cfg.Reset(); err != nil {
		return err
	}

	util.PrintSuccess("已成功登出")
	return nil
}
