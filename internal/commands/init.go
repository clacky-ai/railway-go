package commands

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

// NewInitCommand 创建初始化命令
func NewInitCommand(cfg *config.Config) *cobra.Command {
	var projectName string

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "创建新的Railway项目",
		Long:  "创建一个新的Railway项目并将当前目录链接到该项目。",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				projectName = args[0]
			}
			return runInit(cfg, projectName)
		},
	}

	cmd.Flags().StringVarP(&projectName, "name", "n", "", "项目名称")

	return cmd
}

func runInit(cfg *config.Config, projectName string) error {
	// 认证客户端
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}

	// 选择工作区
	var workspaces gql.UserProjectsResponse
	if err := gqlClient.Query(context.Background(), gql.UserProjectsQuery, nil, &workspaces); err != nil {
		return fmt.Errorf("获取工作区失败: %w", err)
	}

	type ws struct {
		Name   string
		TeamID *string
		Label  string
	}
	var options []ws
	for _, ew := range workspaces.ExternalWorkspaces {
		options = append(options, ws{Name: ew.Name, TeamID: ew.TeamID, Label: ew.Name})
	}
	for _, mw := range workspaces.Me.Workspaces {
		var teamID *string
		if mw.Team != nil {
			tid := mw.Team.ID
			teamID = &tid
		}
		options = append(options, ws{Name: mw.Name, TeamID: teamID, Label: mw.Name})
	}
	if len(options) == 0 {
		return fmt.Errorf("未找到工作区，请先在Railway创建工作区")
	}

	var chosen ws
	if len(options) == 1 {
		chosen = options[0]
		util.PrintInfo(fmt.Sprintf("选择工作区: %s", chosen.Label))
	} else {
		var names []string
		for _, o := range options {
			names = append(names, o.Label)
		}
		picked, err := util.PromptSelect("选择一个工作区", names)
		if err != nil {
			return err
		}
		for _, o := range options {
			if o.Label == picked {
				chosen = o
				break
			}
		}
	}

	// 项目名输入或随机生成
	if projectName == "" {
		name, err := util.PromptText("项目名称 (留空自动生成):")
		if err != nil {
			return err
		}
		projectName = strings.TrimSpace(name)
		if projectName == "" {
			projectName = generateRandomProjectName()
		}
	}

	util.PrintInfo(fmt.Sprintf("创建项目 '%s'...", projectName))

	// 创建项目（返回环境列表）
	vars := map[string]interface{}{
		"name":        projectName,
		"description": nil,
		"teamId":      chosen.TeamID,
	}
	var pr gql.ProjectCreateResponse
	if err := gqlClient.Mutate(context.Background(), gql.ProjectCreateMutation, vars, &pr); err != nil {
		return fmt.Errorf("创建项目失败: %w", err)
	}

	if len(pr.ProjectCreate.Environments.Edges) == 0 {
		return fmt.Errorf("未获取到创建后的环境信息")
	}
	envID := pr.ProjectCreate.Environments.Edges[0].Node.ID
	envName := pr.ProjectCreate.Environments.Edges[0].Node.Name

	// 链接项目
	if err := cfg.LinkProject(pr.ProjectCreate.ID, envID, &pr.ProjectCreate.Name, &envName); err != nil {
		return fmt.Errorf("链接项目失败: %w", err)
	}

	util.PrintSuccess(fmt.Sprintf("Created project %s on %s", pr.ProjectCreate.Name, chosen.Label))
	fmt.Printf("%s\n", fmt.Sprintf("https://%s/project/%s", cfg.GetHost(), pr.ProjectCreate.ID))
	return nil
}

// 生成简易的随机项目名（形容词-名词）
func generateRandomProjectName() string {
	adjectives := []string{"brave", "calm", "eager", "fancy", "gentle", "happy", "jolly", "kind", "lucky", "merry", "nice", "proud", "quick", "royal", "smart", "tidy", "witty"}
	nouns := []string{"pine", "river", "cloud", "meadow", "sun", "moon", "star", "hill", "field", "lake", "forest", "breeze", "shadow", "flame", "stone", "leaf"}
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s-%s", adjectives[rand.Intn(len(adjectives))], nouns[rand.Intn(len(nouns))])
}
