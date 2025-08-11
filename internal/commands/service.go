package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

func NewServiceCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "服务管理",
	}
	cmd.AddCommand(newServiceListCmd(cfg))
	cmd.AddCommand(newServiceCreateCmd(cfg))
	cmd.AddCommand(newServiceDeleteCmd(cfg))
	cmd.AddCommand(newServiceLinkCmd(cfg))
	cmd.AddCommand(newServiceUnlinkCmd(cfg))
	return cmd
}

func newServiceListCmd(cfg *config.Config) *cobra.Command {
	var envArg string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出项目服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			gqlClient, err := client.NewAuthorized(cfg)
			if err != nil {
				return fmt.Errorf("请先登录: %w", err)
			}
			linked, err := cfg.GetLinkedProject()
			if err != nil {
				return err
			}
			environment := envArg
			if strings.TrimSpace(environment) == "" {
				environment = linked.Environment
			}
			var proj gql.ProjectResponse
			if err := gqlClient.Query(context.Background(), gql.ProjectQuery, map[string]any{"id": linked.Project}, &proj); err != nil {
				return err
			}
			// 解析环境ID
			var envID string
			for _, e := range proj.Project.Environments.Edges {
				if eq(e.Node.ID, environment) || eq(e.Node.Name, environment) {
					envID = e.Node.ID
					break
				}
			}
			if envID == "" {
				return fmt.Errorf("未找到环境: %s", environment)
			}
			fmt.Printf("%-36s  %-20s  %-s\n", "ID", "Name", "HasInstanceInEnv")
			for _, s := range proj.Project.Services.Edges {
				has := false
				for _, inst := range s.Node.ServiceInstances.Edges {
					if inst.Node.EnvironmentID == envID {
						has = true
						break
					}
				}
				fmt.Printf("%-36s  %-20s  %-v\n", s.Node.ID, s.Node.Name, has)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&envArg, "environment", "e", "", "环境名称或ID")
	return cmd
}

func newServiceCreateCmd(cfg *config.Config) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("请使用 --name 指定服务名")
			}
			gqlClient, err := client.NewAuthorized(cfg)
			if err != nil {
				return fmt.Errorf("请先登录: %w", err)
			}
			linked, err := cfg.GetLinkedProject()
			if err != nil {
				return err
			}
			input := gql.ServiceCreateInput{ProjectID: linked.Project, Name: name}
			var resp gql.ServiceCreateResponse
			if err := gqlClient.Mutate(context.Background(), gql.ServiceCreateMutation, map[string]any{"input": input}, &resp); err != nil {
				return err
			}
			util.PrintSuccess(fmt.Sprintf("服务已创建: %s (%s)", resp.ServiceCreate.Name, resp.ServiceCreate.ID))
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "服务名称")
	return cmd
}

func newServiceDeleteCmd(cfg *config.Config) *cobra.Command {
	var id string
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "删除服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(id) == "" {
				return fmt.Errorf("请使用 --id 指定服务ID")
			}
			if !yes {
				ok, err := util.PromptConfirm(fmt.Sprintf("确认删除服务 %s ?", id))
				if err != nil {
					return err
				}
				if !ok {
					fmt.Println("已取消")
					return nil
				}
			}
			gqlClient, err := client.NewAuthorized(cfg)
			if err != nil {
				return fmt.Errorf("请先登录: %w", err)
			}
			var resp gql.ServiceDeleteResponse
			if err := gqlClient.Mutate(context.Background(), gql.ServiceDeleteMutation, map[string]any{"id": id}, &resp); err != nil {
				return err
			}
			if !resp.ServiceDelete {
				return fmt.Errorf("后端返回删除失败")
			}
			util.PrintSuccess("服务已删除")
			return nil
		},
	}
	cmd.Flags().StringVarP(&id, "id", "i", "", "服务ID")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "无需确认")
	return cmd
}

func newServiceLinkCmd(cfg *config.Config) *cobra.Command {
	var idOrName string
	cmd := &cobra.Command{
		Use:   "link",
		Short: "将当前目录链接到某个服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(idOrName) == "" {
				return fmt.Errorf("请使用 --service 指定服务ID或名称")
			}
			linked, err := cfg.GetLinkedProject()
			if err != nil {
				return err
			}
			gqlClient, err := client.NewAuthorized(cfg)
			if err != nil {
				return err
			}
			var proj gql.ProjectResponse
			if err := gqlClient.Query(context.Background(), gql.ProjectQuery, map[string]any{"id": linked.Project}, &proj); err != nil {
				return err
			}
			var svcID string
			for _, s := range proj.Project.Services.Edges {
				if eq(s.Node.ID, idOrName) || eq(s.Node.Name, idOrName) {
					svcID = s.Node.ID
					break
				}
			}
			if svcID == "" {
				return fmt.Errorf("未找到服务: %s", idOrName)
			}
			if err := cfg.LinkService(svcID); err != nil {
				return err
			}
			util.PrintSuccess("服务已链接到当前目录")
			return nil
		},
	}
	cmd.Flags().StringVarP(&idOrName, "service", "s", "", "服务ID或名称")
	return cmd
}

func newServiceUnlinkCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "取消当前目录的服务链接",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UnlinkService(); err != nil {
				return err
			}
			util.PrintSuccess("已取消服务链接")
			return nil
		},
	}
	return cmd
}
