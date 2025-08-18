package railway

import (
	"context"

	igql "github.com/railwayapp/cli/internal/gql"
)

// GetWorkflowStatus 获取 workflow 状态
func (c *Client) GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowInfo, error) {
	var resp igql.WorkflowStatusResponse
	if err := c.gqlClient.Query(ctx, igql.WorkflowStatusQuery, map[string]any{
		"workflowId": workflowID,
	}, &resp); err != nil {
		return nil, err
	}

	workflowStatus := resp.WorkflowStatus

	return &WorkflowInfo{
		ID:     "", // workflowStatus 查询不返回 ID
		Status: workflowStatus.Status,
		Error:  workflowStatus.Error,
	}, nil
}
