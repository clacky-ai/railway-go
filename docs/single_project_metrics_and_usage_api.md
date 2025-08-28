# Railway 项目指标和使用量 API 实现总结

## 概述

已成功在 Railway Go SDK 中实现了 `singleProjectMetricsAndUsage` 和 `allProjectUsage` API，这些 API 允许获取项目的指标和使用量数据。

## 实现的文件

### 1. GraphQL 查询定义
**文件**: `internal/gql/queries.go`

- 添加了 `SingleProjectMetricsAndUsageQuery` 常量，包含完整的 GraphQL 查询
- 添加了 `SingleProjectMetricsAndUsageResponse` 结构体，用于解析 GraphQL 响应
- **添加了 `AllProjectUsageQuery` 常量，包含所有项目使用量的 GraphQL 查询**
- **添加了 `AllProjectUsageResponse` 结构体，用于解析所有项目使用量的 GraphQL 响应**

### 2. API 实现
**文件**: `pkg/railway/usage.go`

- 定义了 `MetricMeasurement` 类型和相关的常量
- 实现了 `GetSingleProjectMetricsAndUsage` 方法
- 定义了完整的数据结构用于返回结果
- **实现了 `GetAllProjectUsage` 方法**
- **定义了 `AllProjectUsageResult` 数据结构**

### 3. 示例代码
**文件**: `examples/usage/main.go`

- 提供了完整的使用示例
- 演示了如何调用 API 和处理返回的数据
- **添加了 `allProjectUsage` API 的使用示例**
- **演示了按项目分组和统计功能**

### 4. 文档
**文件**: `examples/usage/README.md`

- 详细的使用说明和 API 文档
- 包含参数说明、返回数据结构和使用示例
- **添加了 `allProjectUsage` API 的完整文档**

## API 功能

### 支持的指标类型

- `MEMORY_USAGE_GB` - 内存使用量（GB）
- `CPU_USAGE` - CPU 使用量
- `NETWORK_TX_GB` - 网络传输量（GB）
- `DISK_USAGE_GB` - 磁盘使用量（GB）
- `BACKUP_USAGE_GB` - 备份使用量（GB）

### 返回的数据类型

#### SingleProjectMetricsAndUsage
1. **指标数据** (`Metrics`) - 时间序列数据，包含时间戳和数值
2. **预估使用量** (`EstimatedUsage`) - 预估的使用量数据
3. **聚合使用量** (`Usage`) - 按项目、服务、插件分组的聚合数据
4. **项目信息** (`Project`) - 项目基本信息，包括服务和插件列表

#### AllProjectUsage
1. **预估使用量** (`EstimatedUsage`) - 所有项目的预估使用量数据
2. **聚合使用量** (`Usage`) - 按项目分组的聚合使用量数据
3. **项目列表** (`Projects`) - 所有项目的基本信息，包括删除状态

## 使用方法

### SingleProjectMetricsAndUsage
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

### AllProjectUsage
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

## 参数说明

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

## 编译和测试

所有代码都已通过编译测试：

```bash
# 编译核心包
go build ./pkg/railway

# 编译示例
go build ./examples/usage

# 清理依赖
go mod tidy
```

## 注意事项

1. **认证**: 需要有效的 Railway API Token
2. **权限**: 项目 ID 必须是有效的且当前用户有访问权限
3. **时间范围**: 建议不超过 90 天
4. **采样率**: 影响数据精度和响应时间
5. **数据可用性**: 某些指标可能在某些项目或时间段内不可用
6. **团队过滤**: `GetAllProjectUsage` 支持按团队或用户过滤，但至少需要提供其中一个参数
7. **删除项目**: 包含已删除项目可能会增加响应时间

## GraphQL 查询详情

### SingleProjectMetricsAndUsage
使用的 GraphQL 查询包含以下字段：

- `metrics`: 获取时间序列指标数据
- `estimatedUsage`: 获取预估使用量
- `usage`: 获取聚合使用量（按项目、服务、插件分组）
- `project`: 获取项目基本信息

### AllProjectUsage
使用的 GraphQL 查询包含以下字段：

- `estimatedUsage`: 获取所有项目的预估使用量
- `usage`: 获取所有项目的聚合使用量（按项目分组）
- `projects`: 获取所有项目的基本信息

查询支持片段（fragments）来复用字段定义，包括：
- `MetricsResultFields`
- `MetricFields`
- `EstimatedUsageFields`
- `AggregatedUsageFields`

## 扩展性

该实现具有良好的扩展性：

1. **新的指标类型**: 可以通过添加新的 `MetricMeasurement` 常量来支持
2. **新的数据字段**: 可以通过修改 GraphQL 查询和响应结构体来添加
3. **自定义处理**: 返回的数据结构允许用户进行自定义的数据处理和分析
4. **团队和用户过滤**: 支持灵活的过滤条件
5. **删除项目处理**: 支持包含或排除已删除项目的数据

## 总结

`singleProjectMetricsAndUsage` 和 `allProjectUsage` API 已成功实现并集成到 Railway Go SDK 中。这些 API 提供了完整的项目指标和使用量数据访问功能，包括：

- 单个项目的详细指标和使用量数据
- 所有项目的使用量统计和汇总
- 时间序列数据、预估使用量和聚合使用量
- 团队和用户级别的数据过滤
- 已删除项目的处理支持

实现遵循了项目的现有模式和最佳实践，提供了完整的文档和示例代码。
