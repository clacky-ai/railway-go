# 卷备份调度列表功能实现

## 概述

在 `@/railway` 目录下成功实现了获取和更新卷备份调度列表的功能，该功能允许用户查询和配置指定卷实例的备份调度。

## 实现内容

### 1. GraphQL 查询定义

在 `internal/gql/queries.go` 中添加了：

```go
// VolumeInstanceBackupScheduleList GraphQL查询
const VolumeInstanceBackupScheduleListQuery = `
query volumeInstanceBackupScheduleList($volumeInstanceId: String!) {
  volumeInstanceBackupScheduleList(volumeInstanceId: $volumeInstanceId) {
    id
    name
    cron
    kind
    retentionSeconds
    createdAt
  }
}
`

// VolumeInstanceBackupScheduleListResponse 卷实例备份调度列表响应
type VolumeInstanceBackupScheduleListResponse struct {
	VolumeInstanceBackupScheduleList []struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		Cron             string `json:"cron"`
		Kind             string `json:"kind"`
		RetentionSeconds int64  `json:"retentionSeconds"`
		CreatedAt        string `json:"createdAt"`
	} `json:"volumeInstanceBackupScheduleList"`
}
```

### 2. GraphQL 变更定义

在 `internal/gql/mutations.go` 中添加了：

```go
// VolumeInstanceBackupScheduleUpdate GraphQL变更
const VolumeInstanceBackupScheduleUpdateMutation = `
mutation volumeInstanceBackupScheduleUpdate($volumeInstanceId: String!, $kinds: [VolumeInstanceBackupScheduleKind!]!) {
  volumeInstanceBackupScheduleUpdate(
    volumeInstanceId: $volumeInstanceId
    kinds: $kinds
  )
}
`

// VolumeInstanceBackupScheduleUpdateResponse 卷备份调度更新响应
type VolumeInstanceBackupScheduleUpdateResponse struct {
	VolumeInstanceBackupScheduleUpdate bool `json:"volumeInstanceBackupScheduleUpdate"`
}
```

### 3. 数据结构定义

在 `pkg/railway/volume.go` 中添加了：

```go
// BackupSchedule 备份调度
type BackupSchedule struct {
	ID               string
	Name             string
	Cron             string
	Kind             string
	RetentionSeconds int64
	CreatedAt        string
}
```

### 4. API 函数实现

在 `pkg/railway/volume.go` 中添加了：

```go
// GetVolumeBackupSchedules 获取卷备份调度列表
func (c *Client) GetVolumeBackupSchedules(ctx context.Context, volumeInstanceID string) ([]BackupSchedule, error) {
	var resp igql.VolumeInstanceBackupScheduleListResponse
	if err := c.gqlClient.Query(ctx, igql.VolumeInstanceBackupScheduleListQuery, map[string]any{"volumeInstanceId": volumeInstanceID}, &resp); err != nil {
		return nil, err
	}

	schedules := make([]BackupSchedule, len(resp.VolumeInstanceBackupScheduleList))
	for i, schedule := range resp.VolumeInstanceBackupScheduleList {
		schedules[i] = BackupSchedule{
			ID:               schedule.ID,
			Name:             schedule.Name,
			Cron:             schedule.Cron,
			Kind:             schedule.Kind,
			RetentionSeconds: schedule.RetentionSeconds,
			CreatedAt:        schedule.CreatedAt,
		}
	}
	return schedules, nil
}

// UpdateVolumeBackupSchedules 更新卷备份调度
func (c *Client) UpdateVolumeBackupSchedules(ctx context.Context, volumeInstanceID string, kinds []string) (bool, error) {
	var resp igql.VolumeInstanceBackupScheduleUpdateResponse
	if err := c.gqlClient.Mutate(ctx, igql.VolumeInstanceBackupScheduleUpdateMutation, map[string]any{
		"volumeInstanceId": volumeInstanceID,
		"kinds":           kinds,
	}, &resp); err != nil {
		return false, err
	}
	return resp.VolumeInstanceBackupScheduleUpdate, nil
}
```

## 功能特性

### 支持的字段

- **ID**: 调度唯一标识符
- **Name**: 调度名称（如 "Daily", "Weekly", "Monthly"）
- **Cron**: Cron 表达式，定义执行时间（如 "8 20 * * *"）
- **Kind**: 调度类型（DAILY、WEEKLY、MONTHLY 等）
- **RetentionSeconds**: 备份保留时间（秒）
- **CreatedAt**: 创建时间

### 支持的调度类型

- `DAILY`: 每日备份
- `WEEKLY`: 每周备份
- `MONTHLY`: 每月备份

### 使用示例

#### 获取调度列表

```go
// 创建客户端
client, err := railway.New(
    railway.WithAPIToken(apiToken),
)
if err != nil {
    log.Fatalf("创建客户端失败: %v", err)
}

// 获取卷备份调度列表
schedules, err := client.GetVolumeBackupSchedules(ctx, volumeInstanceID)
if err != nil {
    log.Fatalf("获取卷备份调度列表失败: %v", err)
}

// 处理结果
for _, schedule := range schedules {
    fmt.Printf("调度: %s, Cron: %s, 类型: %s\n", 
        schedule.Name, schedule.Cron, schedule.Kind)
}
```

#### 更新调度配置

```go
// 设置要启用的调度类型
kinds := []string{"WEEKLY", "MONTHLY", "DAILY"}

// 更新卷备份调度
success, err := client.UpdateVolumeBackupSchedules(ctx, volumeInstanceID, kinds)
if err != nil {
    log.Fatalf("更新卷备份调度失败: %v", err)
}

if success {
    fmt.Println("✅ 卷备份调度更新成功!")
} else {
    fmt.Println("❌ 卷备份调度更新失败")
}
```

## 示例和文档

### 1. 完整示例

更新了 `examples/volume_backup_schedules/main.go` 示例文件，演示了完整的使用流程，包括：
- 获取调度列表
- 更新调度配置
- 验证更新结果

### 2. 文档

更新了 `examples/volume_backup_schedules/README.md` 文档，包含：
- 功能说明
- 使用方法
- API 说明
- 注意事项

### 3. 测试

更新了 `pkg/railway/volume_test.go` 测试文件，验证了：
- 数据结构的正确性
- 参数处理的正确性

## 返回数据格式

### 获取调度列表

该功能返回的数据格式与提供的 GraphQL 查询完全匹配：

```json
{
    "data": {
        "volumeInstanceBackupScheduleList": [
            {
                "id": "01e93b25-40eb-46a7-a9dc-73ec7042fb30",
                "name": "Daily",
                "cron": "8 20 * * *",
                "kind": "DAILY",
                "retentionSeconds": 518400,
                "createdAt": "2025-08-22T01:40:02.966Z"
            }
        ]
    }
}
```

### 更新调度配置

更新操作的返回格式：

```json
{
    "data": {
        "volumeInstanceBackupScheduleUpdate": true
    }
}
```

## 验证

- ✅ 代码编译通过
- ✅ 测试通过
- ✅ 示例代码编译通过
- ✅ 文档完整

## 总结

成功实现了卷备份调度的完整功能，包括：

1. **GraphQL 查询定义** - 完整的查询和响应结构
2. **GraphQL 变更定义** - 完整的变更和响应结构
3. **数据结构** - BackupSchedule 结构体
4. **API 函数** - GetVolumeBackupSchedules 和 UpdateVolumeBackupSchedules 方法
5. **示例代码** - 完整的使用示例
6. **文档** - 详细的使用说明
7. **测试** - 基本的结构和参数验证

该实现完全符合提供的 GraphQL 查询和变更格式，并遵循了现有代码库的设计模式和编码规范。 