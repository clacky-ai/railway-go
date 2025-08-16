package railway

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	igql "github.com/railwayapp/cli/internal/gql"
)

// TemplateVariable 模板变量
type TemplateVariable struct {
	DefaultValue *string `json:"defaultValue,omitempty"`
	Value        *string `json:"value,omitempty"`
	Description  *string `json:"description,omitempty"`
	IsOptional   *bool   `json:"isOptional,omitempty"`
}

// TemplateService 模板服务
type TemplateService struct {
	Name      string                       `json:"name"`
	Variables map[string]*TemplateVariable `json:"variables,omitempty"`
}

// TemplateConfig 模板配置
type TemplateConfig struct {
	Services map[string]*TemplateService `json:"services,omitempty"`
}

// TemplateDeployOptions 模板部署选项
type TemplateDeployOptions struct {
	ProjectID     string
	EnvironmentID string
	TemplateCode  string
	Variables     map[string]string // 用户提供的变量，支持 "Service.Key" 和 "Key" 格式
}

// TemplateDetail 模板详情
type TemplateDetail struct {
	ID               string
	Name             string
	SerializedConfig json.RawMessage
}

// DeployTemplateWithConfig 部署模板（高级API，处理变量解析和用户交互）
func (c *Client) DeployTemplateWithConfig(ctx context.Context, opts TemplateDeployOptions) (*TemplateDeployResult, error) {
	// 1. 获取模板详情
	templateDetail, err := c.GetTemplateDetail(ctx, opts.TemplateCode)
	if err != nil {
		return nil, fmt.Errorf("获取模板详情失败: %w", err)
	}

	// 2. 解析模板配置
	var templateConfig TemplateConfig
	if len(templateDetail.SerializedConfig) > 0 {
		if err := json.Unmarshal(templateDetail.SerializedConfig, &templateConfig); err != nil {
			return nil, fmt.Errorf("解析模板配置失败: %w", err)
		}
	}

	// 3. 处理服务变量
	if err := c.processTemplateVariables(&templateConfig, opts.Variables); err != nil {
		return nil, err
	}

	// 4. 转换为 SerializedTemplateConfig
	serializedConfig, err := c.convertToSerializedConfig(templateConfig)
	if err != nil {
		return nil, fmt.Errorf("转换模板配置失败: %w", err)
	}

	// 5. 部署模板
	return c.DeployTemplate(ctx, opts.ProjectID, opts.EnvironmentID, templateDetail.ID, serializedConfig)
}

// GetTemplateDetail 获取模板详情
func (c *Client) GetTemplateDetail(ctx context.Context, templateCode string) (*TemplateDetail, error) {
	var resp igql.TemplateDetailResponse
	if err := c.gqlClient.Query(ctx, igql.TemplateDetailQuery, map[string]interface{}{
		"code": templateCode,
	}, &resp); err != nil {
		return nil, err
	}

	return &TemplateDetail{
		ID:               resp.Template.ID,
		Name:             resp.Template.Name,
		SerializedConfig: resp.Template.SerializedConfig,
	}, nil
}

// processTemplateVariables 处理模板变量
func (c *Client) processTemplateVariables(templateConfig *TemplateConfig, userVars map[string]string) error {
	if templateConfig.Services == nil {
		return nil
	}

	for _, service := range templateConfig.Services {
		if service.Variables == nil {
			continue
		}

		for key, variable := range service.Variables {
			var value string
			var found bool

			// 优先级1: 服务特定变量 (Service.Key)
			if val, exists := userVars[service.Name+"."+key]; exists {
				value = strings.TrimSpace(val)
				found = true
			} else if val, exists := userVars[key]; exists {
				// 优先级2: 全局变量 (Key)
				value = strings.TrimSpace(val)
				found = true
			} else if variable.DefaultValue != nil && strings.TrimSpace(*variable.DefaultValue) != "" {
				// 优先级3: 默认值
				value = strings.TrimSpace(*variable.DefaultValue)
				found = true
			} else if variable.IsOptional != nil && *variable.IsOptional {
				// 优先级4: 可选变量，跳过
				continue
			} else {
				// 优先级5: 必需变量，返回错误（在CLI层处理用户输入）
				description := ""
				if variable.Description != nil {
					description = *variable.Description
				}
				return fmt.Errorf("环境变量 %s (服务: %s) 是必需的，请提供值。描述: %s", key, service.Name, description)
			}

			if found && value != "" {
				variable.Value = &value
			}
		}
	}

	return nil
}

// convertToSerializedConfig 转换为序列化配置
func (c *Client) convertToSerializedConfig(templateConfig TemplateConfig) (igql.SerializedTemplateConfig, error) {
	serializedConfig := make(igql.SerializedTemplateConfig)
	configBytes, err := json.Marshal(templateConfig)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(configBytes, &serializedConfig); err != nil {
		return nil, err
	}
	return serializedConfig, nil
}
