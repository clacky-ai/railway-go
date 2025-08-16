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
	var serviceOpt string
	var name string
	var repo string
	var image string
	var variables []string
	var databases []string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "向项目添加服务/数据库",
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceFlagSet := cmd.Flags().Changed("service")
			return runAdd(cfg, serviceFlagSet, serviceOpt, name, repo, image, databases, variables, cmd)
		},
	}

	// 与 Rust 行为对齐：--service 可选值（留空表示随机名）
	cmd.Flags().StringVarP(&serviceOpt, "service", "s", "", "创建空服务（可选值为服务名，留空则随机）")
	// 兼容原有 --name（如果同时提供，以 --service 值为准）
	cmd.Flags().StringVarP(&name, "name", "n", "", "服务名称（留空则随机；若使用 --service 传值，则优先生效）")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "GitHub 仓库 <owner>/<repo>")
	cmd.Flags().StringVarP(&image, "image", "i", "", "Docker 镜像")
	cmd.Flags().StringArrayVar(&variables, "variables", []string{}, "初始环境变量，支持多次，如 --variables KEY=VAL")
	cmd.Flags().StringArrayVar(&databases, "database", []string{}, "添加数据库（支持: postgres, mysql, redis, mongo）")

	// 允许 --service 不带值使用
	if f := cmd.Flags().Lookup("service"); f != nil {
		f.NoOptDefVal = ""
	}

	return cmd
}

func runAdd(cfg *config.Config, serviceFlagSet bool, serviceOpt, name, repo, image string, databases, variables []string, cmd *cobra.Command) error {
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}
	linked, err := cfg.GetLinkedProject()
	if err != nil {
		return err
	}

	// 若指定数据库：通过模板部署
	if len(databases) > 0 {
		return addDatabasesViaTemplates(gqlClient, linked.Project, linked.Environment, databases)
	}

	// 决定创建类型
	switch {
	case strings.TrimSpace(repo) != "":
		varMap := parseOrPromptVariables(variables)
		return createService(gqlClient, cfg, linked.Project, linked.Environment, deriveServiceName(serviceFlagSet, serviceOpt, name, true), &gql.Source{Repo: &repo}, varMap, true)
	case strings.TrimSpace(image) != "":
		varMap := parseOrPromptVariables(variables)
		return createService(gqlClient, cfg, linked.Project, linked.Environment, deriveServiceName(serviceFlagSet, serviceOpt, name, true), &gql.Source{Image: &image}, varMap, false)
	case serviceFlagSet:
		// 空服务
		varMap := parseOrPromptVariables(variables)
		return createService(gqlClient, cfg, linked.Project, linked.Environment, deriveServiceName(true, serviceOpt, name, false), nil, varMap, false)
	default:
		// 交互式流程：What do you need?
		choice, err := util.PromptSelect("What do you need?", []string{"Database", "GitHub Repo", "Docker Image", "Empty Service"})
		if err != nil {
			return err
		}
		switch choice {
		case "Database":
			// 交互选择数据库类型
			opts, err := util.PromptMultiSelect("Select databases to add", []string{"postgres", "mysql", "redis", "mongo"})
			if err != nil {
				return err
			}
			if len(opts) == 0 {
				return fmt.Errorf("please select at least one database to add")
			}
			return addDatabasesViaTemplates(gqlClient, linked.Project, linked.Environment, opts)
		case "GitHub Repo":
			r, err := util.PromptText("Enter a repo (<user/org>/<repo name>)")
			if err != nil {
				return err
			}
			varMap := parseOrPromptVariables(variables)
			return createService(gqlClient, cfg, linked.Project, linked.Environment, deriveServiceName(false, "", name, true), &gql.Source{Repo: &r}, varMap, true)
		case "Docker Image":
			img, err := util.PromptText("Enter an image")
			if err != nil {
				return err
			}
			varMap := parseOrPromptVariables(variables)
			return createService(gqlClient, cfg, linked.Project, linked.Environment, deriveServiceName(false, "", name, true), &gql.Source{Image: &img}, varMap, false)
		case "Empty Service":
			varMap := parseOrPromptVariables(variables)
			return createService(gqlClient, cfg, linked.Project, linked.Environment, deriveServiceName(true, "", name, false), nil, varMap, false)
		}
	}

	return nil
}

func addDatabasesViaTemplates(gqlClient *client.Client, projectID, environmentID string, databases []string) error {
	sort.Strings(databases)

	// 创建 linkedProject 结构
	linkedProject := &config.LinkedProject{
		Project:     projectID,
		Environment: environmentID,
	}

	// 创建空的变量 map
	vars := map[string]string{}

	for _, db := range databases {
		code := strings.ToLower(strings.TrimSpace(db))
		if code == "postgresql" {
			code = "postgres"
		}
		if code == "mongodb" {
			code = "mongo"
		}
		if code != "postgres" && code != "mysql" && code != "redis" && code != "mongo" {
			fmt.Printf("未知数据库类型: %s，跳过\n", db)
			continue
		}

		// 使用 fetchAndCreate 方法部署数据库模板
		if err := fetchAndCreate(gqlClient, nil, code, linkedProject, vars); err != nil {
			return fmt.Errorf("创建数据库失败(%s): %w", code, err)
		}
	}
	return nil
}

func parseOrPromptVariables(pairs []string) map[string]*string {
	vars := map[string]*string{}
	for _, v := range pairs {
		k, val, ok := splitKV(v)
		if !ok {
			continue
		}
		vv := val
		vars[k] = &vv
	}
	if len(pairs) == 0 {
		for {
			v, err := util.PromptText("Enter a variable <KEY=VALUE, press enter to skip>")
			if err != nil {
				break
			}
			if strings.TrimSpace(v) == "" {
				break
			}
			k, val, ok := splitKV(v)
			if !ok || strings.TrimSpace(val) == "" {
				continue
			}
			vv := val
			vars[k] = &vv
		}
	}
	return vars
}

func deriveServiceName(serviceFlagSet bool, serviceOpt, name string, promptIfInteractive bool) string {
	// 优先：--service 值 > --name > 交互输入 > 空字符串（由后端随机生成）
	if serviceFlagSet {
		if strings.TrimSpace(serviceOpt) != "" {
			return strings.TrimSpace(serviceOpt)
		}
		// 显式传入 --service 且未给值，表示随机
		return ""
	}
	if strings.TrimSpace(serviceOpt) != "" {
		return strings.TrimSpace(serviceOpt)
	}
	if strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}
	if promptIfInteractive {
		if s, err := util.PromptText("Enter a service name <leave blank for randomly generated>"); err == nil {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func runCreate(gqlClient *client.Client, input gql.ServiceCreateInput, cfg *config.Config) error {
	var resp gql.ServiceCreateResponse
	err := gqlClient.Mutate(context.Background(), gql.ServiceCreateMutation, map[string]any{"input": input}, &resp)
	if err != nil {
		// 回退：不带 variables/environmentId/branch
		inputFallback := gql.ServiceCreateInput{ProjectID: input.ProjectID, Name: input.Name, Source: input.Source}
		if e := gqlClient.Mutate(context.Background(), gql.ServiceCreateMutation, map[string]any{"input": inputFallback}, &resp); e != nil {
			return fmt.Errorf("创建服务失败: %w", err)
		}
		if len(input.Variables) > 0 {
			up := gql.VariableCollectionUpsertInput{ProjectID: input.ProjectID, EnvironmentID: input.EnvironmentID, ServiceID: &resp.ServiceCreate.ID, Variables: input.Variables}
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

func createService(gqlClient *client.Client, cfg *config.Config, projectID, environmentID string, serviceName string, src *gql.Source, vars map[string]*string, resolveBranch bool) error {
	// 解析默认分支
	var branch *string
	if resolveBranch && src != nil && src.Repo != nil && strings.TrimSpace(*src.Repo) != "" {
		var gr gql.GitHubReposResponse
		if err := gqlClient.Query(context.Background(), gql.GitHubReposQuery, nil, &gr); err == nil {
			for _, r := range gr.GitHubRepos {
				if r.FullName == *src.Repo {
					b := r.DefaultBranch
					branch = &b
					break
				}
			}
		}
	}
	input := gql.ServiceCreateInput{
		ProjectID:     projectID,
		Name:          strings.TrimSpace(serviceName),
		Source:        src,
		EnvironmentID: environmentID,
		Variables:     vars,
		Branch:        branch,
	}
	return runCreate(gqlClient, input, cfg)
}

func firstNonEmpty(s string) string { return strings.TrimSpace(s) }
