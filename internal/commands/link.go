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

// NewLinkCommand 创建链接命令
func NewLinkCommand(cfg *config.Config) *cobra.Command {
	var (
		envArg     string
		projectArg string
		serviceArg string
		teamArg    string
	)

	cmd := &cobra.Command{
		Use:   "link",
		Short: "将当前目录链接到Railway项目",
		Long:  "选择工作区/项目/环境/服务，将当前目录链接到指定Railway项目。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLinkFull(cfg, envArg, projectArg, serviceArg, teamArg)
		},
	}

	cmd.Flags().StringVarP(&envArg, "environment", "e", "", "要链接的环境ID或名称")
	cmd.Flags().StringVarP(&projectArg, "project", "p", "", "要链接的项目ID或名称")
	cmd.Flags().StringVarP(&serviceArg, "service", "s", "", "要链接的服务ID或名称")
	cmd.Flags().StringVarP(&teamArg, "team", "t", "", "工作区（团队）ID或名称")

	return cmd
}

func runLinkFull(cfg *config.Config, envArg, projectArg, serviceArg, teamArg string) error {
	// 认证客户端
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}

	// 拉取完整工作区/项目数据
	var data gql.UserProjectsFullResponse
	if err := gqlClient.Query(context.Background(), gql.UserProjectsFullQuery, nil, &data); err != nil {
		return fmt.Errorf("获取工作区信息失败: %w", err)
	}

	// 组装可选工作区列表
	type env struct{ ID, Name string }
	type service struct {
		ID, Name string
		EnvIDs   []string
	}
	type project struct {
		ID       string
		Name     string
		Deleted  bool
		Envs     []env
		Services []service
	}
	type workspace struct {
		Name     string
		TeamID   *string
		Projects []project
	}

	var workspaces []workspace

	for _, ew := range data.ExternalWorkspaces {
		ws := workspace{Name: ew.Name, TeamID: ew.TeamID}
		for _, p := range ew.Projects {
			pr := project{ID: p.ID, Name: p.Name, Deleted: p.DeletedAt != nil}
			for _, e := range p.Environments.Edges {
				pr.Envs = append(pr.Envs, env{ID: e.Node.ID, Name: e.Node.Name})
			}
			for _, s := range p.Services.Edges {
				var envIDs []string
				for _, si := range s.Node.ServiceInstances.Edges {
					envIDs = append(envIDs, si.Node.EnvironmentID)
				}
				pr.Services = append(pr.Services, service{ID: s.Node.ID, Name: s.Node.Name, EnvIDs: envIDs})
			}
			ws.Projects = append(ws.Projects, pr)
		}
		workspaces = append(workspaces, ws)
	}
	for _, mw := range data.Me.Workspaces {
		var tid *string
		var projects []project
		if mw.Team != nil {
			tid = &mw.Team.ID
			for _, edge := range mw.Team.Projects.Edges {
				n := edge.Node
				pr := project{ID: n.ID, Name: n.Name, Deleted: n.DeletedAt != nil}
				for _, e := range n.Environments.Edges {
					pr.Envs = append(pr.Envs, env{ID: e.Node.ID, Name: e.Node.Name})
				}
				for _, s := range n.Services.Edges {
					var envIDs []string
					for _, si := range s.Node.ServiceInstances.Edges {
						envIDs = append(envIDs, si.Node.EnvironmentID)
					}
					pr.Services = append(pr.Services, service{ID: s.Node.ID, Name: s.Node.Name, EnvIDs: envIDs})
				}
				projects = append(projects, pr)
			}
		}
		workspaces = append(workspaces, workspace{Name: mw.Name, TeamID: tid, Projects: projects})
	}

	if len(workspaces) == 0 {
		return fmt.Errorf("未找到任何工作区/项目")
	}

	// 选择工作区
	var chosenWS workspace
	// 通过 projectArg/ teamArg 预筛
	if projectArg != "" {
		for _, ws := range workspaces {
			for _, p := range ws.Projects {
				if eq(p.ID, projectArg) || eq(p.Name, projectArg) {
					chosenWS = ws
					break
				}
			}
			if chosenWS.Name != "" {
				break
			}
		}
		if chosenWS.Name == "" {
			// 未找到，转为交互选择
			names := collect(workspaces, func(w workspace) string { return w.Name })
			pick, err := util.PromptSelect("选择工作区", names)
			if err != nil {
				return err
			}
			for _, ws := range workspaces {
				if ws.Name == pick {
					chosenWS = ws
					break
				}
			}
		}
	} else if teamArg != "" {
		for _, ws := range workspaces {
			if (ws.TeamID != nil && eq(*ws.TeamID, teamArg)) || eq(ws.Name, teamArg) {
				chosenWS = ws
				break
			}
		}
		if chosenWS.Name == "" {
			return fmt.Errorf("未找到指定工作区: %s", teamArg)
		}
	} else {
		if len(workspaces) == 1 {
			chosenWS = workspaces[0]
			util.PrintInfo("选择工作区: " + chosenWS.Name)
		} else {
			names := collect(workspaces, func(w workspace) string { return w.Name })
			pick, err := util.PromptSelect("选择工作区", names)
			if err != nil {
				return err
			}
			for _, ws := range workspaces {
				if ws.Name == pick {
					chosenWS = ws
					break
				}
			}
		}
	}

	// 过滤已删除项目
	var availableProjects []project
	for _, p := range chosenWS.Projects {
		if !p.Deleted {
			availableProjects = append(availableProjects, p)
		}
	}
	if len(availableProjects) == 0 {
		return fmt.Errorf("该工作区下没有可用项目")
	}

	// 选择项目
	var chosenProject project
	if projectArg != "" {
		for _, p := range availableProjects {
			if eq(p.ID, projectArg) || eq(p.Name, projectArg) {
				chosenProject = p
				break
			}
		}
		if chosenProject.ID == "" {
			return fmt.Errorf("未在工作区 '%s' 找到项目 '%s'", chosenWS.Name, projectArg)
		}
		util.PrintInfo("选择项目: " + chosenProject.Name)
	} else {
		names := collect(availableProjects, func(p project) string { return p.Name })
		pick, err := util.PromptSelect("选择项目", names)
		if err != nil {
			return err
		}
		for _, p := range availableProjects {
			if p.Name == pick {
				chosenProject = p
				break
			}
		}
	}

	// 选择环境
	var chosenEnv env
	if envArg != "" {
		for _, e := range chosenProject.Envs {
			if eq(e.ID, envArg) || eq(e.Name, envArg) {
				chosenEnv = e
				break
			}
		}
		if chosenEnv.ID == "" {
			return fmt.Errorf("项目'%s'中未找到环境'%s'", chosenProject.Name, envArg)
		}
		util.PrintInfo("选择环境: " + chosenEnv.Name)
	} else if len(chosenProject.Envs) == 1 {
		chosenEnv = chosenProject.Envs[0]
		util.PrintInfo("选择环境: " + chosenEnv.Name)
	} else {
		names := collect(chosenProject.Envs, func(e env) string { return e.Name })
		pick, err := util.PromptSelect("选择环境", names)
		if err != nil {
			return err
		}
		for _, e := range chosenProject.Envs {
			if e.Name == pick {
				chosenEnv = e
				break
			}
		}
	}

	// 选择服务（可跳过）
	var chosenServiceID *string
	var candidateServices []service
	for _, s := range chosenProject.Services {
		for _, eid := range s.EnvIDs {
			if eid == chosenEnv.ID {
				candidateServices = append(candidateServices, s)
				break
			}
		}
	}
	if serviceArg != "" {
		for _, s := range candidateServices {
			if eq(s.ID, serviceArg) || eq(s.Name, serviceArg) {
				cs := s.ID
				chosenServiceID = &cs
				break
			}
		}
		if chosenServiceID == nil {
			return fmt.Errorf("环境'%s'可用服务中未找到 '%s'", chosenEnv.Name, serviceArg)
		}
		util.PrintInfo("选择服务: " + *chosenServiceID)
	} else if len(candidateServices) > 0 {
		names := append(collect(candidateServices, func(s service) string { return s.Name }), "<跳过>")
		pick, err := util.PromptSelect("选择服务 (或选择 <跳过>)", names)
		if err != nil {
			return err
		}
		if pick != "<跳过>" {
			for _, s := range candidateServices {
				if s.Name == pick {
					cs := s.ID
					chosenServiceID = &cs
					break
				}
			}
		}
	}

	// 写入链接配置
	if err := cfg.LinkProject(chosenProject.ID, chosenEnv.ID, &chosenProject.Name, &chosenEnv.Name); err != nil {
		return fmt.Errorf("链接项目失败: %w", err)
	}
	if chosenServiceID != nil {
		if err := cfg.LinkService(*chosenServiceID); err != nil {
			return fmt.Errorf("链接服务失败: %w", err)
		}
	}

	util.PrintSuccess(fmt.Sprintf("Project %s linked successfully! 🎉", chosenProject.Name))
	if err := cfg.Save(); err != nil {
		return err
	}
	return nil
}

func eq(a, b string) bool { return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b)) }

func collect[T any, R any](arr []T, f func(T) R) []R {
	out := make([]R, 0, len(arr))
	for _, v := range arr {
		out = append(out, f(v))
	}
	return out
}
