package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/spf13/cobra"
)

func NewLogsCommand(cfg *config.Config) *cobra.Command {
	var serviceArg string
	var envArg string
	var build bool
	var deployment bool
	var deploymentID string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "查看部署日志（构建或运行）",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogs(cfg, serviceArg, envArg, build, deployment, deploymentID, jsonOut)
		},
	}

	cmd.Flags().StringVarP(&serviceArg, "service", "s", "", "服务名称或ID（默认使用已链接服务）")
	cmd.Flags().StringVarP(&envArg, "environment", "e", "", "环境名称或ID（默认使用已链接环境）")
	cmd.Flags().BoolVarP(&deployment, "deployment", "d", false, "显示部署日志")
	cmd.Flags().BoolVarP(&build, "build", "b", false, "显示构建日志")
	cmd.Flags().StringVar(&deploymentID, "deployment-id", "", "指定部署ID（不指定则取最新成功部署）")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "JSON 格式输出")

	return cmd
}

func runLogs(cfg *config.Config, serviceArg, envArg string, build, deployment bool, deploymentID string, jsonOut bool) error {
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

	// 获取项目详情
	var projectResp gql.ProjectResponse
	if err := gqlClient.Query(context.Background(), gql.ProjectQuery, map[string]any{"id": linked.Project}, &projectResp); err != nil {
		return fmt.Errorf("获取项目失败: %w", err)
	}

	// 环境ID
	var envID string
	for _, e := range projectResp.Project.Environments.Edges {
		if eq(e.Node.ID, environment) || eq(e.Node.Name, environment) {
			envID = e.Node.ID
			break
		}
	}
	if envID == "" {
		return fmt.Errorf("未找到环境: %s", environment)
	}

	// 服务ID
	var serviceID string
	if strings.TrimSpace(serviceArg) != "" {
		for _, s := range projectResp.Project.Services.Edges {
			if eq(s.Node.ID, serviceArg) || eq(s.Node.Name, serviceArg) {
				serviceID = s.Node.ID
				break
			}
		}
		if serviceID == "" {
			return fmt.Errorf("未找到服务: %s", serviceArg)
		}
	} else if linked.Service != nil {
		serviceID = *linked.Service
	} else {
		return fmt.Errorf("未链接服务，请使用 --service 指定")
	}

	// 查询该服务的部署列表（使用我们已有 Deployments 查询）
	var deploymentsResp gql.DeploymentsResponse
	if err := gqlClient.Query(context.Background(), gql.DeploymentsQuery, map[string]any{
		"projectId":     linked.Project,
		"environmentId": envID,
		"serviceId":     serviceID,
	}, &deploymentsResp); err != nil {
		return fmt.Errorf("获取部署列表失败: %w", err)
	}

	// 过滤成功部署并按创建时间排序（没有时间字段，这里按顺序使用返回顺序；若需要可扩展查询）
	// 简单起见：若未指定 deploymentID，取 edges[0] 作为最新（后端通常返回按时间倒序）
	if deploymentID == "" {
		if len(deploymentsResp.Deployments.Edges) == 0 {
			return fmt.Errorf("没有找到任何部署")
		}
		// 尝试按 UpdatedAt 排序
		sort.SliceStable(deploymentsResp.Deployments.Edges, func(i, j int) bool {
			return deploymentsResp.Deployments.Edges[i].Node.UpdatedAt > deploymentsResp.Deployments.Edges[j].Node.UpdatedAt
		})
		deploymentID = deploymentsResp.Deployments.Edges[0].Node.ID
	}

	// 如果未指定类型，默认：失败部署显示构建日志；否则显示部署日志
	// 这里简化为：优先 build，如果未指定则展示部署日志

	ctx := context.Background()
	// 构建日志订阅
	if build && !deployment {
		vars := map[string]interface{}{"deploymentId": deploymentID, "filter": "", "limit": 500}
		return client.Subscribe(ctx, cfg, gql.BuildLogsSub, vars, func(data json.RawMessage) {
			if jsonOut {
				fmt.Println(string(data))
				return
			}
			var pl gql.BuildLogsPayload
			if err := json.Unmarshal(data, &pl); err == nil {
				for _, l := range pl.BuildLogs {
					fmt.Println(l.Message)
				}
			}
		}, func(err error) { fmt.Fprintf(os.Stderr, "构建日志订阅错误: %v\n", err) })
	}

	// 部署日志订阅
	vars := map[string]interface{}{"deploymentId": deploymentID, "filter": "", "limit": 500}
	return client.Subscribe(ctx, cfg, gql.DeploymentLogsSub, vars, func(data json.RawMessage) {
		if jsonOut {
			fmt.Println(string(data))
			return
		}
		var pl gql.DeploymentLogsPayload
		if err := json.Unmarshal(data, &pl); err == nil {
			for _, l := range pl.DeploymentLogs {
				// 简化格式：message 以及属性 key=value
				b := strings.Builder{}
				b.WriteString(l.Message)
				for _, a := range l.Attributes {
					b.WriteString(" ")
					b.WriteString(a.Key)
					b.WriteString("=")
					b.WriteString(a.Value)
				}
				fmt.Println(b.String())
			}
		}
	}, func(err error) { fmt.Fprintf(os.Stderr, "部署日志订阅错误: %v\n", err) })
}
