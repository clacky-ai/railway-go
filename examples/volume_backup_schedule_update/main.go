package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/railwayapp/cli/pkg/railway"
)

func main() {
	// 检查环境变量
	apiToken := os.Getenv("RAILWAY_API_TOKEN")
	if apiToken == "" {
		log.Fatal("请设置 RAILWAY_API_TOKEN 环境变量")
	}

	volumeInstanceID := os.Getenv("VOLUME_INSTANCE_ID")
	if volumeInstanceID == "" {
		log.Fatal("请设置环境变量 VOLUME_INSTANCE_ID")
	}

	// 创建 Railway 客户端
	client, err := railway.New(
		railway.WithAPIToken(apiToken),
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	// 演示不同的调度配置场景
	scenarios := []struct {
		name  string
		kinds []string
	}{
		{
			name:  "启用所有调度类型",
			kinds: []string{railway.VolumeBackupScheduleDaily, railway.VolumeBackupScheduleWeekly, railway.VolumeBackupScheduleMonthly},
		},
		{
			name:  "只启用每日和每周备份",
			kinds: []string{railway.VolumeBackupScheduleDaily, railway.VolumeBackupScheduleWeekly},
		},
		{
			name:  "只启用每日备份",
			kinds: []string{railway.VolumeBackupScheduleDaily},
		},
		{
			name:  "禁用所有备份调度",
			kinds: []string{},
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n=== 场景 %d: %s ===\n", i+1, scenario.name)
		fmt.Printf("设置调度类型: %v\n", scenario.kinds)

		// 更新卷备份调度
		success, err := client.UpdateVolumeBackupSchedules(ctx, volumeInstanceID, scenario.kinds)
		if err != nil {
			log.Printf("更新卷备份调度失败: %v", err)
			continue
		}

		if success {
			fmt.Println("✅ 卷备份调度更新成功!")
		} else {
			fmt.Println("❌ 卷备份调度更新失败")
		}

		// 验证更新结果
		fmt.Println("验证更新结果:")
		schedules, err := client.GetVolumeBackupSchedules(ctx, volumeInstanceID)
		if err != nil {
			log.Printf("获取调度列表失败: %v", err)
			continue
		}

		if len(schedules) == 0 {
			fmt.Println("  当前没有启用的调度")
		} else {
			fmt.Printf("  当前启用了 %d 个调度:\n", len(schedules))
			for _, schedule := range schedules {
				fmt.Printf("    - %s (%s): %s\n", schedule.Name, schedule.Kind, schedule.Cron)
			}
		}
	}

	fmt.Println("\n=== 使用说明 ===")
	fmt.Println("支持的调度类型:")
	fmt.Printf("  - %s: 每日备份\n", railway.VolumeBackupScheduleDaily)
	fmt.Printf("  - %s: 每周备份\n", railway.VolumeBackupScheduleWeekly)
	fmt.Printf("  - %s: 每月备份\n", railway.VolumeBackupScheduleMonthly)
	fmt.Println("\n注意事项:")
	fmt.Println("  - 更新调度配置会覆盖现有的调度设置")
	fmt.Println("  - 传入空数组会禁用所有备份调度")
	fmt.Println("  - 调度类型必须是有效的枚举值")
}
