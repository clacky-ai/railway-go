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

func NewProjectCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "项目相关操作",
	}
	cmd.AddCommand(NewProjectDeleteCommand(cfg))
	return cmd
}

func NewProjectDeleteCommand(cfg *config.Config) *cobra.Command {
	var projectID string
	var force bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "删除项目",
		Long:  "删除一个Railway项目。若未指定ID，则删除当前目录已链接的项目。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectDelete(cfg, projectID, force)
		},
	}
	cmd.Flags().StringVarP(&projectID, "id", "i", "", "要删除的项目ID（默认使用已链接项目）")
	cmd.Flags().BoolVarP(&force, "yes", "y", false, "无需确认，直接删除")
	return cmd
}

func runProjectDelete(cfg *config.Config, projectID string, force bool) error {
	// 若未指定ID，读取已链接项目
	if strings.TrimSpace(projectID) == "" {
		linked, err := cfg.GetLinkedProject()
		if err != nil {
			return fmt.Errorf("未指定项目ID，且当前目录未链接项目: %w", err)
		}
		projectID = linked.Project
	}

	if !force {
		ok, err := util.PromptConfirm(fmt.Sprintf("确定要删除项目 %s ? 此操作不可撤销!", projectID))
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

	variables := map[string]interface{}{"id": projectID}
	var resp gql.ProjectDeleteResponse
	if err := gqlClient.Mutate(context.Background(), gql.ProjectDeleteMutation, variables, &resp); err != nil {
		return fmt.Errorf("删除项目失败: %w", err)
	}
	if !resp.ProjectDelete {
		return fmt.Errorf("删除项目失败: 后端返回false")
	}

	// 若当前目录链接到该项目，自动unlink
	if linked, err := cfg.GetLinkedProject(); err == nil && linked.Project == projectID {
		_ = cfg.UnlinkProject()
	}

	util.PrintSuccess(fmt.Sprintf("项目 %s 已删除", projectID))
	return nil
}
