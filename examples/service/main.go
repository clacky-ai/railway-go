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

	project, err := cli.GetProject(ctx, "15fe48c7-8d02-4a18-aa4e-e7718e24291e")
	check(err)

	srv := "797664e8-fcf8-4875-9e0f-165e2fabec4a"
	////err = cli.DeleteService(ctx, "76e9ed07-3968-4198-b66d-4788c11ee03d")
	////check(err)
	////available, message, err := cli.CheckCustomDomainAvailable(ctx, "fscc.clackyai.appss")
	////check(err)
	////fmt.Println(available, message)
	//domain, err := cli.CreateCustomDomain(ctx, project.ID, project.Environments[0].ID, "009d618d-d7e5-4f18-a488-bcf4d152e198", "mynginx.clackyai.app", nil)
	//check(err)
	//fmt.Println(domain)
	err = cli.Down(ctx, project.ID, project.Environments[0].ID, srv)

	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
