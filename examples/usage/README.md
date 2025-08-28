# Railway 项目指标和使用量 API 示例

这个示例演示了如何使用 Railway Go SDK 的 `GetSingleProjectMetricsAndUsage` 和 `GetAllProjectUsage` API 来获取项目的指标和使用量数据。

## 功能特性

- 获取单个项目的实时指标数据（内存、CPU、网络、磁盘、备份使用量）
- 获取单个项目的预估使用量数据
- 获取单个项目的聚合使用量数据（按项目、服务、插件分组）
- 获取单个项目的基本信息（包括服务和插件列表）
- **获取所有项目的使用量数据**
- **获取团队或用户的所有项目使用量统计**
- **支持包含已删除项目的使用量数据**

## 支持的指标类型

- `MEMORY_USAGE_GB` - 内存使用量（GB）
- `CPU_USAGE` - CPU 使用量
- `NETWORK_TX_GB` - 网络传输量（GB）
- `DISK_USAGE_GB` - 磁盘使用量（GB）
- `BACKUP_USAGE_GB` - 备份使用量（GB）

## 使用方法

### 1. 单个项目指标和使用量 (GetSingleProjectMetricsAndUsage)

```go
// 创建客户端
client, err := railway.New()
if err != nil {
    log.Fatal(err)
}

// 定义要查询的指标
measurements := []railway.MetricMeasurement{
    railway.MetricMeasurementMemoryUsageGB,
    railway.MetricMeasurementCPUUsage,
    railway.MetricMeasurementNetworkTXGB,
}

// 设置时间范围
startDate := time.Now().AddDate(0, 0, -30)
endDate := time.Now()

// 调用 API
result, err := client.GetSingleProjectMetricsAndUsage(
    context.Background(),
    projectID,
    measurements, // usageMeasurements
    measurements, // metricsMeasurements
    startDate,
    endDate,
    nil, // sampleRateSeconds (可选)
)
```

### 2. 所有项目使用量 (GetAllProjectUsage)

```go
// 设置可选参数
teamID := "your-team-id"
userID := "your-user-id" // 可选
includeDeleted := true

// 调用 API
result, err := client.GetAllProjectUsage(
    context.Background(),
    usageMeasurements,
    startDate,
    endDate,
    &teamID,        // 可选：团队ID
    &userID,        // 可选：用户ID
    &includeDeleted, // 可选：是否包含已删除项目
)
```

### 3. 处理不同类型的数据

```go
// 处理指标数据（时间序列数据）
for _, metric := range result.Metrics {
    fmt.Printf("指标: %s\n", metric.Measurement)
    for _, value := range metric.Values {
        timestamp := time.Unix(value.Timestamp, 0)
        fmt.Printf("  %s: %.6f\n", timestamp.Format("2006-01-02 15:04:05"), value.Value)
    }
}

// 处理预估使用量
for _, usage := range result.EstimatedUsage {
    fmt.Printf("预估 %s: %.2f\n", usage.Measurement, usage.EstimatedValue)
}

// 处理聚合使用量
for _, usage := range result.Usage {
    fmt.Printf("聚合 %s: %.2f\n", usage.Measurement, usage.Value)
    if usage.Tags.ServiceID != nil {
        fmt.Printf("  服务ID: %s\n", *usage.Tags.ServiceID)
    }
}
```

## API 参数说明

### GetSingleProjectMetricsAndUsage 参数

#### 必需参数
- `projectID` (string): 项目 ID
- `usageMeasurements` ([]MetricMeasurement): 使用量测量类型数组
- `metricsMeasurements` ([]MetricMeasurement): 指标测量类型数组
- `startDate` (time.Time): 开始时间
- `endDate` (time.Time): 结束时间

#### 可选参数
- `sampleRateSeconds` (*int): 采样率（秒），用于控制指标数据的时间间隔

### GetAllProjectUsage 参数

#### 必需参数
- `usageMeasurements` ([]MetricMeasurement): 使用量测量类型数组
- `startDate` (time.Time): 开始时间
- `endDate` (time.Time): 结束时间

#### 可选参数
- `teamID` (*string): 团队ID，用于过滤特定团队的项目
- `userID` (*string): 用户ID，用于过滤特定用户的项目
- `includeDeleted` (*bool): 是否包含已删除的项目

## 返回数据结构

### SingleProjectMetricsAndUsageResult
```go
type SingleProjectMetricsAndUsageResult struct {
    Metrics         []MetricResult     // 指标时间序列数据
    EstimatedUsage  []EstimatedUsage   // 预估使用量
    Usage           []AggregatedUsage  // 聚合使用量
    Project         ProjectUsageInfo   // 项目信息
}
```

### AllProjectUsageResult
```go
type AllProjectUsageResult struct {
    EstimatedUsage []EstimatedUsage  // 预估使用量
    Usage          []AggregatedUsage // 聚合使用量
    Projects       []struct {
        ID        string  `json:"id"`
        Name      string  `json:"name"`
        DeletedAt *string `json:"deletedAt"`
        CreatedAt string  `json:"createdAt"`
    } `json:"projects"`
}
```

## 运行示例

1. 确保已设置 Railway API Token 环境变量：
   ```bash
   export RAILWAY_API_TOKEN="your-api-token"
   ```

2. 修改 `main.go` 中的项目 ID 和团队 ID：
   ```go
   projectID := "your-actual-project-id"
   teamID := "your-actual-team-id"
   ```

3. 运行示例：
   ```bash
   go run main.go
   ```

## 注意事项

- 需要有效的 Railway API Token
- 项目 ID 必须是有效的且当前用户有访问权限
- 时间范围不应过大，建议不超过 90 天
- 采样率会影响数据精度和响应时间
- 某些指标可能在某些项目或时间段内不可用
- `GetAllProjectUsage` 支持按团队或用户过滤，但至少需要提供其中一个参数
- 包含已删除项目可能会增加响应时间
