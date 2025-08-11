package railway

import (
	"context"
	"fmt"
)

// ProjectToken 项目访问令牌
type ProjectToken struct {
	ID              string
	Name            string
	ProjectID       string
	ProjectName     string
	EnvironmentID   string
	EnvironmentName string
}

// ProjectTokenContext 通过项目令牌解析到的上下文
type ProjectTokenContext struct {
	ProjectID       string
	ProjectName     string
	EnvironmentID   string
	EnvironmentName string
}

// CreateProjectToken 创建项目访问令牌
func (c *Client) CreateProjectToken(ctx context.Context, projectID, environmentID, name string) (string, error) {
	// 转调 client.go 中实现（保留原逻辑）
	return c.createProjectTokenRaw(ctx, projectID, environmentID, name)
}

// DeleteProjectToken 删除项目访问令牌
func (c *Client) DeleteProjectToken(ctx context.Context, tokenID string) error {
	return c.deleteProjectTokenRaw(ctx, tokenID)
}

// ListProjectTokens 列出项目访问令牌
func (c *Client) ListProjectTokens(ctx context.Context, projectID string) ([]ProjectToken, error) {
	return c.listProjectTokensRaw(ctx, projectID)
}

// CurrentProjectFromToken 从当前 token 解析项目/环境上下文
func (c *Client) CurrentProjectFromToken(ctx context.Context) (*ProjectTokenContext, error) {
	// 保留在 client.go 里的具体查询实现，避免重复 schema 绑定
	var ptc *ProjectTokenContext
	// 调用一次底层方法，若失败直接返回
	var err error
	ptc, err = c.currentProjectFromTokenRaw(ctx)
	if err != nil {
		return nil, err
	}
	if ptc.ProjectID == "" || ptc.EnvironmentID == "" {
		return nil, fmt.Errorf("invalid project token context")
	}
	return ptc, nil
}
