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
	client, err := railway.New(
		railway.WithAPIToken(apiToken),
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 从环境变量或命令行参数获取环境ID
	environmentID := "321cbb6c-5b14-4438-8e06-05df873584a9"

	// 可选参数

	filter := "(  ) @snapshot:fe3c55a7-0ec9-4a27-bc4a-7d76eccd702a OR @replica:fe3c55a7-0ec9-4a27-bc4a-7d76eccd702a"

	beforeLimit := 1000
	beforeDate := time.Now().Format(time.RFC3339)
	anchorDate := ""
	afterDate := ""
	var afterLimit *int = nil

	fmt.Printf("开始订阅环境日志...\n")
	fmt.Printf("环境ID: %s\n", environmentID)
	fmt.Printf("过滤器: %s\n", filter)
	fmt.Printf("时间范围: %s 之前 %d 条日志\n", beforeDate, beforeLimit)
	fmt.Println("按 Ctrl+C 停止订阅")
	fmt.Println()

	// 创建上下文
	ctx := context.Background()

	// 订阅环境日志
	err = client.SubscribeEnvironmentLogs(ctx, environmentID, filter, beforeLimit, beforeDate, anchorDate, afterDate, afterLimit,
		func(timestamp, message, severity string, tags map[string]*string, attributes map[string]string) {
			// 打印日志信息
			fmt.Printf("[%s] %s: %s\n", timestamp, severity, message)

			// 打印标签信息（如果有）
			if len(tags) > 0 {
				fmt.Printf("  标签: ")
				for key, value := range tags {
					if value != nil {
						fmt.Printf("%s=%s ", key, *value)
					}
				}
				fmt.Println()
			}

			// 打印属性信息（如果有）
			if len(attributes) > 0 {
				fmt.Printf("  属性: ")
				for key, value := range attributes {
					fmt.Printf("%s=%s ", key, value)
				}
				fmt.Println()
			}
			fmt.Println()
		})

	if err != nil {
		log.Fatalf("订阅环境日志失败: %v", err)
	}
}
