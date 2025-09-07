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
	token := os.Getenv("RAILWAY_API_TOKEN")
	envID := os.Getenv("ENVIRONMENT_ID")
	svcID := os.Getenv("SERVICE_ID")
	projectID := os.Getenv("PROJECT_ID") // 可选

	if token == "" {
		log.Fatal("请设置 RAILWAY_API_TOKEN 环境变量")
	}
	if envID == "" {
		log.Fatal("请设置 ENVIRONMENT_ID 环境变量")
	}
	if svcID == "" {
		log.Fatal("请设置 SERVICE_ID 环境变量")
	}

	// 创建 Railway 客户端
	cli, err := railway.New(railway.WithAPIToken(token))
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	fmt.Printf("正在为服务 %s 启用应用休眠功能...\n", svcID)
	fmt.Printf("环境ID: %s\n", envID)
	if projectID != "" {
		fmt.Printf("项目ID: %s\n", projectID)
	}

	// 依次调用测试用例,您可以按需注释掉不需要的测试
	testEnableAppSleep(ctx, cli, envID, svcID)

	testDisableAppSleep(ctx, cli, envID, svcID)
	// testVerifyConfig(ctx, cli, envID, svcID)
	// testDisableAppSleepStagingOnly(ctx, cli, envID, svcID, projectID)
	// testEnsureAppSleepEnabled(ctx, cli, envID, svcID)

	fmt.Println("\n=== 使用说明 ===")
	fmt.Println("应用休眠功能说明:")
	fmt.Println("  - 启用后，服务在无流量时会自动休眠以节省资源")
	fmt.Println("  - 有新请求时会自动唤醒服务")
	fmt.Println("  - 适用于开发环境或低流量的生产环境")
	fmt.Println("\n支持的操作:")
	fmt.Println("  - EnableAppSleep: 启用应用休眠")
	fmt.Println("  - DisableAppSleep: 禁用应用休眠")
	fmt.Println("  - SetServiceSleepApplication: 高级配置选项")
	fmt.Println("  - EnsureServiceSleepApplication: 幂等操作，确保达到期望状态")
}

// testEnableAppSleep 示例 1: 使用便捷方法启用应用休眠并提交
func testEnableAppSleep(ctx context.Context, cli *railway.Client, envID, svcID string) {
	fmt.Println("\n=== 示例 1: 启用应用休眠 ===")
	stageID, commitID, err := cli.EnableAppSleep(ctx, envID, svcID, true)
	if err != nil {
		log.Fatalf("启用应用休眠失败: %v", err)
	}
	fmt.Printf("✅ 启用成功! 暂存ID: %s", stageID)
	if commitID != nil {
		fmt.Printf(", 提交ID: %s", *commitID)
	}
	fmt.Println()
}

func testDisableAppSleep(ctx context.Context, cli *railway.Client, envID, svcID string) {
	fmt.Println("\n=== 示例 2: 禁用应用休眠 ===")
	stageID, commitID, err := cli.DisableAppSleep(ctx, envID, svcID, true)
	if err != nil {
		log.Fatalf("禁用应用休眠失败: %v", err)
	}
	fmt.Printf("✅ 禁用成功! 暂存ID: %s", stageID)
	if commitID != nil {
		fmt.Printf(", 提交ID: %s", *commitID)
	}
	fmt.Println()
}

// testVerifyConfig 示例 2: 读取配置验证
func testVerifyConfig(ctx context.Context, cli *railway.Client, envID, svcID string) {
	fmt.Println("\n=== 示例 2: 验证配置 ===")
	cfg, err := cli.GetEnvironmentConfig(ctx, envID, false, true)
	if err != nil {
		log.Fatalf("读取环境配置失败: %v", err)
	}

	// 检查 staged changes 中的配置
	sleepEnabled := false
	found := false

	if cfg.EnvironmentStagedChanges.Patch != nil {
		if services, ok := cfg.EnvironmentStagedChanges.Patch["services"].(map[string]interface{}); ok {
			if serviceConfig, ok := services[svcID].(map[string]interface{}); ok {
				if sleepApp, ok := serviceConfig["sleepApplication"].(bool); ok {
					sleepEnabled = sleepApp
					found = true
					fmt.Printf("从 staged changes 中读取到 sleepApplication: %t\n", sleepEnabled)
				}
			}
		}
	}

	// 如果 staged changes 中没有，检查当前配置
	if !found && cfg.Environment.Config != nil {
		if services, ok := cfg.Environment.Config["services"].(map[string]interface{}); ok {
			if serviceConfig, ok := services[svcID].(map[string]interface{}); ok {
				if sleepApp, ok := serviceConfig["sleepApplication"].(bool); ok {
					sleepEnabled = sleepApp
					found = true
					fmt.Printf("从当前配置中读取到 sleepApplication: %t\n", sleepEnabled)
				}
			}
		}
	}

	if !found {
		fmt.Printf("⚠️  未在配置中找到 sleepApplication 设置，可能尚未生效\n")
	} else if sleepEnabled {
		fmt.Printf("✅ 验证成功: 服务 %s 的应用休眠已启用\n", svcID)
	} else {
		fmt.Printf("❌ 验证失败: 服务 %s 的应用休眠未启用\n", svcID)
	}
}

// testDisableAppSleepStagingOnly 示例 3: 使用高级选项禁用应用休眠（仅暂存）
func testDisableAppSleepStagingOnly(ctx context.Context, cli *railway.Client, envID, svcID, projectID string) {
	fmt.Println("\n=== 示例 3: 使用高级选项禁用应用休眠（仅暂存） ===")
	opts := railway.SetServiceSleepOptions{
		EnvironmentID: envID,
		ServiceID:     svcID,
		Enable:        false,
		// Commit 为 nil，表示仅暂存不提交
	}

	if projectID != "" {
		opts.ProjectID = &projectID
	}

	stageID2, commitID2, err := cli.SetServiceSleepApplication(ctx, opts)
	if err != nil {
		log.Fatalf("设置应用休眠失败: %v", err)
	}
	fmt.Printf("✅ 暂存成功! 暂存ID: %s", stageID2)
	if commitID2 != nil {
		fmt.Printf(", 提交ID: %s", *commitID2)
	} else {
		fmt.Printf(" (未提交)")
	}
	fmt.Println()
}

// testEnsureAppSleepEnabled 示例 4: 使用幂等方法确保应用休眠已启用
func testEnsureAppSleepEnabled(ctx context.Context, cli *railway.Client, envID, svcID string) {
	fmt.Println("\n=== 示例 4: 使用幂等方法确保应用休眠已启用 ===")
	changed, stageID3, commitID3, err := cli.EnsureServiceSleepApplication(ctx, envID, svcID, true, true)
	if err != nil {
		log.Fatalf("确保应用休眠状态失败: %v", err)
	}

	if changed {
		fmt.Printf("✅ 状态已变更! 暂存ID: %s", deref(stageID3))
		if commitID3 != nil {
			fmt.Printf(", 提交ID: %s", *commitID3)
		}
		fmt.Println()
	} else {
		fmt.Printf("ℹ️  状态未变更，应用休眠已处于期望状态\n")
	}
}

// deref 安全地解引用字符串指针
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
