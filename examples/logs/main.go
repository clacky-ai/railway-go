package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	project, err := cli.GetProject(ctx, "95dac56d-97bc-4437-b764-2fe3800ad0c3")
	check(err)

	srv := "ec170d40-9da6-489f-8b82-2e90f64137e9"
	//deploy := "f62b534a-6970-4ee3-9db1-e43349dd96a0"
	deployments, err := cli.ListDeployments(ctx, project.ID, project.Environments[0].ID, &srv)
	check(err)
	fmt.Println(deployments)

	//err = cli.SubscribeDeploymentLogs(ctx, deployments[0].ID, "", 1000, func(timestamp, message string, attributes map[string]string) {
	//	log.Printf("[%s] %s", timestamp, message)
	//})

	cli.SubscribeBuildLogs(ctx, deployments[0].ID, "", 1000, func(timestamp, message string, attr map[string]string) {
		parsedTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("[%v] %s %v", parsedTime, message, attr)
	})
	//err = cli.SubscribeEnvironmentLogs(ctx, project.Environments[0].ID, "(  ) @snapshot:fe3c55a7-0ec9-4a27-bc4a-7d76eccd702a OR @replica:fe3c55a7-0ec9-4a27-bc4a-7d76eccd702a", 1000, "2025-08-27T02:55:21.874Z", "", "", nil, func(timestamp, message, severity string, tags map[string]*string, attr map[string]string) {
	//	fmt.Println(timestamp, message, severity, tags, attr)
	//})

	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
