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

	fmt.Printf("正在测试 API Token...\n")
	fmt.Printf("Token 前缀: %s...\n", apiToken[:10])

	// 创建 Railway 客户端
	cli, err := railway.New(
		railway.WithAPIToken(apiToken),
		railway.WithEnvironment("production"),
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	// 尝试获取项目列表来验证 token
	fmt.Println("正在获取项目列表...")
	projects, err := cli.ListProjects(ctx)
	if err != nil {
		log.Fatalf("获取项目列表失败: %v", err)
	}

	fmt.Printf("✅ Token 有效！\n")
	fmt.Printf("找到 %d 个项目:\n", len(projects))

	for i, project := range projects {
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, project.Name, project.ID)
	}

	// 如果有项目，尝试获取第一个项目的环境
	if len(projects) > 0 {
		projectID := projects[0].ID
		fmt.Printf("\n正在获取项目 '%s' 的环境...\n", projects[0].Name)

		project, err := cli.GetProject(ctx, projectID)
		if err != nil {
			log.Printf("获取项目详情失败: %v", err)
		} else {
			fmt.Printf("项目环境:\n")
			for i, env := range project.Environments {
				fmt.Printf("  %d. %s (ID: %s)\n", i+1, env.Name, env.ID)
			}
		}
	}
}
