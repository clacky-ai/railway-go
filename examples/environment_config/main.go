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

	project, err := cli.GetProject(ctx, "c9796eb2-a1fe-42d7-bd9d-04e4a5a150d0")
	check(err)

	// 环境ID - 请替换为实际的环境ID
	environmentID := "fed9b227-7645-4091-8d6e-005077dc0e2c"
	environmentID = project.Environments[0].ID

	fmt.Printf("正在查询环境配置...\n")
	fmt.Printf("环境ID: %s\n", environmentID)
	fmt.Printf("API Token: %s...\n", apiToken[:10]) // 只显示前10个字符

	// 获取环境配置
	config, err := cli.GetEnvironmentConfig(ctx, environmentID, true, true)
	check(err)

	fmt.Println(config)

	instances := config.Environment.VolumeInstances
	fmt.Println(instances)

	for _, instance := range instances {
		if instance.ServiceID == "76e9ed07-3968-4198-b66d-4788c11ee03d" {
			fmt.Println(instance.ID)
			backup, err := cli.CreateVolumeBackup(ctx, instance.ID)
			check(err)
			fmt.Println(backup)
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
