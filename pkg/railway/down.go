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
	var resp igql.ServiceInstanceStopResponse
	if err := c.gqlClient.Mutate(ctx, igql.ServiceInstanceStopMutation, map[string]any{"input": igql.ServiceInstanceStopInput{ServiceID: serviceID, EnvironmentID: environmentID}}, &resp); err == nil {
		if resp.ServiceInstanceStop {
			return nil
		}
	}

	// 退路：缩放为 0 副本
	var scaleResp igql.ServiceInstanceScaleResponse
	if err := c.gqlClient.Mutate(ctx, igql.ServiceInstanceScaleMutation, map[string]any{"input": igql.ServiceInstanceScaleInput{ServiceID: serviceID, EnvironmentID: environmentID, Replicas: 0}}, &scaleResp); err == nil {
		if scaleResp.ServiceInstanceScale {
			return nil
		}
	}

	// 退路：参数式 stop
	var paramResp igql.ServiceInstanceStopResponse
	if err := c.gqlClient.Mutate(ctx, igql.ServiceInstanceStopByParamsMutation, map[string]any{"serviceId": serviceID, "environmentId": environmentID}, &paramResp); err == nil {
		if paramResp.ServiceInstanceStop {
			return nil
		}
	}

	// 退路：参数式 scale=0
	var paramScaleResp igql.ServiceInstanceScaleResponse
	if err := c.gqlClient.Mutate(ctx, igql.ServiceInstanceScaleByParamsMutation, map[string]any{"serviceId": serviceID, "environmentId": environmentID}, &paramScaleResp); err == nil {
		if paramScaleResp.ServiceInstanceScale {
			return nil
		}
	}

	return errors.New("stop service instance not supported by backend")
}

// StopDeployment 尝试停止/取消指定部署（多种 GraphQL 变体容错）
func (c *Client) StopDeployment(ctx context.Context, deploymentID string) error {
	// 尝试返回对象形式
	var resp igql.DeploymentStopResponse
	if err := c.gqlClient.Mutate(ctx, igql.DeploymentStopMutation, map[string]any{"id": deploymentID}, &resp); err == nil {
		if resp.DeploymentStop.ID != "" {
			return nil
		}
	}

	// 布尔返回
	var simpleResp igql.DeploymentStopSimpleResponse
	if err := c.gqlClient.Mutate(ctx, igql.DeploymentStopSimpleMutation, map[string]any{"id": deploymentID}, &simpleResp); err == nil {
		if simpleResp.DeploymentStop {
			return nil
		}
	}

	// 兼容其他命名
	var cancelResp igql.DeploymentCancelResponse
	if err := c.gqlClient.Mutate(ctx, igql.DeploymentCancelMutation, map[string]any{"id": deploymentID}, &cancelResp); err == nil {
		if cancelResp.DeploymentCancel {
			return nil
		}
	}

	var abortResp igql.DeploymentAbortResponse
	if err := c.gqlClient.Mutate(ctx, igql.DeploymentAbortMutation, map[string]any{"id": deploymentID}, &abortResp); err == nil {
		if abortResp.DeploymentAbort {
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
