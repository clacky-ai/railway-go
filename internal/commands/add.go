package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

func NewAddCommand(cfg *config.Config) *cobra.Command {
	var svcFlag bool
	var name string
	var repo string
	var image string
	var variables []string
	var databases []string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "向项目添加服务/数据库",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(cfg, svcFlag, name, repo, image, databases, variables)
		},
	}

	cmd.Flags().BoolVarP(&svcFlag, "service", "s", false, "创建空服务（可配合 --name）")
	cmd.Flags().StringVarP(&name, "name", "n", "", "服务名称（留空则随机）")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "GitHub 仓库 <owner>/<repo>")
	cmd.Flags().StringVarP(&image, "image", "i", "", "Docker 镜像")
	cmd.Flags().StringArrayVar(&variables, "variables", []string{}, "初始环境变量，支持多次，如 --variables KEY=VAL")
	cmd.Flags().StringArrayVar(&databases, "database", []string{}, "添加数据库（占位：将通过模板实现）")

	return cmd
}

func runAdd(cfg *config.Config, svcFlag bool, name, repo, image string, databases, variables []string) error {
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}
	linked, err := cfg.GetLinkedProject()
	if err != nil {
		return err
	}

	// 数据库（占位）
	if len(databases) > 0 {
		sort.Strings(databases)
		for _, db := range databases {
			fmt.Printf("TODO: 通过模板添加数据库: %s\n", db)
		}
		return nil
	}

	// variables -> map
	varMap := map[string]*string{}
	for _, v := range variables {
		k, val, ok := splitKV(v)
		if !ok {
			continue
		}
		vv := val
		varMap[k] = &vv
	}

	// Source
	var src *gql.Source
	if strings.TrimSpace(repo) != "" {
		src = &gql.Source{Repo: &repo}
	} else if strings.TrimSpace(image) != "" {
		src = &gql.Source{Image: &image}
	}

	// 一次性创建（与Rust一致）：在 ServiceCreate 里带 environmentId 与 variables
	input := gql.ServiceCreateInput{
		ProjectID:     linked.Project,
		Name:          firstNonEmpty(name),
		Source:        src,
		EnvironmentID: linked.Environment,
		Variables:     varMap,
		Branch:        nil,
	}
	var resp gql.ServiceCreateResponse
	err = gqlClient.Mutate(context.Background(), gql.ServiceCreateMutation, map[string]any{"input": input}, &resp)

	if err != nil {
		// 兼容回退：老后端不支持一次性变量，改为两步
		inputFallback := gql.ServiceCreateInput{
			ProjectID: linked.Project,
			Name:      firstNonEmpty(name),
			Source:    src,
		}
		if e := gqlClient.Mutate(context.Background(), gql.ServiceCreateMutation, map[string]any{"input": inputFallback}, &resp); e != nil {
			return fmt.Errorf("创建服务失败: %w", err)
		}
		if len(varMap) > 0 {
			up := gql.VariableCollectionUpsertInput{
				ProjectID:     linked.Project,
				EnvironmentID: linked.Environment,
				ServiceID:     &resp.ServiceCreate.ID,
				Variables:     varMap,
			}
			var upResp gql.VariableCollectionUpsertResponse
			if e := gqlClient.Mutate(context.Background(), gql.VariableCollectionUpsertMutation, map[string]any{"input": up}, &upResp); e != nil {
				return fmt.Errorf("设置初始变量失败: %w", e)
			}
		}
	}

	// 链接服务
	if err := cfg.LinkService(resp.ServiceCreate.ID); err == nil {
		util.PrintSuccess(fmt.Sprintf("已创建并链接服务: %s (%s)", resp.ServiceCreate.Name, resp.ServiceCreate.ID))
	} else {
		util.PrintSuccess(fmt.Sprintf("已创建服务: %s (%s)", resp.ServiceCreate.Name, resp.ServiceCreate.ID))
	}

	return nil
}

func firstNonEmpty(s string) string { return strings.TrimSpace(s) }
