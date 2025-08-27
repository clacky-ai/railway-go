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

	serviceID := "e81eb2f2-35f1-4c84-b89a-6e8cb9effa03"
	environmentID := "241e0310-96cb-4d94-9a70-cb8420991c2a"

	s := "sssss"
	variables := map[string]*string{
		"ssss": &s,
		"fs":   nil,
	}

	ctx := context.Background()

	stageID, err := client.StageServiceVariables(ctx, environmentID, serviceID, variables)
	if err != nil {
		log.Fatalf("暂存服务变量失败: %v", err)
	}

	fmt.Printf("✅ 暂存成功! 暂存ID: %s\n", stageID)

	// 示例 2: 使用完整的配置结构暂存环境变更
	fmt.Println("\n=== 示例 2: 暂存完整环境配置 ===")

}
