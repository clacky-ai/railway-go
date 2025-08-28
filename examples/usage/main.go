package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/railwayapp/cli/pkg/railway"
)

func main() {
	apiToken := os.Getenv("RAILWAY_API_TOKEN")
	if apiToken == "" {
		log.Fatal("请设置 RAILWAY_API_TOKEN 环境变量")
	}

	// 创建 Railway 客户端
	client, err := railway.New(
		railway.WithAPIToken(apiToken),
		railway.WithEnvironment("production"),
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 设置上下文
	ctx := context.Background()

	// 项目ID（需要替换为实际的项目ID）
	projectID := "b95b2ae7-0fb0-4aba-bd37-532f3663f491"

	// 定义要查询的指标类型
	usageMeasurements := []railway.MetricMeasurement{
		railway.MetricMeasurementMemoryUsageGB,
		railway.MetricMeasurementCPUUsage,
		railway.MetricMeasurementNetworkTXGB,
		railway.MetricMeasurementDiskUsageGB,
		railway.MetricMeasurementBackupUsageGB,
	}

	metricsMeasurements := []railway.MetricMeasurement{
		railway.MetricMeasurementMemoryUsageGB,
		railway.MetricMeasurementCPUUsage,
		railway.MetricMeasurementNetworkTXGB,
	}

	// 设置时间范围（最近30天）
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// 可选：设置采样率（秒）
	sampleRateSeconds := 10000

	// 调用 API
	result, err := client.GetSingleProjectMetricsAndUsage(
		ctx,
		projectID,
		usageMeasurements,
		metricsMeasurements,
		startDate,
		endDate,
		&sampleRateSeconds,
	)
	if err != nil {
		log.Fatalf("获取项目指标和使用量失败: %v", err)
	}

	// 打印结果
	fmt.Printf("项目信息:\n")
	fmt.Printf("  ID: %s\n", result.Project.ID)
	fmt.Printf("  名称: %s\n", result.Project.Name)
	fmt.Printf("  创建时间: %s\n", result.Project.CreatedAt)
	if result.Project.DeletedAt != nil {
		fmt.Printf("  删除时间: %s\n", *result.Project.DeletedAt)
	}

	fmt.Printf("\n服务列表:\n")
	for _, service := range result.Project.Services {
		fmt.Printf("  - %s (%s)\n", service.Name, service.ID)
		if service.DeletedAt != nil {
			fmt.Printf("    已删除: %s\n", *service.DeletedAt)
		}
	}

	fmt.Printf("\n插件列表:\n")
	for _, plugin := range result.Project.Plugins {
		fmt.Printf("  - %s (%s)\n", plugin.Name, plugin.ID)
		if plugin.DeletedAt != nil {
			fmt.Printf("    已删除: %s\n", *plugin.DeletedAt)
		}
	}

	fmt.Printf("\n预估使用量:\n")
	for _, usage := range result.EstimatedUsage {
		fmt.Printf("  %s: %.2f\n", usage.Measurement, usage.EstimatedValue)
	}

	fmt.Printf("\n聚合使用量:\n")
	for _, usage := range result.Usage {
		fmt.Printf("  %s: %.2f\n", usage.Measurement, usage.Value)
		if usage.Tags.ServiceID != nil {
			fmt.Printf("    服务ID: %s\n", *usage.Tags.ServiceID)
		}
		if usage.Tags.PluginID != nil {
			fmt.Printf("    插件ID: %s\n", *usage.Tags.PluginID)
		}
	}

	fmt.Printf("\n指标数据:\n")
	for _, metric := range result.Metrics {
		fmt.Printf("  %s (%d 个数据点):\n", metric.Measurement, len(metric.Values))
		// 只显示前5个数据点
		for i, value := range metric.Values {
			if i >= 5 {
				fmt.Printf("    ... 还有 %d 个数据点\n", len(metric.Values)-5)
				break
			}
			timestamp := time.Unix(value.Timestamp, 0)
			fmt.Printf("    %s: %.6f\n", timestamp.Format("2006-01-02 15:04:05"), value.Value)
		}
	}

	// 演示 allProjectUsage API
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("演示 allProjectUsage API\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	// 设置团队ID（可选）
	teamID := "b0bd84cd-a416-496a-a619-c9259892e888"
	includeDeleted := true

	// 调用 allProjectUsage API
	allResult, err := client.GetAllProjectUsage(
		ctx,
		usageMeasurements,
		startDate,
		endDate,
		&teamID,
		nil, // userID (可选)
		&includeDeleted,
	)
	if err != nil {
		log.Fatalf("获取所有项目使用量失败: %v", err)
	}

	// 打印所有项目使用量结果
	fmt.Printf("项目总数: %d\n", len(allResult.Projects))
	fmt.Printf("预估使用量记录数: %d\n", len(allResult.EstimatedUsage))
	fmt.Printf("聚合使用量记录数: %d\n", len(allResult.Usage))

	// 按项目分组显示预估使用量
	projectUsageMap := make(map[string][]railway.EstimatedUsage)
	for _, usage := range allResult.EstimatedUsage {
		projectUsageMap[usage.ProjectID] = append(projectUsageMap[usage.ProjectID], usage)
	}

	// 创建项目名称映射
	projectNameMap := make(map[string]string)
	for _, project := range allResult.Projects {
		projectNameMap[project.ID] = project.Name
	}

	fmt.Printf("\n各项目预估使用量:\n")
	for projectID, usages := range projectUsageMap {
		projectName := projectNameMap[projectID]
		if projectName == "" {
			projectName = "未知项目"
		}
		fmt.Printf("  %s (%s):\n", projectName, projectID)
		for _, usage := range usages {
			fmt.Printf("    %s: %.2f\n", usage.Measurement, usage.EstimatedValue)
		}
	}

	// 显示聚合使用量统计
	fmt.Printf("\n聚合使用量统计:\n")
	usageByMeasurement := make(map[string]float64)
	for _, usage := range allResult.Usage {
		usageByMeasurement[usage.Measurement] += usage.Value
	}
	for measurement, totalValue := range usageByMeasurement {
		fmt.Printf("  %s 总计: %.2f\n", measurement, totalValue)
	}

	// 显示项目状态统计
	fmt.Printf("\n项目状态统计:\n")
	activeCount := 0
	deletedCount := 0
	for _, project := range allResult.Projects {
		if project.DeletedAt != nil {
			deletedCount++
		} else {
			activeCount++
		}
	}
	fmt.Printf("  活跃项目: %d\n", activeCount)
	fmt.Printf("  已删除项目: %d\n", deletedCount)
}
