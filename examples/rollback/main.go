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

	project, err := cli.GetProject(ctx, "3d2ba02b-5a57-4de9-9824-7152168050de")
	check(err)

	//serviceID := "572d034d-db00-42d3-86cc-28a0a8dedfc3"
	deployments, err := cli.ListDeployments(ctx, project.ID, project.Environments[0].ID, nil)
	check(err)
	fmt.Println(deployments)

	deploymentID := "d067de00-bb82-4296-b643-46743017ad91"

	success, err := cli.RollbackDeployment(ctx, deploymentID)
	check(err)
	fmt.Printf("部署回滚成功: %t\n", success)

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
