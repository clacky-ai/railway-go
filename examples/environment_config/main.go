package main

import (
	"context"
	"encoding/json"
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

	// 环境ID - 请替换为实际的环境ID
	environmentID := "fed9b227-7645-4091-8d6e-005077dc0e2c"

	// 检查环境ID是否为空
	if environmentID == "" {
		log.Fatal("请设置有效的环境ID")
	}

	fmt.Printf("正在查询环境配置...\n")
	fmt.Printf("环境ID: %s\n", environmentID)
	fmt.Printf("API Token: %s...\n", apiToken[:10]) // 只显示前10个字符

	// 获取环境配置
	config, err := cli.GetEnvironmentConfig(ctx, environmentID, true, true)
	if err != nil {
		log.Fatalf("获取环境配置失败: %v", err)
	}

	// 打印环境基本信息
	fmt.Printf("\n✅ 查询成功！\n")
	fmt.Printf("环境ID: %s\n", config.Environment.ID)
	fmt.Printf("服务实例数量: %d\n", len(config.Environment.ServiceInstances))
	fmt.Printf("卷实例数量: %d\n", len(config.Environment.VolumeInstances))
	fmt.Printf("暂存变更状态: %s\n", config.EnvironmentStagedChanges.Status)

	// 打印服务实例信息
	fmt.Println("\n=== 服务实例 ===")
	for i, service := range config.Environment.ServiceInstances {
		fmt.Printf("服务 %d:\n", i+1)
		fmt.Printf("  ID: %s\n", service.ID)
		fmt.Printf("  服务ID: %s\n", service.ServiceID)
		fmt.Printf("  环境ID: %s\n", service.EnvironmentID)
		fmt.Printf("  可更新: %t\n", service.IsUpdatable)

		if service.LatestDeployment != nil {
			fmt.Printf("  最新部署:\n")
			fmt.Printf("    部署ID: %s\n", service.LatestDeployment.ID)
			fmt.Printf("    状态: %s\n", service.LatestDeployment.Status)
			fmt.Printf("    创建时间: %s\n", service.LatestDeployment.CreatedAt)
			if service.LatestDeployment.StaticURL != nil {
				fmt.Printf("    静态URL: %s\n", *service.LatestDeployment.StaticURL)
			}
		}
		fmt.Println()
	}

	// 打印卷实例信息
	fmt.Println("=== 卷实例 ===")
	for i, volume := range config.Environment.VolumeInstances {
		fmt.Printf("卷 %d:\n", i+1)
		fmt.Printf("  ID: %s\n", volume.ID)
		fmt.Printf("  卷ID: %s\n", volume.VolumeID)
		fmt.Printf("  服务ID: %s\n", volume.ServiceID)
		fmt.Printf("  外部ID: %s\n", volume.ExternalID)
		fmt.Printf("  状态: %s\n", volume.State)
		fmt.Printf("  类型: %s\n", volume.Type)
		fmt.Printf("  待删除: %t\n", volume.IsPendingDeletion)
		fmt.Println()
	}

	// 打印暂存变更信息
	fmt.Println("=== 暂存变更 ===")
	fmt.Printf("ID: %s\n", config.EnvironmentStagedChanges.ID)
	fmt.Printf("状态: %s\n", config.EnvironmentStagedChanges.Status)
	fmt.Printf("创建时间: %s\n", config.EnvironmentStagedChanges.CreatedAt)
	fmt.Printf("更新时间: %s\n", config.EnvironmentStagedChanges.UpdatedAt)
	if config.EnvironmentStagedChanges.LastAppliedError != nil {
		fmt.Printf("最后应用错误: %s\n", *config.EnvironmentStagedChanges.LastAppliedError)
	}

	// 打印完整配置（JSON格式）
	fmt.Println("\n=== 完整配置 (JSON) ===")
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("序列化配置失败: %v", err)
	} else {
		fmt.Println(string(configJSON))
	}
}
