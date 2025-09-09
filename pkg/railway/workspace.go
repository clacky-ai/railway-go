package railway

import (
	"context"

	igql "github.com/railwayapp/cli/internal/gql"
)

// Workspace 工作区（兼容 externalWorkspaces 与 me.workspaces）
type Workspace struct {
	ID     string
	Name   string
	TeamID string
}

// ServiceEnvironmentRef 服务在某环境的引用
type ServiceEnvironmentRef struct {
	ServiceID     string
	ServiceName   string
	EnvironmentID string
}

// ProjectSummary 项目概览（用于工作区完整列表）
type ProjectSummary struct {
	ID           string
	Name         string
	DeletedAt    *string
	Environments []Environment
	Services     []ServiceEnvironmentRef
}

// WorkspaceWithProjects 带项目的工作区
type WorkspaceWithProjects struct {
	ID       string
	Name     string
	TeamID   *string
	Projects []ProjectSummary
}

// ListWorkspaces 列出工作区（合并 externalWorkspaces 与 me.workspaces）
func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	var resp igql.UserProjectsResponse
	if err := c.gqlClient.Query(ctx, igql.UserProjectsQuery, nil, &resp); err != nil {
		return nil, err
	}
	// 使用 map 去重
	m := map[string]Workspace{}
	for _, ew := range resp.ExternalWorkspaces {
		w := Workspace{ID: ew.ID, Name: ew.Name, TeamID: ew.TeamID}
		m[w.ID] = w
	}
	for _, mw := range resp.Me.Workspaces {
		var teamID string
		if mw.Team != nil {
			id := mw.Team.ID
			teamID = id
		}
		w := Workspace{ID: mw.ID, Name: mw.Name, TeamID: teamID}
		m[w.ID] = w
	}
	out := make([]Workspace, 0, len(m))
	for _, w := range m {
		out = append(out, w)
	}
	return out, nil
}

// ListWorkspacesWithProjects 列出包含项目详情的工作区
func (c *Client) ListWorkspacesWithProjects(ctx context.Context) ([]WorkspaceWithProjects, error) {
	var resp igql.UserProjectsFullResponse
	if err := c.gqlClient.Query(ctx, igql.UserProjectsFullQuery, nil, &resp); err != nil {
		return nil, err
	}
	// 聚合 externalWorkspaces
	m := map[string]WorkspaceWithProjects{}
	for _, ew := range resp.ExternalWorkspaces {
		ww := WorkspaceWithProjects{ID: ew.ID, Name: ew.Name, TeamID: ew.TeamID}
		for _, p := range ew.Projects {
			ps := ProjectSummary{ID: p.ID, Name: p.Name, DeletedAt: p.DeletedAt}
			for _, e := range p.Environments.Edges {
				ps.Environments = append(ps.Environments, Environment{ID: e.Node.ID, Name: e.Node.Name})
			}
			for _, s := range p.Services.Edges {
				for _, inst := range s.Node.ServiceInstances.Edges {
					ps.Services = append(ps.Services, ServiceEnvironmentRef{ServiceID: s.Node.ID, ServiceName: s.Node.Name, EnvironmentID: inst.Node.EnvironmentID})
				}
			}
			ww.Projects = append(ww.Projects, ps)
		}
		m[ww.ID] = ww
	}
	// 聚合 me.workspaces.team.projects
	for _, mw := range resp.Me.Workspaces {
		id := mw.ID
		w, ok := m[id]
		if !ok {
			w = WorkspaceWithProjects{ID: mw.ID, Name: mw.Name}
		}
		if mw.Team != nil {
			tid := mw.Team.ID
			w.TeamID = &tid
			for _, edge := range mw.Team.Projects.Edges {
				p := edge.Node
				ps := ProjectSummary{ID: p.ID, Name: p.Name, DeletedAt: p.DeletedAt}
				for _, e := range p.Environments.Edges {
					ps.Environments = append(ps.Environments, Environment{ID: e.Node.ID, Name: e.Node.Name})
				}
				for _, s := range p.Services.Edges {
					for _, inst := range s.Node.ServiceInstances.Edges {
						ps.Services = append(ps.Services, ServiceEnvironmentRef{ServiceID: s.Node.ID, ServiceName: s.Node.Name, EnvironmentID: inst.Node.EnvironmentID})
					}
				}
				// 去重追加
				exists := false
				for _, e := range w.Projects {
					if e.ID == ps.ID {
						exists = true
						break
					}
				}
				if !exists {
					w.Projects = append(w.Projects, ps)
				}
			}
		}
		m[w.ID] = w
	}
	out := make([]WorkspaceWithProjects, 0, len(m))
	for _, w := range m {
		out = append(out, w)
	}
	return out, nil
}
