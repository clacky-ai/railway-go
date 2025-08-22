# 卷备份调度更新功能实现

## 概述

在 `@/railway` 目录下成功实现了更新卷备份调度配置的功能，该功能允许用户动态配置指定卷实例的备份调度类型。

## GraphQL 变更

### 变更定义

```graphql
mutation volumeInstanceBackupScheduleUpdate($volumeInstanceId: String!, $kinds: [VolumeInstanceBackupScheduleKind!]!) {
  volumeInstanceBackupScheduleUpdate(
    volumeInstanceId: $volumeInstanceId
    kinds: $kinds
  )
}
```

### 请求参数

- `volumeInstanceId`: 卷实例 ID（必需）
- `kinds`: 要启用的调度类型数组（必需）

### 响应格式

```json
{
  "data": {
    "volumeInstanceBackupScheduleUpdate": true
  }
}
```

## 实现内容

### 1. GraphQL 变更定义

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

### 2. API 函数实现

在 `pkg/railway/volume.go` 中添加了：

```go
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

### 3. 调度类型常量

在 `pkg/railway/volume.go` 中添加了：

```go
// 卷备份调度类型常量
const (
	VolumeBackupScheduleDaily   = "DAILY"
	VolumeBackupScheduleWeekly  = "WEEKLY"
	VolumeBackupScheduleMonthly = "MONTHLY"
)
```

## 功能特性

### 支持的调度类型

- `railway.VolumeBackupScheduleDaily`: 每日备份
- `railway.VolumeBackupScheduleWeekly`: 每周备份
- `railway.VolumeBackupScheduleMonthly`: 每月备份

### 使用场景

1. **启用所有调度类型**
   ```go
   kinds := []string{railway.VolumeBackupScheduleDaily, railway.VolumeBackupScheduleWeekly, railway.VolumeBackupScheduleMonthly}
   ```

2. **只启用特定调度类型**
   ```go
   kinds := []string{railway.VolumeBackupScheduleDaily, railway.VolumeBackupScheduleWeekly}
   ```

3. **只启用单个调度类型**
   ```go
   kinds := []string{railway.VolumeBackupScheduleDaily}
   ```

4. **禁用所有备份调度**
   ```go
   kinds := []string{}
   ```

## 使用示例

### 基本用法

```go
// 创建客户端
client, err := railway.New(
    railway.WithAPIToken(apiToken),
)
if err != nil {
    log.Fatalf("创建客户端失败: %v", err)
}

// 设置要启用的调度类型
kinds := []string{railway.VolumeBackupScheduleWeekly, railway.VolumeBackupScheduleMonthly, railway.VolumeBackupScheduleDaily}

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

### 完整工作流程

```go
// 1. 获取当前调度配置
schedules, err := client.GetVolumeBackupSchedules(ctx, volumeInstanceID)
if err != nil {
    log.Fatalf("获取调度列表失败: %v", err)
}

fmt.Printf("当前启用了 %d 个调度\n", len(schedules))

// 2. 更新调度配置
newKinds := []string{railway.VolumeBackupScheduleDaily, railway.VolumeBackupScheduleWeekly}
success, err := client.UpdateVolumeBackupSchedules(ctx, volumeInstanceID, newKinds)
if err != nil {
    log.Fatalf("更新调度失败: %v", err)
}

// 3. 验证更新结果
updatedSchedules, err := client.GetVolumeBackupSchedules(ctx, volumeInstanceID)
if err != nil {
    log.Fatalf("获取更新后的调度列表失败: %v", err)
}

fmt.Printf("更新后启用了 %d 个调度\n", len(updatedSchedules))
```

## 示例代码

### 1. 完整示例

创建了 `examples/volume_backup_schedules/main.go` 示例文件，演示了：
- 获取调度列表
- 更新调度配置
- 验证更新结果

### 2. 更新专用示例

创建了 `examples/volume_backup_schedule_update/main.go` 示例文件，演示了：
- 多种调度配置场景
- 错误处理
- 结果验证

## 注意事项

### 重要提醒

1. **覆盖性更新**: 更新调度配置会完全覆盖现有的调度设置
2. **空数组处理**: 传入空数组会禁用所有备份调度
3. **枚举值验证**: 调度类型必须是有效的枚举值
4. **权限要求**: 需要相应的 API 权限

### 最佳实践

1. **先查询后更新**: 建议先获取当前配置，再进行更新
2. **错误处理**: 始终检查返回的错误信息
3. **验证结果**: 更新后建议验证配置是否生效
4. **备份配置**: 重要配置更新前建议备份当前设置

### 错误处理

```go
success, err := client.UpdateVolumeBackupSchedules(ctx, volumeInstanceID, kinds)
if err != nil {
    // 处理不同类型的错误
    switch {
    case strings.Contains(err.Error(), "permission"):
        log.Fatal("权限不足，无法更新调度配置")
    case strings.Contains(err.Error(), "not found"):
        log.Fatal("卷实例不存在")
    default:
        log.Fatalf("更新失败: %v", err)
    }
}
```

## 验证

- ✅ 代码编译通过
- ✅ 测试通过
- ✅ 示例代码编译通过
- ✅ 文档完整

## 总结

成功实现了卷备份调度更新功能，包括：

1. **GraphQL 变更定义** - 完整的变更和响应结构
2. **API 函数** - UpdateVolumeBackupSchedules 方法
3. **调度类型常量** - 预定义的调度类型常量
4. **示例代码** - 多种使用场景的示例
5. **文档** - 详细的使用说明和最佳实践
6. **测试** - 参数验证测试

该实现完全符合提供的 GraphQL 变更格式，并遵循了现有代码库的设计模式和编码规范。 