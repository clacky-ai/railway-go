package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

func NewVariablesCommand(cfg *config.Config) *cobra.Command {
	var serviceArg string
	var envArg string
	var kv bool
	var setPairs []string
	var jsonOut bool
	var skipDeploys bool

	cmd := &cobra.Command{
		Use:   "variables",
		Short: "查看或设置环境变量",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVariables(cfg, serviceArg, envArg, kv, setPairs, jsonOut, skipDeploys)
		},
	}

	cmd.Flags().StringVarP(&serviceArg, "service", "s", "", "服务名称或ID（默认使用已链接服务）")
	cmd.Flags().StringVarP(&envArg, "environment", "e", "", "环境名称或ID（默认使用已链接环境）")
	cmd.Flags().BoolVarP(&kv, "kv", "k", false, "以 key=value 形式输出")
	cmd.Flags().StringArrayVar(&setPairs, "set", []string{}, "设置变量，如 --set KEY=VALUE，可重复")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "JSON 格式输出")
	cmd.Flags().BoolVar(&skipDeploys, "skip-deploys", false, "设置变量时跳过触发部署")

	return cmd
}

func runVariables(cfg *config.Config, serviceArg, envArg string, kv bool, setPairs []string, jsonOut, skipDeploys bool) error {
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}

	linked, err := cfg.GetLinkedProject()
	if err != nil {
		return fmt.Errorf("未链接项目: %w", err)
	}

	environment := envArg
	if strings.TrimSpace(environment) == "" {
		environment = linked.Environment
	}

	// 获取项目详情以解析服务和环境ID
	var projectResp gql.ProjectResponse
	if err := gqlClient.Query(context.Background(), gql.ProjectQuery, map[string]any{"id": linked.Project}, &projectResp); err != nil {
		return fmt.Errorf("获取项目失败: %w", err)
	}

	// 解析环境ID
	var environmentID string
	for _, edge := range projectResp.Project.Environments.Edges {
		if eq(edge.Node.ID, environment) || eq(edge.Node.Name, environment) {
			environmentID = edge.Node.ID
			break
		}
	}
	if environmentID == "" {
		return fmt.Errorf("未找到环境: %s", environment)
	}

	// 解析服务ID
	var serviceID string
	if strings.TrimSpace(serviceArg) != "" {
		for _, edge := range projectResp.Project.Services.Edges {
			if eq(edge.Node.ID, serviceArg) || eq(edge.Node.Name, serviceArg) {
				serviceID = edge.Node.ID
				break
			}
		}
		if serviceID == "" {
			return fmt.Errorf("未找到服务: %s", serviceArg)
		}
	} else if linked.Service != nil {
		serviceID = *linked.Service
	}
	if serviceID == "" {
		return fmt.Errorf("当前未链接服务，请使用 --service 指定或先 link service")
	}

	if len(setPairs) > 0 {
		// 解析 setPairs -> map[string]*string
		vars := map[string]*string{}
		for _, p := range setPairs {
			k, v, ok := splitKV(p)
			if !ok || strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
				continue
			}
			vv := v
			vars[k] = &vv
		}
		if len(vars) == 0 {
			fmt.Println("未解析到有效变量键值对")
			return nil
		}

		// 调用 VariableCollectionUpsert
		input := gql.VariableCollectionUpsertInput{
			ProjectID:     linked.Project,
			EnvironmentID: environmentID,
			ServiceID:     &serviceID,
			Replace:       nil,
			Variables:     vars,
		}
		var upsertResp gql.VariableCollectionUpsertResponse
		if err := gqlClient.Mutate(context.Background(), gql.VariableCollectionUpsertMutation,
			map[string]any{"input": input}, &upsertResp); err != nil {
			return fmt.Errorf("设置变量失败: %w", err)
		}
		util.PrintSuccess("变量已设置")
		if !skipDeploys {
			fmt.Println("提示：如需跳过部署，可添加 --skip-deploys")
		}
		return nil
	}

	// 查询变量
	// 由于 GraphQL 查询 VariablesForServiceDeployment 的返回是 map-like，这里沿用我们的 gql.VariablesForServiceDeploymentQuery 并解析
	var varsResp map[string]any
	if err := gqlClient.Query(context.Background(), gql.VariablesForServiceDeploymentQuery, map[string]any{
		"projectId":     linked.Project,
		"environmentId": environmentID,
		"serviceId":     serviceID,
	}, &varsResp); err != nil {
		return fmt.Errorf("获取变量失败: %w", err)
	}

	// 兼容不同根键
	var varsMap map[string]any
	if v, ok := varsResp["variables"]; ok {
		if m, ok := v.(map[string]any); ok {
			varsMap = m
		}
	}
	if varsMap == nil {
		if v, ok := varsResp["variablesForServiceDeployment"]; ok {
			if m, ok := v.(map[string]any); ok {
				varsMap = m
			}
		}
	}
	if varsMap == nil {
		return fmt.Errorf("响应格式不支持")
	}

	// 转换为 map[string]string（过滤空值）
	out := map[string]string{}
	for k, v := range varsMap {
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	if len(out) == 0 {
		fmt.Println("No variables found")
		return nil
	}

	if kv {
		// key=value 排序输出
		keys := make([]string, 0, len(out))
		for k := range out {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s=%s\n", k, out[k])
		}
		return nil
	}

	if jsonOut {
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	// 表格输出
	fmt.Printf("%-30s | %-s\n", "Key", "Value")
	fmt.Println(strings.Repeat("-", 80))
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%-30s | %s\n", k, out[k])
	}

	return nil
}

func splitKV(s string) (string, string, bool) {
	idx := strings.IndexByte(s, '=')
	if idx <= 0 {
		return "", "", false
	}
	return s[:idx], s[idx+1:], true
}
