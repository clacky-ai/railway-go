# 环境变更暂存功能实现

## 概述

在 `@/railway` 目录下的 `variables.go` 文件中成功实现了环境变更暂存功能，该功能允许用户暂存环境配置变更，包括服务变量的配置。

## 实现内容

### 1. GraphQL 变更定义

在 `internal/gql/mutations.go` 中添加了：

```go
// EnvironmentStageChanges GraphQL变更
const EnvironmentStageChangesMutation = `
mutation stageEnvironmentChanges($environmentId: String!, $payload: EnvironmentConfig!) {
  environmentStageChanges(environmentId: $environmentId, input: $payload) {
    id
  }
}
`

// EnvironmentStageChangesResponse 环境变更暂存响应
type EnvironmentStageChangesResponse struct {
	EnvironmentStageChanges struct {
		ID string `json:"id"`
	} `json:"environmentStageChanges"`
}
```

### 2. 数据结构定义

在 `pkg/railway/variables.go` 中添加了：

```go
// StageEnvironmentConfig 暂存环境配置结构
type StageEnvironmentConfig struct {
	Services map[string]StageServiceConfig `json:"services,omitempty"`
}

// StageServiceConfig 暂存服务配置结构
type StageServiceConfig struct {
	Variables map[string]StageVariableConfig `json:"variables,omitempty"`
}

// StageVariableConfig 暂存变量配置结构
type StageVariableConfig struct {
	Value string `json:"value"`
}
```

### 3. API 函数实现

在 `pkg/railway/variables.go` 中添加了：

```go
// StageEnvironmentChanges 暂存环境变更
func (c *Client) StageEnvironmentChanges(ctx context.Context, environmentID string, payload StageEnvironmentConfig) (string, error) {
	var resp igql.EnvironmentStageChangesResponse
	if err := c.gqlClient.Mutate(ctx, igql.EnvironmentStageChangesMutation, map[string]any{
		"environmentId": environmentID,
		"payload":       payload,
	}, &resp); err != nil {
		return "", err
	}
	return resp.EnvironmentStageChanges.ID, nil
}

// StageServiceVariables 暂存服务变量变更的便捷方法
func (c *Client) StageServiceVariables(ctx context.Context, environmentID, serviceID string, variables map[string]string) (string, error) {
	// 将简单的键值对转换为 StageVariableConfig 结构
	serviceVariables := make(map[string]StageVariableConfig)
	for key, value := range variables {
		serviceVariables[key] = StageVariableConfig{Value: value}
	}

	payload := StageEnvironmentConfig{
		Services: map[string]StageServiceConfig{
			serviceID: {
				Variables: serviceVariables,
			},
		},
	}

	return c.StageEnvironmentChanges(ctx, environmentID, payload)
}
```

## 功能特性

### 支持的配置类型

- **服务变量配置**: 支持为单个或多个服务配置变量
- **复杂配置结构**: 支持嵌套的服务和变量配置
- **JSON 序列化**: 完全支持 JSON 序列化和反序列化

### 使用场景

1. **暂存单个服务的变量**
   ```go
   variables := map[string]string{
       "ss":   "12",
       "ssss": "sssssssssss",
   }
   stageID, err := client.StageServiceVariables(ctx, environmentID, serviceID, variables)
   ```

2. **暂存完整环境配置**
   ```go
   payload := StageEnvironmentConfig{
       Services: map[string]StageServiceConfig{
           serviceID: {
               Variables: map[string]StageVariableConfig{
                   "ss":   {Value: "12"},
                   "ssss": {Value: "sssssssssss"},
               },
           },
       },
   }
   stageID, err := client.StageEnvironmentChanges(ctx, environmentID, payload)
   ```

3. **暂存多个服务的变量**
   ```go
   payload := StageEnvironmentConfig{
       Services: map[string]StageServiceConfig{
           "service1": {Variables: map[string]StageVariableConfig{"var1": {Value: "value1"}}},
           "service2": {Variables: map[string]StageVariableConfig{"var2": {Value: "value2"}}},
       },
   }
   ```

## 使用示例

### 基本用法

```go
// 创建客户端
client, err := railway.New(
    railway.WithAPIToken(apiToken),
)
if err != nil {
    log.Fatalf("创建客户端失败: %v", err)
}

// 暂存服务变量
variables := map[string]string{
    "ss":   "12",
    "ssss": "sssssssssss",
}

stageID, err := client.StageServiceVariables(ctx, environmentID, serviceID, variables)
if err != nil {
    log.Fatalf("暂存服务变量失败: %v", err)
}

fmt.Printf("暂存成功! 暂存ID: %s\n", stageID)
```

### 高级用法

```go
// 创建完整的配置结构
payload := StageEnvironmentConfig{
    Services: map[string]StageServiceConfig{
        serviceID: {
            Variables: map[string]StageVariableConfig{
                "ss":   {Value: "12"},
                "ssss": {Value: "sssssssssss"},
            },
        },
    },
}

// 暂存环境变更
stageID, err := client.StageEnvironmentChanges(ctx, environmentID, payload)
if err != nil {
    log.Fatalf("暂存环境变更失败: %v", err)
}

fmt.Printf("暂存成功! 暂存ID: %s\n", stageID)
```

## 示例和文档

### 1. 完整示例

创建了 `examples/environment_stage_changes/main.go` 示例文件，演示了：
- 暂存服务变量
- 暂存完整环境配置
- 暂存多个服务的变量

### 2. 文档

创建了 `examples/environment_stage_changes/README.md` 文档，包含：
- 功能说明
- 使用方法
- API 说明
- 注意事项

### 3. 测试

创建了 `pkg/railway/variables_test.go` 测试文件，验证了：
- 数据结构的正确性
- JSON 序列化和反序列化
- 各种配置场景

## 返回数据格式

该功能返回的数据格式与提供的 GraphQL 变更完全匹配：

```json
{
    "data": {
        "environmentStageChanges": {
            "id": "83aea0da-6633-4c66-97b7-7613e052e1dd"
        }
    }
}
```

## 工作流程

1. **暂存变更**: 使用 `StageEnvironmentChanges` 或 `StageServiceVariables` 暂存变更
2. **获取暂存ID**: 保存返回的暂存ID
3. **提交变更**: 使用暂存ID调用提交操作来应用变更
4. **验证结果**: 检查变更是否成功应用

## 注意事项

### 重要提醒

1. **暂存操作**: 暂存操作不会立即应用变更，需要后续提交
2. **暂存ID**: 返回的暂存ID用于后续的提交或回滚操作
3. **权限要求**: 需要相应的 API 权限
4. **环境ID**: 确保提供的环境ID是有效的
5. **服务ID**: 确保提供的服务ID在指定环境中存在

### 最佳实践

1. **先暂存后提交**: 建议先暂存变更，确认无误后再提交
2. **错误处理**: 始终检查返回的错误信息
3. **验证结果**: 提交后建议验证配置是否生效
4. **备份配置**: 重要配置变更前建议备份当前设置

## 验证

- ✅ 代码编译通过
- ✅ 测试通过
- ✅ 示例代码编译通过
- ✅ 文档完整

## 总结

成功实现了环境变更暂存功能，包括：

1. **GraphQL 变更定义** - 完整的变更和响应结构
2. **数据结构** - StageEnvironmentConfig 等结构体
3. **API 函数** - StageEnvironmentChanges 和 StageServiceVariables 方法
4. **示例代码** - 多种使用场景的示例
5. **文档** - 详细的使用说明和最佳实践
6. **测试** - 完整的数据结构和 JSON 序列化测试

该实现完全符合提供的 GraphQL 变更格式，并遵循了现有代码库的设计模式和编码规范。 