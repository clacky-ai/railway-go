package railway

import (
	"context"
	"fmt"
	"strings"

	igql "github.com/railwayapp/cli/internal/gql"
)

// ResolveEnvironmentID 根据 ID 或名称解析环境 ID
func (c *Client) ResolveEnvironmentID(ctx context.Context, projectID, environmentRef string) (string, error) {
	if strings.TrimSpace(environmentRef) == "" {
		return "", fmt.Errorf("environment ref is empty")
	}
	var proj igql.ProjectResponse
	if err := c.gqlClient.Query(ctx, igql.ProjectQuery, map[string]any{"id": projectID}, &proj); err != nil {
		return "", err
	}
	for _, e := range proj.Project.Environments.Edges {
		if environmentRef == e.Node.ID || environmentRef == e.Node.Name {
			return e.Node.ID, nil
		}
	}
	return "", fmt.Errorf("environment not found: %s", environmentRef)
}

// ResolveServiceID 根据 ID 或名称解析服务 ID
func (c *Client) ResolveServiceID(ctx context.Context, projectID, serviceRef string) (string, error) {
	if strings.TrimSpace(serviceRef) == "" {
		return "", fmt.Errorf("service ref is empty")
	}
	var proj igql.ProjectResponse
	if err := c.gqlClient.Query(ctx, igql.ProjectQuery, map[string]any{"id": projectID}, &proj); err != nil {
		return "", err
	}
	for _, s := range proj.Project.Services.Edges {
		if serviceRef == s.Node.ID || serviceRef == s.Node.Name {
			return s.Node.ID, nil
		}
	}
	return "", fmt.Errorf("service not found: %s", serviceRef)
}
