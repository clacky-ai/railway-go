package railway

import (
	"context"
	"errors"
	"fmt"
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

// Down 尝试停止服务实例；失败时退而求其次停止最新部署
func (c *Client) Down(ctx context.Context, projectID, environmentID, serviceID string) error {
	// 首选停止服务实例
	if err := c.StopServiceInstance(ctx, serviceID, environmentID); err == nil {
		return nil
	}

	// 列出部署并尝试停止最新一个
	deps, err := c.ListDeployments(ctx, projectID, environmentID, &serviceID)
	if err != nil {
		return err
	}
	if len(deps) == 0 {
		return nil
	}
	// 假设列表顺序为新->旧（若非如此，可做一次反向或选择最新 status 非终态项）
	return c.StopDeployment(ctx, deps[0].ID)
}
