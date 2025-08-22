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

	environmentID := os.Getenv("ENVIRONMENT_ID")
	if environmentID == "" {
		log.Fatal("请设置环境变量 ENVIRONMENT_ID")
	}

	serviceID := os.Getenv("SERVICE_ID")
	if serviceID == "" {
		log.Fatal("请设置环境变量 SERVICE_ID")
	}

	// 创建 Railway 客户端
	client, err := railway.New(
		railway.WithAPIToken(apiToken),
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	// 示例 1: 使用便捷方法暂存服务变量
	fmt.Println("=== 示例 1: 暂存服务变量 ===")

	variables := map[string]string{
		"ss":   "12",
		"ssss": "sssssssssss",
	}

	fmt.Printf("暂存变量: %v\n", variables)

	stageID, err := client.StageServiceVariables(ctx, environmentID, serviceID, variables)
	if err != nil {
		log.Fatalf("暂存服务变量失败: %v", err)
	}

	fmt.Printf("✅ 暂存成功! 暂存ID: %s\n", stageID)

	// 示例 2: 使用完整的配置结构暂存环境变更
	fmt.Println("\n=== 示例 2: 暂存完整环境配置 ===")

	// 创建完整的配置结构
	payload := railway.StageEnvironmentConfig{
		Services: map[string]railway.StageServiceConfig{
			serviceID: {
				Variables: map[string]railway.StageVariableConfig{
					"ss":   {Value: "12"},
					"ssss": {Value: "sssssssssss"},
				},
			},
		},
	}

	fmt.Printf("暂存配置: %+v\n", payload)

	stageID2, err := client.StageEnvironmentChanges(ctx, environmentID, payload)
	if err != nil {
		log.Fatalf("暂存环境变更失败: %v", err)
	}

	fmt.Printf("✅ 暂存成功! 暂存ID: %s\n", stageID2)

	// 示例 3: 暂存多个服务的变量
	fmt.Println("\n=== 示例 3: 暂存多个服务的变量 ===")

	multiServicePayload := railway.StageEnvironmentConfig{
		Services: map[string]railway.StageServiceConfig{
			serviceID: {
				Variables: map[string]railway.StageVariableConfig{
					"service1_var1": {Value: "value1"},
					"service1_var2": {Value: "value2"},
				},
			},
			"another-service-id": {
				Variables: map[string]railway.StageVariableConfig{
					"service2_var1": {Value: "value3"},
				},
			},
		},
	}

	fmt.Printf("暂存多服务配置\n")

	stageID3, err := client.StageEnvironmentChanges(ctx, environmentID, multiServicePayload)
	if err != nil {
		log.Fatalf("暂存多服务配置失败: %v", err)
	}

	fmt.Printf("✅ 暂存成功! 暂存ID: %s\n", stageID3)

	fmt.Println("\n=== 使用说明 ===")
	fmt.Println("暂存功能说明:")
	fmt.Println("  - 暂存操作不会立即应用变更")
	fmt.Println("  - 需要后续调用提交操作来应用变更")
	fmt.Println("  - 暂存ID用于后续的提交或回滚操作")
	fmt.Println("\n支持的变量类型:")
	fmt.Println("  - 字符串变量")
	fmt.Println("  - 可以同时暂存多个服务的变量")
	fmt.Println("  - 支持复杂的配置结构")
}
