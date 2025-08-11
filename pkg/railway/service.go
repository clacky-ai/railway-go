package railway

import (
	"context"
	"fmt"
	"strings"

	igql "github.com/railwayapp/cli/internal/gql"
)

// Service 服务信息
type Service struct {
	ID   string
	Name string
}

// ServiceInEnvironment 展示服务在某环境是否有实例
type ServiceInEnvironment struct {
	Service
	HasInstance bool
}

// CreateService 创建服务
func (c *Client) CreateService(ctx context.Context, projectID, name string) (*Service, error) {
	input := igql.ServiceCreateInput{ProjectID: projectID, Name: name}
	var resp igql.ServiceCreateResponse
	if err := c.gqlClient.Mutate(ctx, igql.ServiceCreateMutation, map[string]any{"input": input}, &resp); err != nil {
		return nil, err
	}
	return &Service{ID: resp.ServiceCreate.ID, Name: resp.ServiceCreate.Name}, nil
}

// DeleteService 删除服务
func (c *Client) DeleteService(ctx context.Context, serviceID string) error {
	var resp igql.ServiceDeleteResponse
	if err := c.gqlClient.Mutate(ctx, igql.ServiceDeleteMutation, map[string]any{"id": serviceID}, &resp); err != nil {
		return err
	}
	if !resp.ServiceDelete {
		return fmt.Errorf("service delete failed")
	}
	return nil
}

// ListServices 列出项目服务，并标识在指定环境下是否存在实例
// environmentRef 可传 ID 或 Name；若为空则不计算 HasInstance（均为 false）
func (c *Client) ListServices(ctx context.Context, projectID string, environmentRef string) ([]ServiceInEnvironment, error) {
	var proj igql.ProjectResponse
	if err := c.gqlClient.Query(ctx, igql.ProjectQuery, map[string]any{"id": projectID}, &proj); err != nil {
		return nil, err
	}
	// 解析环境ID
	envID := ""
	if strings.TrimSpace(environmentRef) != "" {
		for _, e := range proj.Project.Environments.Edges {
			if environmentRef == e.Node.ID || environmentRef == e.Node.Name {
				envID = e.Node.ID
				break
			}
		}
	}
	out := make([]ServiceInEnvironment, 0, len(proj.Project.Services.Edges))
	for _, s := range proj.Project.Services.Edges {
		has := false
		if envID != "" {
			for _, inst := range s.Node.ServiceInstances.Edges {
				if inst.Node.EnvironmentID == envID {
					has = true
					break
				}
			}
		}
		out = append(out, ServiceInEnvironment{Service: Service{ID: s.Node.ID, Name: s.Node.Name}, HasInstance: has})
	}
	return out, nil
}
