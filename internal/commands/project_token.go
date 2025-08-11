package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

func NewProjectTokenCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "项目Token管理",
	}
	cmd.AddCommand(newProjectTokenCreateCmd(cfg))
	cmd.AddCommand(newProjectTokenDeleteCmd(cfg))
	cmd.AddCommand(newProjectTokenListCmd(cfg))
	return cmd
}

func newProjectTokenCreateCmd(cfg *config.Config) *cobra.Command {
	var projectID string
	var tokenName string
	var environmentID string
	// 为兼容旧调用，保留 -d/--description 作为别名
	var descriptionAlias string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建项目访问Token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenName == "" {
				tokenName = descriptionAlias
			}
			return runProjectTokenCreate(cfg, projectID, environmentID, tokenName)
		},
	}
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "项目ID（默认使用已链接项目）")
	cmd.Flags().StringVarP(&environmentID, "environment", "e", "", "环境ID（默认使用已链接环境）")
	cmd.Flags().StringVarP(&tokenName, "name", "n", "", "Token名称")
	cmd.Flags().StringVarP(&descriptionAlias, "description", "d", "", "Token名称（兼容别名）")
	return cmd
}

func newProjectTokenDeleteCmd(cfg *config.Config) *cobra.Command {
	var tokenID string
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "删除项目访问Token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectTokenDelete(cfg, tokenID, yes)
		},
	}
	cmd.Flags().StringVarP(&tokenID, "id", "i", "", "要删除的Token ID")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "无需确认")
	return cmd
}

func newProjectTokenListCmd(cfg *config.Config) *cobra.Command {
	var projectID string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出项目访问Token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectTokenList(cfg, projectID)
		},
	}
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "项目ID（默认使用已链接项目）")
	return cmd
}

func runProjectTokenCreate(cfg *config.Config, projectID, environmentID, tokenName string) error {
	if strings.TrimSpace(projectID) == "" {
		linked, err := cfg.GetLinkedProject()
		if err != nil {
			return fmt.Errorf("未指定项目ID，且未链接项目: %w", err)
		}
		projectID = linked.Project
		if strings.TrimSpace(environmentID) == "" {
			environmentID = linked.Environment
		}
	}
	if strings.TrimSpace(environmentID) == "" {
		return fmt.Errorf("请使用 -e/--environment 指定环境ID，或先链接项目以自动获取")
	}
	if strings.TrimSpace(tokenName) == "" {
		tokenName = "cli-token"
	}

	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}

	// 使用官方形态：projectTokenCreate(input: { name, projectId, environmentId }): String
	var raw map[string]any
	input := map[string]any{
		"name":          tokenName,
		"projectId":     projectID,
		"environmentId": environmentID,
	}
	query := "mutation($input:ProjectTokenCreateInput!){ projectTokenCreate(input:$input) }"
	if err := gqlClient.Mutate(context.Background(), query, map[string]any{"input": input}, &raw); err == nil {
		if token, ok := raw["projectTokenCreate"].(string); ok && token != "" {
			util.PrintSuccess("项目Token已创建")
			fmt.Printf("Token: %s\n", token)
			return nil
		}
	}

	// 退路：部分后端可能接受参数式（带 name）
	fallbacks := []struct {
		query string
		vars  map[string]any
		key   string
	}{
		{query: "mutation($projectId:String!,$environmentId:String!,$name:String!){ projectTokenCreate(projectId:$projectId, environmentId:$environmentId, name:$name) }", vars: map[string]any{"projectId": projectID, "environmentId": environmentID, "name": tokenName}, key: "projectTokenCreate"},
	}
	for _, fb := range fallbacks {
		raw = map[string]any{}
		if err := gqlClient.Mutate(context.Background(), fb.query, fb.vars, &raw); err != nil {
			continue
		}
		if v, ok := raw[fb.key].(string); ok && v != "" {
			util.PrintSuccess("项目Token已创建")
			fmt.Printf("Token: %s\n", v)
			return nil
		}
	}

	return errors.New("创建项目Token失败：后端未支持的API或返回异常")
}

func runProjectTokenDelete(cfg *config.Config, tokenID string, yes bool) error {
	if strings.TrimSpace(tokenID) == "" {
		return fmt.Errorf("请使用 -i/--id 指定要删除的Token ID")
	}

	if !yes {
		ok, err := util.PromptConfirm(fmt.Sprintf("确定要删除项目Token %s ? 此操作不可撤销!", tokenID))
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

	// 首选：按schema使用 id 参数
	type deleteResp struct {
		ProjectTokenDelete bool `json:"projectTokenDelete"`
	}
	var resp deleteResp
	mutation := "mutation($id:String!){ projectTokenDelete(id:$id) }"
	if err := gqlClient.Mutate(context.Background(), mutation, map[string]any{"id": tokenID}, &resp); err == nil {
		if resp.ProjectTokenDelete {
			util.PrintSuccess("项目Token已删除")
			return nil
		}
	}

	// 退路：部分后端可能使用 input 形式
	var fbResp deleteResp
	fbMutation := "mutation($input:ProjectTokenDeleteInput!){ projectTokenDelete(input:$input) }"
	if err := gqlClient.Mutate(context.Background(), fbMutation, map[string]any{"input": map[string]any{"id": tokenID}}, &fbResp); err == nil {
		if fbResp.ProjectTokenDelete {
			util.PrintSuccess("项目Token已删除")
			return nil
		}
	}

	return errors.New("删除项目Token失败：后端未支持的API或返回异常")
}

func runProjectTokenList(cfg *config.Config, projectID string) error {
	if strings.TrimSpace(projectID) == "" {
		linked, err := cfg.GetLinkedProject()
		if err != nil {
			return fmt.Errorf("未指定项目ID，且未链接项目: %w", err)
		}
		projectID = linked.Project
	}

	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}

	query := "query($projectId:String!,$after:String){ projectTokens(projectId:$projectId, first:50, after:$after) { edges { cursor node { id name project { id name } environment { id name } } } pageInfo { hasNextPage endCursor } } }"

	type tokenNode struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Project struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		Environment struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"environment"`
	}
	type edge struct {
		Cursor string    `json:"cursor"`
		Node   tokenNode `json:"node"`
	}
	type pageInfo struct {
		HasNextPage bool    `json:"hasNextPage"`
		EndCursor   *string `json:"endCursor"`
	}
	type resp struct {
		ProjectTokens struct {
			Edges    []edge   `json:"edges"`
			PageInfo pageInfo `json:"pageInfo"`
		} `json:"projectTokens"`
	}

	printedHeader := false
	var after string
	total := 0
	for {
		variables := map[string]any{
			"projectId": projectID,
			"after":     nullIfEmpty(after),
		}
		var r resp
		if err := gqlClient.Query(context.Background(), query, variables, &r); err != nil {
			return err
		}
		if len(r.ProjectTokens.Edges) == 0 && total == 0 {
			fmt.Println("无Token")
			return nil
		}
		if !printedHeader {
			fmt.Printf("%-36s  %-24s  %-s\n", "ID", "Name", "Environment")
			printedHeader = true
		}
		for _, e := range r.ProjectTokens.Edges {
			env := e.Node.Environment.Name
			if strings.TrimSpace(env) == "" {
				env = e.Node.Environment.ID
			}
			fmt.Printf("%-36s  %-24s  %-s\n", e.Node.ID, e.Node.Name, env)
			total++
		}
		if !r.ProjectTokens.PageInfo.HasNextPage || r.ProjectTokens.PageInfo.EndCursor == nil || *r.ProjectTokens.PageInfo.EndCursor == "" {
			break
		}
		after = *r.ProjectTokens.PageInfo.EndCursor
	}
	return nil
}

func extractIDToken(v any) (string, string, bool) {
	m, ok := v.(map[string]any)
	if !ok {
		return "", "", false
	}
	id, _ := m["id"].(string)
	token, _ := m["token"].(string)
	return id, token, id != "" && token != ""
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
