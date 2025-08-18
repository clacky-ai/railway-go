package railway

import (
	"context"
	"strings"

	igql "github.com/railwayapp/cli/internal/gql"
)

// Deployment 部署信息（最小字段）
type Deployment struct {
	ID      string
	Status  string
	URL     *string
	Service Service
}

// ListDeployments 列出部署
func (c *Client) ListDeployments(ctx context.Context, projectID, environmentID string, serviceID *string) ([]Deployment, error) {
	vars := map[string]any{"projectId": projectID, "environmentId": environmentID, "serviceId": nil}
	if serviceID != nil && strings.TrimSpace(*serviceID) != "" {
		vars["serviceId"] = *serviceID
	}
	var resp igql.DeploymentsResponse
	if err := c.gqlClient.Query(ctx, igql.DeploymentsQuery, vars, &resp); err != nil {
		return nil, err
	}
	out := make([]Deployment, 0, len(resp.Deployments.Edges))
	for _, e := range resp.Deployments.Edges {
		out = append(out, Deployment{ID: e.Node.ID, Status: e.Node.Status, URL: e.Node.URL, Service: Service{ID: e.Node.Service.ID, Name: e.Node.Service.Name}})
	}
	return out, nil
}

// DeployServiceInstance 触发服务实例部署
func (c *Client) DeployServiceInstance(ctx context.Context, serviceID, environmentID string) (deploymentID, status string, err error) {
	input := igql.ServiceInstanceDeployInput{ServiceID: serviceID, EnvironmentID: environmentID}
	var resp igql.ServiceInstanceDeployResponse
	if err := c.gqlClient.Mutate(ctx, igql.ServiceInstanceDeployMutation, map[string]any{"input": input}, &resp); err != nil {
		return "", "", err
	}
	return resp.ServiceInstanceDeploy.ID, resp.ServiceInstanceDeploy.Status, nil
}

// RedeployDeployment 重新部署指定部署
func (c *Client) RedeployDeployment(ctx context.Context, deploymentID string) (id, status string, err error) {
	var resp igql.DeploymentRedeployResponse
	if err := c.gqlClient.Mutate(ctx, igql.DeploymentRedeployMutation, map[string]any{"id": deploymentID}, &resp); err != nil {
		return "", "", err
	}
	return resp.DeploymentRedeploy.ID, resp.DeploymentRedeploy.Status, nil
}

// StopDeploymentSimple 停止部署（简单版本，返回布尔值）
func (c *Client) StopDeploymentSimple(ctx context.Context, deploymentID string) (bool, error) {
	var resp igql.DeploymentStopSimpleResponse
	if err := c.gqlClient.Mutate(ctx, igql.DeploymentStopSimpleMutation, map[string]any{"id": deploymentID}, &resp); err != nil {
		return false, err
	}
	return resp.DeploymentStop, nil
}

// CancelDeployment 取消部署
func (c *Client) CancelDeployment(ctx context.Context, deploymentID string) (bool, error) {
	var resp igql.DeploymentCancelResponse
	if err := c.gqlClient.Mutate(ctx, igql.DeploymentCancelMutation, map[string]any{"id": deploymentID}, &resp); err != nil {
		return false, err
	}
	return resp.DeploymentCancel, nil
}

// AbortDeployment 中止部署
func (c *Client) AbortDeployment(ctx context.Context, deploymentID string) (bool, error) {
	var resp igql.DeploymentAbortResponse
	if err := c.gqlClient.Mutate(ctx, igql.DeploymentAbortMutation, map[string]any{"id": deploymentID}, &resp); err != nil {
		return false, err
	}
	return resp.DeploymentAbort, nil
}

// RemoveDeployment 删除部署
func (c *Client) RemoveDeployment(ctx context.Context, deploymentID string) error {
	return c.gqlClient.Mutate(ctx, igql.DeploymentRemoveMutation, map[string]any{"id": deploymentID}, nil)
}

// ScaleServiceInstance 缩放服务实例
func (c *Client) ScaleServiceInstance(ctx context.Context, serviceID, environmentID string, replicas int) (bool, error) {
	input := igql.ServiceInstanceScaleInput{ServiceID: serviceID, EnvironmentID: environmentID, Replicas: replicas}
	var resp igql.ServiceInstanceScaleResponse
	if err := c.gqlClient.Mutate(ctx, igql.ServiceInstanceScaleMutation, map[string]any{"input": input}, &resp); err != nil {
		return false, err
	}
	return resp.ServiceInstanceScale, nil
}
