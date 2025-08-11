package railway

import (
	"context"
	"fmt"

	igql "github.com/railwayapp/cli/internal/gql"
)

// GetVariables 拉取服务在指定环境下的变量
func (c *Client) GetVariables(ctx context.Context, projectID, environmentID, serviceID string) (map[string]string, error) {
	var raw map[string]any
	if err := c.gqlClient.Query(ctx, igql.VariablesForServiceDeploymentQuery, map[string]any{
		"projectId":     projectID,
		"environmentId": environmentID,
		"serviceId":     serviceID,
	}, &raw); err != nil {
		return nil, err
	}
	// 兼容不同根键
	var varsMap map[string]any
	if v, ok := raw["variables"]; ok {
		varsMap, _ = v.(map[string]any)
	}
	if varsMap == nil {
		if v, ok := raw["variablesForServiceDeployment"]; ok {
			varsMap, _ = v.(map[string]any)
		}
	}
	if varsMap == nil {
		return nil, fmt.Errorf("unsupported response shape")
	}
	out := make(map[string]string)
	for k, v := range varsMap {
		if s, ok := v.(string); ok && s != "" {
			out[k] = s
		}
	}
	return out, nil
}

// SetVariables 设置变量（不触发替换全部，仅增量 upsert）
func (c *Client) SetVariables(ctx context.Context, projectID, environmentID, serviceID string, vars map[string]string) error {
	m := make(map[string]*string, len(vars))
	for k, v := range vars {
		vv := v
		m[k] = &vv
	}
	input := igql.VariableCollectionUpsertInput{ProjectID: projectID, EnvironmentID: environmentID, ServiceID: &serviceID, Variables: m}
	var resp igql.VariableCollectionUpsertResponse
	if err := c.gqlClient.Mutate(ctx, igql.VariableCollectionUpsertMutation, map[string]any{"input": input}, &resp); err != nil {
		return err
	}
	if !resp.VariableCollectionUpsert {
		return fmt.Errorf("variable upsert failed")
	}
	return nil
}

// UpsertVariables 支持 replace 语义（当 replace=true，后端将覆盖替换集合）
func (c *Client) UpsertVariables(ctx context.Context, projectID, environmentID string, serviceID *string, replace bool, vars map[string]string) error {
	m := make(map[string]*string, len(vars))
	for k, v := range vars {
		vv := v
		m[k] = &vv
	}
	input := igql.VariableCollectionUpsertInput{ProjectID: projectID, EnvironmentID: environmentID, ServiceID: serviceID, Replace: &replace, Variables: m}
	var resp igql.VariableCollectionUpsertResponse
	if err := c.gqlClient.Mutate(ctx, igql.VariableCollectionUpsertMutation, map[string]any{"input": input}, &resp); err != nil {
		return err
	}
	if !resp.VariableCollectionUpsert {
		return fmt.Errorf("variable upsert failed")
	}
	return nil
}
