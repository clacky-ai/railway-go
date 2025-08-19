package railway

import (
	"context"
	"fmt"
	"strings"

	igql "github.com/railwayapp/cli/internal/gql"
)

// Environment 环境信息
type Environment struct {
	ID   string
	Name string
}

// Volume 卷信息
type Volume struct {
	ID        string
	Name      string
	CreatedAt string
	ProjectID string
}

// Project 项目信息（含环境、服务与卷的最小字段）
type Project struct {
	ID           string
	Name         string
	Environments []Environment
	Services     []Service
	Volumes      []Volume
}

// ProjectListItem 项目列表条目（轻量，仅保留必要字段）
type ProjectListItem struct {
	ID   string
	Name string
}

// GetProject 获取项目详情（携带环境、服务与卷核心字段）
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	var resp igql.ProjectResponse
	if err := c.gqlClient.Query(ctx, igql.ProjectQuery, map[string]any{"id": projectID}, &resp); err != nil {
		return nil, err
	}
	p := Project{ID: resp.Project.ID, Name: resp.Project.Name}
	for _, e := range resp.Project.Environments.Edges {
		p.Environments = append(p.Environments, Environment{ID: e.Node.ID, Name: e.Node.Name})
	}
	for _, s := range resp.Project.Services.Edges {
		p.Services = append(p.Services, Service{ID: s.Node.ID, Name: s.Node.Name})
	}
	for _, v := range resp.Project.Volumes.Edges {
		p.Volumes = append(p.Volumes, Volume{
			ID:        v.Node.ID,
			Name:      v.Node.Name,
			CreatedAt: v.Node.CreatedAt,
			ProjectID: v.Node.ProjectID,
		})
	}
	return &p, nil
}

// CreateProject 创建项目（返回包含环境列表的最小信息）
func (c *Client) CreateProject(ctx context.Context, name string, description *string, teamID *string) (*Project, error) {
	vars := map[string]any{"name": nullIfEmpty(name), "description": description, "teamId": teamID}
	var resp igql.ProjectCreateResponse
	if err := c.gqlClient.Mutate(ctx, igql.ProjectCreateMutation, vars, &resp); err != nil {
		return nil, err
	}
	p := Project{ID: resp.ProjectCreate.ID, Name: resp.ProjectCreate.Name}
	for _, e := range resp.ProjectCreate.Environments.Edges {
		p.Environments = append(p.Environments, Environment{ID: e.Node.ID, Name: e.Node.Name})
	}
	return &p, nil
}

// DeleteProject 删除项目
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	var resp igql.ProjectDeleteResponse
	if err := c.gqlClient.Mutate(ctx, igql.ProjectDeleteMutation, map[string]any{"id": projectID}, &resp); err != nil {
		return err
	}
	if !resp.ProjectDelete {
		return fmt.Errorf("project delete failed")
	}
	return nil
}

// CreateEnvironment 创建环境
func (c *Client) CreateEnvironment(ctx context.Context, projectID, name string) (*Environment, error) {
	input := igql.EnvironmentCreateInput{ProjectID: projectID, Name: name}
	var resp igql.EnvironmentCreateResponse
	if err := c.gqlClient.Mutate(ctx, igql.EnvironmentCreateMutation, map[string]any{"input": input}, &resp); err != nil {
		return nil, err
	}
	return &Environment{ID: resp.EnvironmentCreate.ID, Name: resp.EnvironmentCreate.Name}, nil
}

// ListProjects 列出当前可见的项目
func (c *Client) ListProjects(ctx context.Context) ([]ProjectListItem, error) {
	// 使用更稳定的 UserProjectsFullQuery 聚合生成项目列表，避免后端 schema 差异导致的失败
	var resp igql.UserProjectsFullResponse
	if err := c.gqlClient.Query(ctx, igql.UserProjectsFullQuery, nil, &resp); err != nil {
		return nil, err
	}
	type acc struct{ ID, Name string }
	m := map[string]acc{}
	// externalWorkspaces
	for _, ew := range resp.ExternalWorkspaces {
		for _, p := range ew.Projects {
			if p.DeletedAt != nil {
				continue
			}
			a := acc{ID: p.ID, Name: p.Name}
			m[a.ID] = a
		}
	}
	// me.workspaces.team.projects
	for _, mw := range resp.Me.Workspaces {
		if mw.Team == nil {
			continue
		}
		for _, edge := range mw.Team.Projects.Edges {
			n := edge.Node
			if n.DeletedAt != nil {
				continue
			}
			a := acc{ID: n.ID, Name: n.Name}
			m[a.ID] = a
		}
	}
	out := make([]ProjectListItem, 0, len(m))
	for _, v := range m {
		out = append(out, ProjectListItem{ID: v.ID, Name: v.Name})
	}
	return out, nil
}

// ListProjectsFull 参考 link.go，返回包含环境与服务实例环境引用的完整项目列表
// workspaceRef 可为空；非空时按工作区名称或团队ID过滤
func (c *Client) ListProjectsFull(ctx context.Context, workspaceRef string) ([]ProjectInfo, error) {
	var resp igql.UserProjectsFullResponse
	if err := c.gqlClient.Query(ctx, igql.UserProjectsFullQuery, nil, &resp); err != nil {
		return nil, err
	}

	// 选择工作区过滤器
	matchesWS := func(name string, teamID *string) bool {
		if strings.TrimSpace(workspaceRef) == "" {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(workspaceRef)) {
			return true
		}
		if teamID != nil && strings.EqualFold(strings.TrimSpace(*teamID), strings.TrimSpace(workspaceRef)) {
			return true
		}
		return false
	}

	// 汇总去重（按项目ID）
	projMap := map[string]*ProjectInfo{}

	// externalWorkspaces
	for _, ew := range resp.ExternalWorkspaces {
		if !matchesWS(ew.Name, ew.TeamID) {
			continue
		}
		for _, p := range ew.Projects {
			pi, ok := projMap[p.ID]
			if !ok {
				pi = &ProjectInfo{ID: p.ID, Name: p.Name}
				projMap[p.ID] = pi
			}
			// environments
			for _, e := range p.Environments.Edges {
				pi.Environments = append(pi.Environments, Environment{ID: e.Node.ID, Name: e.Node.Name})
			}
			// services
			for _, s := range p.Services.Edges {
				si := ServiceInfo{ID: s.Node.ID, Name: s.Node.Name}
				for _, inst := range s.Node.ServiceInstances.Edges {
					si.InstanceEnvironmentIDs = append(si.InstanceEnvironmentIDs, inst.Node.EnvironmentID)
				}
				pi.Services = append(pi.Services, si)
			}
		}
	}

	// me.workspaces.team.projects
	for _, mw := range resp.Me.Workspaces {
		var tid *string
		if mw.Team != nil {
			tid = &mw.Team.ID
		}
		if !matchesWS(mw.Name, tid) {
			continue
		}
		if mw.Team == nil {
			continue
		}
		for _, edge := range mw.Team.Projects.Edges {
			p := edge.Node
			pi, ok := projMap[p.ID]
			if !ok {
				pi = &ProjectInfo{ID: p.ID, Name: p.Name}
				projMap[p.ID] = pi
			}
			for _, e := range p.Environments.Edges {
				pi.Environments = append(pi.Environments, Environment{ID: e.Node.ID, Name: e.Node.Name})
			}
			for _, s := range p.Services.Edges {
				si := ServiceInfo{ID: s.Node.ID, Name: s.Node.Name}
				for _, inst := range s.Node.ServiceInstances.Edges {
					si.InstanceEnvironmentIDs = append(si.InstanceEnvironmentIDs, inst.Node.EnvironmentID)
				}
				pi.Services = append(pi.Services, si)
			}
		}
	}

	// 输出列表
	out := make([]ProjectInfo, 0, len(projMap))
	for _, v := range projMap {
		out = append(out, *v)
	}
	return out, nil
}
