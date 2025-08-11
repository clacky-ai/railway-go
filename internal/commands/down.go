package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

// NewDownCommand 删除最近一次成功的部署
func NewDownCommand(cfg *config.Config) *cobra.Command {
	var (
		service     string
		environment string
		yes         bool
	)

	cmd := &cobra.Command{
		Use:   "down",
		Short: "删除最近一次成功的部署",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDown(cfg, service, environment, yes)
		},
	}

	cmd.Flags().StringVarP(&service, "service", "s", "", "服务ID或名称（默认使用已链接服务）")
	cmd.Flags().StringVarP(&environment, "environment", "e", "", "环境ID或名称（默认使用已链接环境）")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "无需确认，直接删除")

	return cmd
}

func runDown(cfg *config.Config, serviceArg, environmentArg string, yes bool) error {
	linked, err := cfg.GetLinkedProject()
	if err != nil {
		return fmt.Errorf("未找到已链接的项目: %w", err)
	}

	// 解析环境
	envIdentifier := strings.TrimSpace(environmentArg)
	if envIdentifier == "" {
		envIdentifier = linked.Environment
	}

	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}

	// 查询项目，解析环境与服务
	var proj gql.ProjectResponse
	if err := gqlClient.Query(context.Background(), gql.ProjectQuery, map[string]any{"id": linked.Project}, &proj); err != nil {
		return err
	}

	// 找到环境ID
	var environmentID string
	for _, e := range proj.Project.Environments.Edges {
		if eq(e.Node.ID, envIdentifier) || eq(e.Node.Name, envIdentifier) {
			environmentID = e.Node.ID
			break
		}
	}
	if environmentID == "" {
		return fmt.Errorf("未找到环境: %s", envIdentifier)
	}

	// 解析服务ID
	var serviceID string
	if s := strings.TrimSpace(serviceArg); s != "" {
		for _, se := range proj.Project.Services.Edges {
			if eq(se.Node.ID, s) || eq(se.Node.Name, s) {
				serviceID = se.Node.ID
				break
			}
		}
		if serviceID == "" {
			return fmt.Errorf("未找到服务: %s", s)
		}
	} else if linked.Service != nil && *linked.Service != "" {
		serviceID = *linked.Service
	} else {
		return fmt.Errorf("未链接服务，请使用 -s/--service 指定服务ID或名称")
	}

	// 查询部署列表
	vars := map[string]any{
		"projectId":     linked.Project,
		"environmentId": environmentID,
		"serviceId":     serviceID,
	}
	var deps gql.DeploymentsResponse
	if err := gqlClient.Query(context.Background(), gql.DeploymentsQuery, vars, &deps); err != nil {
		return err
	}

	// 过滤成功部署并按时间倒序
	type depNode = struct {
		ID        string
		Status    string
		CreatedAt string
	}
	nodes := make([]depNode, 0, len(deps.Deployments.Edges))
	for _, ed := range deps.Deployments.Edges {
		if strings.EqualFold(ed.Node.Status, "SUCCESS") {
			nodes = append(nodes, depNode{ID: ed.Node.ID, Status: ed.Node.Status, CreatedAt: ed.Node.CreatedAt})
		}
	}
	if len(nodes) == 0 {
		return fmt.Errorf("未找到成功的部署")
	}
	sort.Slice(nodes, func(i, j int) bool {
		// 优先尝试按时间解析比较
		ti, ei := parseTime(nodes[i].CreatedAt)
		tj, ej := parseTime(nodes[j].CreatedAt)
		if ei == nil && ej == nil {
			return ti.After(tj)
		}
		// 回退到字符串倒序
		return nodes[i].CreatedAt > nodes[j].CreatedAt
	})
	latest := nodes[0]

	// 确认
	if !yes {
		projName := ""
		if proj.Project.Name != "" {
			projName = proj.Project.Name
		} else if linked.Name != nil {
			projName = *linked.Name
		} else {
			projName = linked.Project
		}
		envName := ""
		if linked.EnvironmentName != nil {
			envName = *linked.EnvironmentName
		} else {
			envName = environmentID
		}
		ok, err := util.PromptConfirm(fmt.Sprintf("确定要删除项目 %s 的环境 %s 的最新部署吗?", projName, envName))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("已取消")
			return nil
		}
	}

	// 执行删除
	var resp struct {
		DeploymentRemove bool `json:"deploymentRemove"`
	}
	if err := gqlClient.Mutate(context.Background(), gql.DeploymentRemoveMutation, map[string]any{"id": latest.ID}, &resp); err != nil {
		return err
	}
	if !resp.DeploymentRemove {
		return fmt.Errorf("后端返回删除失败")
	}

	util.PrintSuccess("最近一次部署已删除")
	return nil
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// 常见回退格式
	layouts := []string{
		"2006-01-02T15:04:05.000Z07:00",
		time.DateTime,
		time.RFC1123Z,
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time: %s", s)
}
