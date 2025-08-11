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
	var raw map[string]any
	input := map[string]any{"name": name, "projectId": projectID, "environmentId": environmentID}
	query := "mutation($input:ProjectTokenCreateInput!){ projectTokenCreate(input:$input) }"
	if err := c.gqlClient.Mutate(ctx, query, map[string]any{"input": input}, &raw); err == nil {
		if t, ok := raw["projectTokenCreate"].(string); ok && t != "" {
			return t, nil
		}
	}
	// fallback
	raw = map[string]any{}
	q2 := "mutation($projectId:String!,$environmentId:String!,$name:String!){ projectTokenCreate(projectId:$projectId, environmentId:$environmentId, name:$name) }"
	if err := c.gqlClient.Mutate(ctx, q2, map[string]any{"projectId": projectID, "environmentId": environmentID, "name": name}, &raw); err != nil {
		return "", err
	}
	if t, ok := raw["projectTokenCreate"].(string); ok && t != "" {
		return t, nil
	}
	return "", fmt.Errorf("project token create failed")
}

// 底层 raw：DeleteProjectToken
func (c *Client) deleteProjectTokenRaw(ctx context.Context, tokenID string) error {
	type delResp struct {
		ProjectTokenDelete bool `json:"projectTokenDelete"`
	}
	var r delResp
	q := "mutation($id:String!){ projectTokenDelete(id:$id) }"
	if err := c.gqlClient.Mutate(ctx, q, map[string]any{"id": tokenID}, &r); err == nil {
		if r.ProjectTokenDelete {
			return nil
		}
	}
	r = delResp{}
	q2 := "mutation($input:ProjectTokenDeleteInput!){ projectTokenDelete(input:$input) }"
	if err := c.gqlClient.Mutate(ctx, q2, map[string]any{"input": map[string]any{"id": tokenID}}, &r); err != nil {
		return err
	}
	if !r.ProjectTokenDelete {
		return fmt.Errorf("project token delete failed")
	}
	return nil
}

// 底层 raw：ListProjectTokens
func (c *Client) listProjectTokensRaw(ctx context.Context, projectID string) ([]ProjectToken, error) {
	query := "query($projectId:String!,$after:String){ projectTokens(projectId:$projectId, first:50, after:$after) { edges { cursor node { id name project { id name } environment { id name } } } pageInfo { hasNextPage endCursor } } }"
	type tokenNode struct {
		ID          string                    `json:"id"`
		Name        string                    `json:"name"`
		Project     struct{ ID, Name string } `json:"project"`
		Environment struct{ ID, Name string } `json:"environment"`
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

	after := ""
	var out []ProjectToken
	for {
		vars := map[string]any{"projectId": projectID, "after": nullIfEmpty(after)}
		var r resp
		if err := c.gqlClient.Query(ctx, query, vars, &r); err != nil {
			return nil, err
		}
		for _, e := range r.ProjectTokens.Edges {
			out = append(out, ProjectToken{ID: e.Node.ID, Name: e.Node.Name, ProjectID: e.Node.Project.ID, ProjectName: e.Node.Project.Name, EnvironmentID: e.Node.Environment.ID, EnvironmentName: e.Node.Environment.Name})
		}
		if !r.ProjectTokens.PageInfo.HasNextPage || r.ProjectTokens.PageInfo.EndCursor == nil || *r.ProjectTokens.PageInfo.EndCursor == "" {
			break
		}
		after = *r.ProjectTokens.PageInfo.EndCursor
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
