package railway

import (
	"context"

	igql "github.com/railwayapp/cli/internal/gql"
)

// TemplateDeployResult 模板部署结果
type TemplateDeployResult struct {
	ProjectID  string
	WorkflowID string
}

// DeployTemplate 基于模板进行部署
func (c *Client) DeployTemplate(ctx context.Context, projectID, environmentID, templateID string, serializedConfig igql.SerializedTemplateConfig) (*TemplateDeployResult, error) {
	var resp igql.TemplateDeployResponse
	if err := c.gqlClient.Mutate(ctx, igql.TemplateDeployMutation, map[string]any{
		"projectId":        projectID,
		"environmentId":    environmentID,
		"templateId":       templateID,
		"serializedConfig": serializedConfig,
	}, &resp); err != nil {
		return nil, err
	}
	return &TemplateDeployResult{
		ProjectID:  resp.TemplateDeployV2.ProjectID,
		WorkflowID: resp.TemplateDeployV2.WorkflowID,
	}, nil
}
