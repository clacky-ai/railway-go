# 卷备份调度常量实现

## 概述

在 `@/railway` 目录下的 `volume.go` 文件中成功添加了卷备份调度类型的常量定义，提供了类型安全的调度类型使用方式。

## 实现内容

### 常量定义

在 `pkg/railway/volume.go` 中添加了：

```go
// 卷备份调度类型常量
const (
	VolumeBackupScheduleDaily   = "DAILY"
	VolumeBackupScheduleWeekly  = "WEEKLY"
	VolumeBackupScheduleMonthly = "MONTHLY"
)
```

## 常量说明

### 支持的调度类型

- `VolumeBackupScheduleDaily`: 每日备份调度
- `VolumeBackupScheduleWeekly`: 每周备份调度
- `VolumeBackupScheduleMonthly`: 每月备份调度

### 使用方式

#### 1. 基本使用

```go
// 使用常量而不是硬编码字符串
kinds := []string{
    railway.VolumeBackupScheduleDaily,
    railway.VolumeBackupScheduleWeekly,
    railway.VolumeBackupScheduleMonthly,
}

success, err := client.UpdateVolumeBackupSchedules(ctx, volumeInstanceID, kinds)
```

#### 2. 单个调度类型

```go
// 只启用每日备份
kinds := []string{railway.VolumeBackupScheduleDaily}

// 只启用每周备份
kinds := []string{railway.VolumeBackupScheduleWeekly}

// 只启用每月备份
kinds := []string{railway.VolumeBackupScheduleMonthly}
```

#### 3. 组合使用

```go
// 启用每日和每周备份
kinds := []string{
    railway.VolumeBackupScheduleDaily,
    railway.VolumeBackupScheduleWeekly,
}

// 启用每周和每月备份
kinds := []string{
    railway.VolumeBackupScheduleWeekly,
    railway.VolumeBackupScheduleMonthly,
}
```

## 优势

### 1. 类型安全

使用常量可以避免拼写错误：

```go
// ❌ 容易出错
kinds := []string{"DAILY", "WEEKLY", "MONTHLY"}

// ✅ 类型安全
kinds := []string{
    railway.VolumeBackupScheduleDaily,
    railway.VolumeBackupScheduleWeekly,
    railway.VolumeBackupScheduleMonthly,
}
```

### 2. 代码可维护性

当调度类型需要修改时，只需要在一个地方更新：

```go
// 如果需要修改调度类型名称，只需要修改常量定义
const (
    VolumeBackupScheduleDaily   = "DAILY_BACKUP"  // 修改这里
    VolumeBackupScheduleWeekly  = "WEEKLY_BACKUP" // 修改这里
    VolumeBackupScheduleMonthly = "MONTHLY_BACKUP" // 修改这里
)
```

### 3. IDE 支持

使用常量可以获得更好的 IDE 支持：
- 自动补全
- 重构支持
- 错误检查

### 4. 文档化

常量本身提供了清晰的文档说明：

```go
// 卷备份调度类型常量
const (
	VolumeBackupScheduleDaily   = "DAILY"   // 每日备份
	VolumeBackupScheduleWeekly  = "WEEKLY"  // 每周备份
	VolumeBackupScheduleMonthly = "MONTHLY" // 每月备份
)
```

## 更新内容

### 1. 示例文件更新

更新了以下示例文件以使用新的常量：

- `examples/volume_backup_schedules/main.go`
- `examples/volume_backup_schedule_update/main.go`

### 2. 文档更新

更新了以下文档以反映常量的使用：

- `examples/volume_backup_schedules/README.md`
- `docs/volume_backup_schedule_update.md`

### 3. 测试验证

创建了测试文件 `pkg/railway/volume_test.go` 来验证常量的正确性。

## 使用示例

### 完整示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/railwayapp/cli/pkg/railway"
)

func main() {
    // 创建客户端
    client, err := railway.New(railway.WithAPIToken(apiToken))
    if err != nil {
        log.Fatalf("创建客户端失败: %v", err)
    }

    ctx := context.Background()

    // 使用常量设置调度类型
    kinds := []string{
        railway.VolumeBackupScheduleDaily,
        railway.VolumeBackupScheduleWeekly,
        railway.VolumeBackupScheduleMonthly,
    }

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
}
```

## 验证

- ✅ 代码编译通过
- ✅ 测试通过
- ✅ 示例代码编译通过
- ✅ 文档更新完成

## 总结

成功添加了卷备份调度类型常量，包括：

1. **常量定义** - 在 `volume.go` 中定义了三个调度类型常量
2. **示例更新** - 更新了所有示例文件以使用新常量
3. **文档更新** - 更新了相关文档以反映常量的使用
4. **测试验证** - 创建了测试来验证常量的正确性

这些常量的添加提高了代码的类型安全性、可维护性和可读性，同时保持了与现有 API 的完全兼容性。 