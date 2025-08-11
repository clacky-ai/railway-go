package railway

import (
	"context"
	"strings"

	igql "github.com/railwayapp/cli/internal/gql"
)

// Deployment 部署信息（最小字段）
type Deployment struct {
	ID     string
	Status string
	URL    *string
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
		out = append(out, Deployment{ID: e.Node.ID, Status: e.Node.Status, URL: e.Node.URL})
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
