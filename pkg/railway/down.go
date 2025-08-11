package railway

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	igql "github.com/railwayapp/cli/internal/gql"
)

// StopServiceInstance 尝试停止指定环境下的服务实例（多种 GraphQL 变体容错）
func (c *Client) StopServiceInstance(ctx context.Context, serviceID, environmentID string) error {
	// 优先 input 形式：serviceInstanceStop
	var raw map[string]any
	q1 := "mutation($input:ServiceInstanceStopInput!){ serviceInstanceStop(input:$input) }"
	if err := c.gqlClient.Mutate(ctx, q1, map[string]any{"input": map[string]any{"serviceId": serviceID, "environmentId": environmentID}}, &raw); err == nil {
		if v, ok := raw["serviceInstanceStop"].(bool); ok && v {
			return nil
		}
	}

	// 退路：缩放为 0 副本
	raw = map[string]any{}
	q2 := "mutation($input:ServiceInstanceScaleInput!){ serviceInstanceScale(input:$input) }"
	if err := c.gqlClient.Mutate(ctx, q2, map[string]any{"input": map[string]any{"serviceId": serviceID, "environmentId": environmentID, "replicas": 0}}, &raw); err == nil {
		if v, ok := raw["serviceInstanceScale"].(bool); ok && v {
			return nil
		}
	}

	// 退路：参数式 stop
	raw = map[string]any{}
	q3 := "mutation($serviceId:String!,$environmentId:String!){ serviceInstanceStop(serviceId:$serviceId, environmentId:$environmentId) }"
	if err := c.gqlClient.Mutate(ctx, q3, map[string]any{"serviceId": serviceID, "environmentId": environmentID}, &raw); err == nil {
		if v, ok := raw["serviceInstanceStop"].(bool); ok && v {
			return nil
		}
	}

	// 退路：参数式 scale=0
	raw = map[string]any{}
	q4 := "mutation($serviceId:String!,$environmentId:String!){ serviceInstanceScale(serviceId:$serviceId, environmentId:$environmentId, replicas:0) }"
	if err := c.gqlClient.Mutate(ctx, q4, map[string]any{"serviceId": serviceID, "environmentId": environmentID}, &raw); err == nil {
		if v, ok := raw["serviceInstanceScale"].(bool); ok && v {
			return nil
		}
	}

	return errors.New("stop service instance not supported by backend")
}

// StopDeployment 尝试停止/取消指定部署（多种 GraphQL 变体容错）
func (c *Client) StopDeployment(ctx context.Context, deploymentID string) error {
	// 尝试返回对象形式
	var obj map[string]any
	q1 := "mutation($id:String!){ deploymentStop(id:$id) { id status deploymentStopped } }"
	if err := c.gqlClient.Mutate(ctx, q1, map[string]any{"id": deploymentID}, &obj); err == nil {
		if v, ok := obj["deploymentStop"].(map[string]any); ok {
			_ = v
			return nil
		}
	}

	// 布尔返回
	var raw map[string]any
	q2 := "mutation($id:String!){ deploymentStop(id:$id) }"
	if err := c.gqlClient.Mutate(ctx, q2, map[string]any{"id": deploymentID}, &raw); err == nil {
		if v, ok := raw["deploymentStop"].(bool); ok && v {
			return nil
		}
	}

	// 兼容其他命名
	raw = map[string]any{}
	q3 := "mutation($id:String!){ deploymentCancel(id:$id) }"
	if err := c.gqlClient.Mutate(ctx, q3, map[string]any{"id": deploymentID}, &raw); err == nil {
		if v, ok := raw["deploymentCancel"].(bool); ok && v {
			return nil
		}
	}

	raw = map[string]any{}
	q4 := "mutation($id:String!){ deploymentAbort(id:$id) }"
	if err := c.gqlClient.Mutate(ctx, q4, map[string]any{"id": deploymentID}, &raw); err == nil {
		if v, ok := raw["deploymentAbort"].(bool); ok && v {
			return nil
		}
	}

	return fmt.Errorf("failed to stop deployment %s: backend not supported", deploymentID)
}

// DeleteDeployment 删除部署（对齐 CLI down 子命令行为）
func (c *Client) DeleteDeployment(ctx context.Context, deploymentID string) error {
	var resp struct {
		DeploymentRemove bool `json:"deploymentRemove"`
	}
	if err := c.gqlClient.Mutate(ctx, igql.DeploymentRemoveMutation, map[string]any{"id": deploymentID}, &resp); err != nil {
		return err
	}
	if !resp.DeploymentRemove {
		return fmt.Errorf("backend refused deployment removal")
	}
	return nil
}

// Down 删除最近一次成功的部署（参考 internal/commands/down.go 行为）
func (c *Client) Down(ctx context.Context, projectID, environmentID, serviceID string) error {
	// 查询部署（需要 createdAt 字段便于排序）
	var deps igql.DeploymentsResponse
	vars := map[string]any{
		"projectId":     projectID,
		"environmentId": environmentID,
		"serviceId":     serviceID,
	}
	if err := c.gqlClient.Query(ctx, igql.DeploymentsQuery, vars, &deps); err != nil {
		return err
	}

	type depNode struct{ ID, Status, CreatedAt string }
	nodes := make([]depNode, 0, len(deps.Deployments.Edges))
	for _, ed := range deps.Deployments.Edges {
		if strings.EqualFold(ed.Node.Status, "SUCCESS") {
			nodes = append(nodes, depNode{ID: ed.Node.ID, Status: ed.Node.Status, CreatedAt: ed.Node.CreatedAt})
		}
	}
	if len(nodes) == 0 {
		return fmt.Errorf("no successful deployment found")
	}

	parse := func(s string) (time.Time, bool) {
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return t, true
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t, true
		}
		layouts := []string{
			"2006-01-02T15:04:05.000Z07:00",
			time.DateTime,
			time.RFC1123Z,
		}
		for _, l := range layouts {
			if t, err := time.Parse(l, s); err == nil {
				return t, true
			}
		}
		return time.Time{}, false
	}
	sort.Slice(nodes, func(i, j int) bool {
		ti, okI := parse(nodes[i].CreatedAt)
		tj, okJ := parse(nodes[j].CreatedAt)
		if okI && okJ {
			return ti.After(tj)
		}
		return nodes[i].CreatedAt > nodes[j].CreatedAt
	})

	// 删除最新
	return c.DeleteDeployment(ctx, nodes[0].ID)
}
