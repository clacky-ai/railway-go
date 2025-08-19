package railway

import (
	"context"
	"encoding/json"

	"github.com/railwayapp/cli/internal/gql"
)

// EnvironmentConfig 环境配置结构
type EnvironmentConfig struct {
	Environment              EnvironmentDetail        `json:"environment"`
	EnvironmentStagedChanges EnvironmentStagedChanges `json:"environmentStagedChanges"`
}

// EnvironmentDetail 环境详情
type EnvironmentDetail struct {
	ID               string                 `json:"id"`
	Config           map[string]interface{} `json:"config"`
	ServiceInstances []ServiceInstance      `json:"serviceInstances"`
	VolumeInstances  []VolumeInstance       `json:"volumeInstances"`
}

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID               string                 `json:"id"`
	IsUpdatable      bool                   `json:"isUpdatable"`
	ServiceID        string                 `json:"serviceId"`
	EnvironmentID    string                 `json:"environmentId"`
	RailpackInfo     *json.RawMessage       `json:"railpackInfo"`
	LatestDeployment *EnvironmentDeployment `json:"latestDeployment"`
}

// EnvironmentDeployment 环境配置中的部署信息
type EnvironmentDeployment struct {
	ID                      string                 `json:"id"`
	ServiceID               string                 `json:"serviceId"`
	EnvironmentID           string                 `json:"environmentId"`
	CreatedAt               string                 `json:"createdAt"`
	UpdatedAt               string                 `json:"updatedAt"`
	StatusUpdatedAt         string                 `json:"statusUpdatedAt"`
	Status                  string                 `json:"status"`
	StaticURL               *string                `json:"staticUrl"`
	SuggestAddServiceDomain bool                   `json:"suggestAddServiceDomain"`
	Meta                    map[string]interface{} `json:"meta"`
}

// VolumeInstance 卷实例
type VolumeInstance struct {
	ID                string `json:"id"`
	VolumeID          string `json:"volumeId"`
	EnvironmentID     string `json:"environmentId"`
	ServiceID         string `json:"serviceId"`
	ExternalID        string `json:"externalId"`
	IsPendingDeletion bool   `json:"isPendingDeletion"`
	State             string `json:"state"`
	Type              string `json:"type"`
}

// EnvironmentStagedChanges 环境阶段性变更
type EnvironmentStagedChanges struct {
	ID               string                 `json:"id"`
	CreatedAt        string                 `json:"createdAt"`
	UpdatedAt        string                 `json:"updatedAt"`
	Status           string                 `json:"status"`
	LastAppliedError *string                `json:"lastAppliedError"`
	Patch            map[string]interface{} `json:"patch"`
}

// GetEnvironmentConfig 获取环境配置
func (c *Client) GetEnvironmentConfig(ctx context.Context, environmentID string, decryptVariables bool, decryptPatchVariables bool) (*EnvironmentConfig, error) {
	// 构建查询变量
	variables := map[string]interface{}{
		"environmentId": environmentID,
	}

	// 添加可选参数
	variables["decryptVariables"] = decryptVariables
	variables["decryptPatchVariables"] = decryptPatchVariables

	// 执行查询
	var response gql.EnvironmentConfigResponse
	err := c.gqlClient.QueryInternal(ctx, gql.EnvironmentConfigQuery, variables, &response)
	if err != nil {
		return nil, err
	}

	// 转换响应数据
	result := &EnvironmentConfig{}

	// 解析环境详情
	result.Environment.ID = response.Environment.ID

	// 解析配置
	if err := json.Unmarshal(response.Environment.Config, &result.Environment.Config); err != nil {
		return nil, err
	}

	// 解析服务实例
	for _, edge := range response.Environment.ServiceInstances.Edges {
		node := edge.Node
		serviceInstance := ServiceInstance{
			ID:            node.ID,
			IsUpdatable:   node.IsUpdatable,
			ServiceID:     node.ServiceID,
			EnvironmentID: node.EnvironmentID,
			RailpackInfo:  &node.RailpackInfo,
		}

		// 解析最新部署信息
		if node.LatestDeployment != nil {
			deployment := &EnvironmentDeployment{
				ID:                      node.LatestDeployment.ID,
				ServiceID:               node.LatestDeployment.ServiceID,
				EnvironmentID:           node.LatestDeployment.EnvironmentID,
				CreatedAt:               node.LatestDeployment.CreatedAt,
				UpdatedAt:               node.LatestDeployment.UpdatedAt,
				StatusUpdatedAt:         node.LatestDeployment.StatusUpdatedAt,
				Status:                  node.LatestDeployment.Status,
				StaticURL:               node.LatestDeployment.StaticURL,
				SuggestAddServiceDomain: node.LatestDeployment.SuggestAddServiceDomain,
			}

			// 解析 Meta
			if err := json.Unmarshal(node.LatestDeployment.Meta, &deployment.Meta); err != nil {
				return nil, err
			}

			serviceInstance.LatestDeployment = deployment
		}

		result.Environment.ServiceInstances = append(result.Environment.ServiceInstances, serviceInstance)
	}

	// 解析卷实例
	for _, edge := range response.Environment.VolumeInstances.Edges {
		node := edge.Node
		volumeInstance := VolumeInstance{
			ID:                node.ID,
			VolumeID:          node.VolumeID,
			EnvironmentID:     node.EnvironmentID,
			ServiceID:         node.ServiceID,
			ExternalID:        node.ExternalID,
			IsPendingDeletion: node.IsPendingDeletion,
			State:             node.State,
			Type:              node.Type,
		}
		result.Environment.VolumeInstances = append(result.Environment.VolumeInstances, volumeInstance)
	}

	// 解析环境阶段性变更
	result.EnvironmentStagedChanges.ID = response.EnvironmentStagedChanges.ID
	result.EnvironmentStagedChanges.CreatedAt = response.EnvironmentStagedChanges.CreatedAt
	result.EnvironmentStagedChanges.UpdatedAt = response.EnvironmentStagedChanges.UpdatedAt
	result.EnvironmentStagedChanges.Status = response.EnvironmentStagedChanges.Status
	result.EnvironmentStagedChanges.LastAppliedError = response.EnvironmentStagedChanges.LastAppliedError

	// 解析 Patch
	if err := json.Unmarshal(response.EnvironmentStagedChanges.Patch, &result.EnvironmentStagedChanges.Patch); err != nil {
		return nil, err
	}

	return result, nil
}
