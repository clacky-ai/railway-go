# 卷备份调度列表示例

这个示例演示了如何使用 Railway Go SDK 获取和更新卷备份调度列表。

## 功能说明

### GetVolumeBackupSchedules 函数
用于获取指定卷实例的所有备份调度配置，包括：
- 调度 ID
- 调度名称
- Cron 表达式
- 调度类型（DAILY、WEEKLY、MONTHLY 等）
- 保留时间（秒）
- 创建时间

### UpdateVolumeBackupSchedules 函数
用于更新指定卷实例的备份调度配置，可以启用或禁用特定类型的调度。

## 使用方法

### 1. 设置环境变量

```bash
export RAILWAY_API_TOKEN="your_railway_api_token"
export VOLUME_INSTANCE_ID="d9e8972e-5757-447a-a095-dcf40865d227"
```

### 2. 运行示例

```bash
go run main.go
```

### 3. 输出示例

```
=== 获取卷备份调度列表 ===
卷实例 d9e8972e-5757-447a-a095-dcf40865d227 的备份调度列表:
找到 3 个调度:

调度 1:
  ID: 01e93b25-40eb-46a7-a9dc-73ec7042fb30
  名称: Daily
  Cron 表达式: 8 20 * * *
  类型: DAILY
  保留时间: 518400 秒 (6 天)
  创建时间: 2025-08-22T01:40:02.966Z

调度 2:
  ID: 02b7512a-179d-4a17-a897-4cabe6b9400b
  名称: Weekly
  Cron 表达式: 3 14 * * 6
  类型: WEEKLY
  保留时间: 2332800 秒 (27 天)
  创建时间: 2025-08-22T01:34:20.275Z

调度 3:
  ID: 89c2040a-6600-4335-9523-6ab3314bf3c6
  名称: Monthly
  Cron 表达式: 15 0 1 * * 
  类型: MONTHLY
  保留时间: 7689600 秒 (89 天)
  创建时间: 2025-08-22T01:40:02.500Z

=== 更新卷备份调度 ===
设置调度类型: [WEEKLY MONTHLY DAILY]
✅ 卷备份调度更新成功!

=== 验证更新后的调度列表 ===
更新后找到 3 个调度:
...
```

## API 说明

### GetVolumeBackupSchedules 函数

```go
func (c *Client) GetVolumeBackupSchedules(ctx context.Context, volumeInstanceID string) ([]BackupSchedule, error)
```

**参数：**
- `ctx`: 上下文
- `volumeInstanceID`: 卷实例 ID

**返回值：**
- `[]BackupSchedule`: 备份调度列表
- `error`: 错误信息

### UpdateVolumeBackupSchedules 函数

```go
func (c *Client) UpdateVolumeBackupSchedules(ctx context.Context, volumeInstanceID string, kinds []string) (bool, error)
```

**参数：**
- `ctx`: 上下文
- `volumeInstanceID`: 卷实例 ID
- `kinds`: 要启用的调度类型数组（如 [railway.VolumeBackupScheduleDaily, railway.VolumeBackupScheduleWeekly, railway.VolumeBackupScheduleMonthly]）

**返回值：**
- `bool`: 更新是否成功
- `error`: 错误信息

### BackupSchedule 结构体

```go
type BackupSchedule struct {
    ID               string
    Name             string
    Cron             string
    Kind             string
    RetentionSeconds int64
    CreatedAt        string
}
```

**字段说明：**
- `ID`: 调度唯一标识符
- `Name`: 调度名称
- `Cron`: Cron 表达式，定义执行时间
- `Kind`: 调度类型（DAILY、WEEKLY、MONTHLY 等）
- `RetentionSeconds`: 备份保留时间（秒）
- `CreatedAt`: 创建时间

## 使用示例

### 获取调度列表

```go
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

### 更新调度配置

```go
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

## 支持的调度类型

- `railway.VolumeBackupScheduleDaily`: 每日备份
- `railway.VolumeBackupScheduleWeekly`: 每周备份
- `railway.VolumeBackupScheduleMonthly`: 每月备份

## 注意事项

1. 确保已正确设置 `RAILWAY_API_TOKEN` 环境变量
2. 确保提供的 `VOLUME_INSTANCE_ID` 是有效的卷实例 ID
3. 该功能需要相应的 API 权限
4. 更新调度配置会覆盖现有的调度设置
5. 调度类型数组中的值必须是有效的枚举值 