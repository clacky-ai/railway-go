package railway

import (
	"context"

	igql "github.com/railwayapp/cli/internal/gql"
)

// TemplateDeployResult 模板部署结果
type TemplateDeployResult struct {
	ID   string
	Name string
}

// DeployTemplate 基于模板进行部署
func (c *Client) DeployTemplate(ctx context.Context, projectID, environmentID, templateID string, serializedConfig any) (*TemplateDeployResult, error) {
	input := igql.TemplateDeployInput{ProjectID: projectID, EnvironmentID: environmentID, TemplateID: templateID, SerializedConfig: serializedConfig}
	var resp igql.TemplateDeployResponse
	if err := c.gqlClient.Mutate(ctx, igql.TemplateDeployMutation, map[string]any{"input": input}, &resp); err != nil {
		return nil, err
	}
	return &TemplateDeployResult{ID: resp.TemplateDeploy.ID, Name: resp.TemplateDeploy.Name}, nil
}
