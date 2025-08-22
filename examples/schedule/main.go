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
	cli, err := railway.New(
		railway.WithAPIToken(apiToken),
		railway.WithEnvironment("production"),
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	project, err := cli.GetProject(ctx, "15fe48c7-8d02-4a18-aa4e-e7718e24291e")
	check(err)

	serviceD := "e81eb2f2-35f1-4c84-b89a-6e8cb9effa03"

	// 获取环境配置
	config, err := cli.GetEnvironmentConfig(ctx, project.Environments[0].ID, true, true)
	check(err)

	fmt.Println(config)

	instances := config.Environment.VolumeInstances
	fmt.Println(instances)

	for _, instance := range instances {
		if instance.ServiceID == serviceD {
			schedules, err := cli.GetVolumeBackupSchedules(ctx, instance.ID)
			check(err)
			fmt.Println(schedules)
			backupSchedules, err := cli.UpdateVolumeBackupSchedules(ctx, instance.ID, []string{railway.VolumeBackupScheduleMonthly, railway.VolumeBackupScheduleWeekly})
			check(err)
			fmt.Println(backupSchedules)
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
