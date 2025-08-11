package railway

import (
	"context"
	"fmt"

	igql "github.com/railwayapp/cli/internal/gql"
)

// GraphQLQuery 直接执行原始 GraphQL 查询（高级用法）
func (c *Client) GraphQLQuery(ctx context.Context, query string, variables map[string]any, out interface{}) error {
	return c.gqlClient.Query(ctx, query, variables, out)
}

// GraphQLMutate 直接执行原始 GraphQL 变更（高级用法）
func (c *Client) GraphQLMutate(ctx context.Context, mutation string, variables map[string]any, out interface{}) error {
	return c.gqlClient.Mutate(ctx, mutation, variables, out)
}

// 底层 raw：CreateProjectToken
func (c *Client) createProjectTokenRaw(ctx context.Context, projectID, environmentID, name string) (string, error) {
	var resp igql.ProjectTokenCreateResponse
	input := igql.ProjectTokenCreateInput{Name: name, ProjectID: projectID, EnvironmentID: environmentID}
	if err := c.gqlClient.Mutate(ctx, igql.ProjectTokenCreateMutation, map[string]any{"input": input}, &resp); err == nil {
		if resp.ProjectTokenCreate != "" {
			return resp.ProjectTokenCreate, nil
		}
	}
	// fallback
	var resp2 igql.ProjectTokenCreateResponse
	if err := c.gqlClient.Mutate(ctx, igql.ProjectTokenCreateByParamsMutation, map[string]any{"projectId": projectID, "environmentId": environmentID, "name": name}, &resp2); err != nil {
		return "", err
	}
	if resp2.ProjectTokenCreate != "" {
		return resp2.ProjectTokenCreate, nil
	}
	return "", fmt.Errorf("project token create failed")
}

// 底层 raw：DeleteProjectToken
func (c *Client) deleteProjectTokenRaw(ctx context.Context, tokenID string) error {
	var resp igql.ProjectTokenDeleteResponse
	if err := c.gqlClient.Mutate(ctx, igql.ProjectTokenDeleteMutation, map[string]any{"id": tokenID}, &resp); err == nil {
		if resp.ProjectTokenDelete {
			return nil
		}
	}
	var resp2 igql.ProjectTokenDeleteResponse
	input := igql.ProjectTokenDeleteInput{ID: tokenID}
	if err := c.gqlClient.Mutate(ctx, igql.ProjectTokenDeleteByInputMutation, map[string]any{"input": input}, &resp2); err != nil {
		return err
	}
	if !resp2.ProjectTokenDelete {
		return fmt.Errorf("project token delete failed")
	}
	return nil
}

// 底层 raw：ListProjectTokens
func (c *Client) listProjectTokensRaw(ctx context.Context, projectID string) ([]ProjectToken, error) {
	after := ""
	var out []ProjectToken
	for {
		vars := map[string]any{"projectId": projectID, "after": nullIfEmpty(after)}
		var resp igql.ProjectTokensResponse
		if err := c.gqlClient.Query(ctx, igql.ProjectTokensQuery, vars, &resp); err != nil {
			return nil, err
		}
		for _, e := range resp.ProjectTokens.Edges {
			out = append(out, ProjectToken{ID: e.Node.ID, Name: e.Node.Name, ProjectID: e.Node.Project.ID, ProjectName: e.Node.Project.Name, EnvironmentID: e.Node.Environment.ID, EnvironmentName: e.Node.Environment.Name})
		}
		if !resp.ProjectTokens.PageInfo.HasNextPage || resp.ProjectTokens.PageInfo.EndCursor == nil || *resp.ProjectTokens.PageInfo.EndCursor == "" {
			break
		}
		after = *resp.ProjectTokens.PageInfo.EndCursor
	}
	return out, nil
}

// 底层 raw：CurrentProjectFromToken
func (c *Client) currentProjectFromTokenRaw(ctx context.Context) (*ProjectTokenContext, error) {
	var resp igql.ProjectTokenResponse
	if err := c.gqlClient.Query(ctx, igql.ProjectTokenQuery, nil, &resp); err != nil {
		return nil, err
	}
	return &ProjectTokenContext{ProjectID: resp.ProjectToken.Project.ID, ProjectName: resp.ProjectToken.Project.Name, EnvironmentID: resp.ProjectToken.Environment.ID, EnvironmentName: resp.ProjectToken.Environment.Name}, nil
}
