package main

import (
	"context"
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

	//project, err := cli.GetProject(ctx, "c9796eb2-a1fe-42d7-bd9d-04e4a5a150d0")
	//check(err)

	err = cli.DeleteService(ctx, "76e9ed07-3968-4198-b66d-4788c11ee03d")
	check(err)

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
