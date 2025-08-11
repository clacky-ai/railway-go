package commands

import (
	"context"
	"fmt"

	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/spf13/cobra"
)

// NewWhoamiCommand 创建whoami命令
func NewWhoamiCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "显示当前登录的用户信息",
		Long:  "显示当前登录的Railway用户的详细信息。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoami(cfg)
		},
	}

	return cmd
}

func runWhoami(cfg *config.Config) error {
	// 创建认证客户端
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("未登录或认证失败: %w", err)
	}

	// 获取用户信息
	var response gql.UserMetaResponse
	err = gqlClient.Query(context.Background(), gql.UserMetaQuery, nil, &response)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 显示用户信息
	fmt.Printf("用户ID: %s\n", response.Me.ID)
	if response.Me.Name != nil {
		fmt.Printf("姓名: %s\n", *response.Me.Name)
	}
	fmt.Printf("邮箱: %s\n", response.Me.Email)
	if response.Me.Avatar != nil {
		fmt.Printf("头像: %s\n", *response.Me.Avatar)
	}

	return nil
}
