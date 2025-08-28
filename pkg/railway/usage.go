package railway

import (
	"context"
	"time"

	igql "github.com/railwayapp/cli/internal/gql"
)

// MetricMeasurement 指标测量类型
type MetricMeasurement string

const (
	MetricMeasurementMemoryUsageGB MetricMeasurement = "MEMORY_USAGE_GB"
	MetricMeasurementCPUUsage      MetricMeasurement = "CPU_USAGE"
	MetricMeasurementNetworkTXGB   MetricMeasurement = "NETWORK_TX_GB"
	MetricMeasurementDiskUsageGB   MetricMeasurement = "DISK_USAGE_GB"
	MetricMeasurementBackupUsageGB MetricMeasurement = "BACKUP_USAGE_GB"
)

// MetricValue 指标值
type MetricValue struct {
	Timestamp int64   `json:"ts"`
	Value     float64 `json:"value"`
}

// MetricResult 指标结果
type MetricResult struct {
	Measurement string        `json:"measurement"`
	Values      []MetricValue `json:"values"`
	Tags        struct {
		ProjectID *string `json:"projectId"`
	} `json:"tags"`
}

// EstimatedUsage 预估使用量
type EstimatedUsage struct {
	Measurement    string  `json:"measurement"`
	EstimatedValue float64 `json:"estimatedValue"`
	ProjectID      string  `json:"projectId"`
}

// UsageTags 使用量标签
type UsageTags struct {
	ProjectID *string `json:"projectId"`
	ServiceID *string `json:"serviceId"`
	PluginID  *string `json:"pluginId"`
}

// AggregatedUsage 聚合使用量
type AggregatedUsage struct {
	Measurement string    `json:"measurement"`
	Value       float64   `json:"value"`
	Tags        UsageTags `json:"tags"`
}

// ProjectUsageInfo 项目使用量信息
type ProjectUsageInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	DeletedAt *string `json:"deletedAt"`
	CreatedAt string  `json:"createdAt"`
	Plugins   []struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		DeletedAt *string `json:"deletedAt"`
	} `json:"plugins"`
	Services []struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		DeletedAt *string `json:"deletedAt"`
	} `json:"services"`
}

// SingleProjectMetricsAndUsageResult 单项目指标和使用量结果
type SingleProjectMetricsAndUsageResult struct {
	Metrics        []MetricResult    `json:"metrics"`
	EstimatedUsage []EstimatedUsage  `json:"estimatedUsage"`
	Usage          []AggregatedUsage `json:"usage"`
	Project        ProjectUsageInfo  `json:"project"`
}

// GetSingleProjectMetricsAndUsage 获取单个项目的指标和使用量数据
func (c *Client) GetSingleProjectMetricsAndUsage(
	ctx context.Context,
	projectID string,
	usageMeasurements []MetricMeasurement,
	metricsMeasurements []MetricMeasurement,
	startDate time.Time,
	endDate time.Time,
	sampleRateSeconds *int,
) (*SingleProjectMetricsAndUsageResult, error) {
	// 转换测量类型为字符串数组
	usageMeasurementsStr := make([]string, len(usageMeasurements))
	for i, m := range usageMeasurements {
		usageMeasurementsStr[i] = string(m)
	}

	metricsMeasurementsStr := make([]string, len(metricsMeasurements))
	for i, m := range metricsMeasurements {
		metricsMeasurementsStr[i] = string(m)
	}

	// 构建查询变量
	vars := map[string]any{
		"projectId":           projectID,
		"usageMeasurements":   usageMeasurementsStr,
		"metricsMeasurements": metricsMeasurementsStr,
		"startDate":           startDate.Format(time.RFC3339),
		"endDate":             endDate.Format(time.RFC3339),
	}

	if sampleRateSeconds != nil {
		vars["sampleRateSeconds"] = *sampleRateSeconds
	}

	var resp igql.SingleProjectMetricsAndUsageResponse
	if err := c.gqlClient.Query(ctx, igql.SingleProjectMetricsAndUsageQuery, vars, &resp); err != nil {
		return nil, err
	}

	// 转换响应数据
	result := &SingleProjectMetricsAndUsageResult{
		Metrics:        make([]MetricResult, len(resp.Metrics)),
		EstimatedUsage: make([]EstimatedUsage, len(resp.EstimatedUsage)),
		Usage:          make([]AggregatedUsage, len(resp.Usage)),
	}

	// 转换指标数据
	for i, m := range resp.Metrics {
		result.Metrics[i] = MetricResult{
			Measurement: m.Measurement,
			Values:      make([]MetricValue, len(m.Values)),
			Tags: struct {
				ProjectID *string `json:"projectId"`
			}{
				ProjectID: m.Tags.ProjectID,
			},
		}
		for j, v := range m.Values {
			result.Metrics[i].Values[j] = MetricValue{
				Timestamp: v.TS,
				Value:     v.Value,
			}
		}
	}

	// 转换预估使用量数据
	for i, eu := range resp.EstimatedUsage {
		result.EstimatedUsage[i] = EstimatedUsage{
			Measurement:    eu.Measurement,
			EstimatedValue: eu.EstimatedValue,
			ProjectID:      eu.ProjectID,
		}
	}

	// 转换聚合使用量数据
	for i, u := range resp.Usage {
		result.Usage[i] = AggregatedUsage{
			Measurement: u.Measurement,
			Value:       u.Value,
			Tags: UsageTags{
				ProjectID: u.Tags.ProjectID,
				ServiceID: u.Tags.ServiceID,
				PluginID:  u.Tags.PluginID,
			},
		}
	}

	// 转换项目信息
	result.Project = ProjectUsageInfo{
		ID:        resp.Project.ID,
		Name:      resp.Project.Name,
		DeletedAt: resp.Project.DeletedAt,
		CreatedAt: resp.Project.CreatedAt,
		Plugins: make([]struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			DeletedAt *string `json:"deletedAt"`
		}, len(resp.Project.Plugins.Edges)),
		Services: make([]struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			DeletedAt *string `json:"deletedAt"`
		}, len(resp.Project.Services.Edges)),
	}

	for i, p := range resp.Project.Plugins.Edges {
		result.Project.Plugins[i] = struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			DeletedAt *string `json:"deletedAt"`
		}{
			ID:        p.Node.ID,
			Name:      p.Node.Name,
			DeletedAt: p.Node.DeletedAt,
		}
	}

	for i, s := range resp.Project.Services.Edges {
		result.Project.Services[i] = struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			DeletedAt *string `json:"deletedAt"`
		}{
			ID:        s.Node.ID,
			Name:      s.Node.Name,
			DeletedAt: s.Node.DeletedAt,
		}
	}

	return result, nil
}

// AllProjectUsageResult 所有项目使用量结果
type AllProjectUsageResult struct {
	EstimatedUsage []EstimatedUsage  `json:"estimatedUsage"`
	Usage          []AggregatedUsage `json:"usage"`
	Projects       []struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		DeletedAt *string `json:"deletedAt"`
		CreatedAt string  `json:"createdAt"`
	} `json:"projects"`
}

// GetAllProjectUsage 获取所有项目的使用量数据
func (c *Client) GetAllProjectUsage(
	ctx context.Context,
	usageMeasurements []MetricMeasurement,
	startDate time.Time,
	endDate time.Time,
	teamID *string,
	userID *string,
	includeDeleted *bool,
) (*AllProjectUsageResult, error) {
	// 转换测量类型为字符串数组
	usageMeasurementsStr := make([]string, len(usageMeasurements))
	for i, m := range usageMeasurements {
		usageMeasurementsStr[i] = string(m)
	}

	// 构建查询变量
	vars := map[string]any{
		"usageMeasurements": usageMeasurementsStr,
		"startDate":         startDate.Format(time.RFC3339),
		"endDate":           endDate.Format(time.RFC3339),
	}

	// 添加可选参数
	if teamID != nil {
		vars["teamId"] = *teamID
	}
	if userID != nil {
		vars["userId"] = *userID
	}
	if includeDeleted != nil {
		vars["includeDeleted"] = *includeDeleted
	}

	var resp igql.AllProjectUsageResponse
	if err := c.gqlClient.Query(ctx, igql.AllProjectUsageQuery, vars, &resp); err != nil {
		return nil, err
	}

	// 转换响应数据
	result := &AllProjectUsageResult{
		EstimatedUsage: make([]EstimatedUsage, len(resp.EstimatedUsage)),
		Usage:          make([]AggregatedUsage, len(resp.Usage)),
		Projects: make([]struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			DeletedAt *string `json:"deletedAt"`
			CreatedAt string  `json:"createdAt"`
		}, len(resp.Projects.Edges)),
	}

	// 转换预估使用量数据
	for i, eu := range resp.EstimatedUsage {
		result.EstimatedUsage[i] = EstimatedUsage{
			Measurement:    eu.Measurement,
			EstimatedValue: eu.EstimatedValue,
			ProjectID:      eu.ProjectID,
		}
	}

	// 转换聚合使用量数据
	for i, u := range resp.Usage {
		result.Usage[i] = AggregatedUsage{
			Measurement: u.Measurement,
			Value:       u.Value,
			Tags: UsageTags{
				ProjectID: u.Tags.ProjectID,
				ServiceID: u.Tags.ServiceID,
				PluginID:  u.Tags.PluginID,
			},
		}
	}

	// 转换项目信息
	for i, p := range resp.Projects.Edges {
		result.Projects[i] = struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			DeletedAt *string `json:"deletedAt"`
			CreatedAt string  `json:"createdAt"`
		}{
			ID:        p.Node.ID,
			Name:      p.Node.Name,
			DeletedAt: p.Node.DeletedAt,
			CreatedAt: p.Node.CreatedAt,
		}
	}

	return result, nil
}
