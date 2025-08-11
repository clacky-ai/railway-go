package railway

import (
	"context"
	"fmt"

	igql "github.com/railwayapp/cli/internal/gql"
)

// Environment 环境信息
type Environment struct {
	ID   string
	Name string
}

// Project 项目信息（含环境与服务的最小字段）
type Project struct {
	ID           string
	Name         string
	Environments []Environment
	Services     []Service
}

// GetProject 获取项目详情（携带环境与服务核心字段）
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
