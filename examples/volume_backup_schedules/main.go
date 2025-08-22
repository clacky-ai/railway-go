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

	// 创建 Railway 客户端
	client, err := railway.New(
		railway.WithAPIToken(apiToken),
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 从环境变量或命令行参数获取卷实例ID
	volumeInstanceID := os.Getenv("VOLUME_INSTANCE_ID")
	if volumeInstanceID == "" {
		log.Fatal("请设置环境变量 VOLUME_INSTANCE_ID")
	}

	// 创建上下文
	ctx := context.Background()

	// 1. 获取卷备份调度列表
	fmt.Println("=== 获取卷备份调度列表 ===")
	schedules, err := client.GetVolumeBackupSchedules(ctx, volumeInstanceID)
	if err != nil {
		log.Fatalf("获取卷备份调度列表失败: %v", err)
	}

	// 打印结果
	fmt.Printf("卷实例 %s 的备份调度列表:\n", volumeInstanceID)
	fmt.Printf("找到 %d 个调度:\n\n", len(schedules))

	for i, schedule := range schedules {
		fmt.Printf("调度 %d:\n", i+1)
		fmt.Printf("  ID: %s\n", schedule.ID)
		fmt.Printf("  名称: %s\n", schedule.Name)
		fmt.Printf("  Cron 表达式: %s\n", schedule.Cron)
		fmt.Printf("  类型: %s\n", schedule.Kind)
		fmt.Printf("  保留时间: %d 秒 (%d 天)\n", schedule.RetentionSeconds, schedule.RetentionSeconds/86400)
		fmt.Printf("  创建时间: %s\n", schedule.CreatedAt)
		fmt.Println()
	}

	// 2. 更新卷备份调度
	fmt.Println("=== 更新卷备份调度 ===")

	// 设置要启用的调度类型
	kinds := []string{railway.VolumeBackupScheduleWeekly, railway.VolumeBackupScheduleMonthly, railway.VolumeBackupScheduleDaily}
	fmt.Printf("设置调度类型: %v\n", kinds)

	success, err := client.UpdateVolumeBackupSchedules(ctx, volumeInstanceID, kinds)
	if err != nil {
		log.Fatalf("更新卷备份调度失败: %v", err)
	}

	if success {
		fmt.Println("✅ 卷备份调度更新成功!")
	} else {
		fmt.Println("❌ 卷备份调度更新失败")
	}

	// 3. 再次获取调度列表以验证更新
	fmt.Println("\n=== 验证更新后的调度列表 ===")
	updatedSchedules, err := client.GetVolumeBackupSchedules(ctx, volumeInstanceID)
	if err != nil {
		log.Fatalf("获取更新后的卷备份调度列表失败: %v", err)
	}

	fmt.Printf("更新后找到 %d 个调度:\n\n", len(updatedSchedules))
	for i, schedule := range updatedSchedules {
		fmt.Printf("调度 %d:\n", i+1)
		fmt.Printf("  ID: %s\n", schedule.ID)
		fmt.Printf("  名称: %s\n", schedule.Name)
		fmt.Printf("  Cron 表达式: %s\n", schedule.Cron)
		fmt.Printf("  类型: %s\n", schedule.Kind)
		fmt.Printf("  保留时间: %d 秒 (%d 天)\n", schedule.RetentionSeconds, schedule.RetentionSeconds/86400)
		fmt.Printf("  创建时间: %s\n", schedule.CreatedAt)
		fmt.Println()
	}
}
