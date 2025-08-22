# 环境变更暂存功能示例

这个示例演示了如何使用 Railway Go SDK 暂存环境变更，包括服务变量的配置。

## 功能说明

### StageEnvironmentChanges 函数
用于暂存环境配置变更，支持复杂的配置结构，包括多个服务的变量配置。

### StageServiceVariables 函数
用于暂存单个服务的变量变更，提供便捷的接口。

## 使用方法

### 1. 设置环境变量

```bash
export RAILWAY_API_TOKEN="your_railway_api_token"
export ENVIRONMENT_ID="241e0310-96cb-4d94-9a70-cb8420991c2a"
export SERVICE_ID="e81eb2f2-35f1-4c84-b89a-6e8cb9effa03"
```

### 2. 运行示例

```bash
go run main.go
```

### 3. 输出示例

```
=== 示例 1: 暂存服务变量 ===
暂存变量: map[ss:12 ssss:sssssssssss]
✅ 暂存成功! 暂存ID: 83aea0da-6633-4c66-97b7-7613e052e1dd

=== 示例 2: 暂存完整环境配置 ===
暂存配置: {Services:map[e81eb2f2-35f1-4c84-b89a-6e8cb9effa03:{Variables:map[ss:{Value:12} ssss:{Value:sssssssssss}]}]}
✅ 暂存成功! 暂存ID: 83aea0da-6633-4c66-97b7-7613e052e1dd

=== 示例 3: 暂存多个服务的变量 ===
暂存多服务配置
✅ 暂存成功! 暂存ID: 83aea0da-6633-4c66-97b7-7613e052e1dd

=== 使用说明 ===
暂存功能说明:
  - 暂存操作不会立即应用变更
  - 需要后续调用提交操作来应用变更
  - 暂存ID用于后续的提交或回滚操作

支持的变量类型:
  - 字符串变量
  - 可以同时暂存多个服务的变量
  - 支持复杂的配置结构
```

## API 说明

### StageEnvironmentChanges 函数

```go
func (c *Client) StageEnvironmentChanges(ctx context.Context, environmentID string, payload StageEnvironmentConfig) (string, error)
```

**参数：**
- `ctx`: 上下文
- `environmentID`: 环境 ID
- `payload`: 环境配置结构

**返回值：**
- `string`: 暂存 ID
- `error`: 错误信息

### StageServiceVariables 函数

```go
func (c *Client) StageServiceVariables(ctx context.Context, environmentID, serviceID string, variables map[string]string) (string, error)
```

**参数：**
- `ctx`: 上下文
- `environmentID`: 环境 ID
- `serviceID`: 服务 ID
- `variables`: 变量键值对

**返回值：**
- `string`: 暂存 ID
- `error`: 错误信息

### 数据结构

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

## 使用示例

### 基本用法 - 暂存服务变量

```go
// 创建客户端
client, err := railway.New(
    railway.WithAPIToken(apiToken),
)
if err != nil {
    log.Fatalf("创建客户端失败: %v", err)
}

// 设置变量
variables := map[string]string{
    "ss":   "12",
    "ssss": "sssssssssss",
}

// 暂存服务变量
stageID, err := client.StageServiceVariables(ctx, environmentID, serviceID, variables)
if err != nil {
    log.Fatalf("暂存服务变量失败: %v", err)
}

fmt.Printf("暂存成功! 暂存ID: %s\n", stageID)
```

### 高级用法 - 暂存完整环境配置

```go
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

// 暂存环境变更
stageID, err := client.StageEnvironmentChanges(ctx, environmentID, payload)
if err != nil {
    log.Fatalf("暂存环境变更失败: %v", err)
}

fmt.Printf("暂存成功! 暂存ID: %s\n", stageID)
```

### 多服务配置

```go
// 暂存多个服务的变量
multiServicePayload := railway.StageEnvironmentConfig{
    Services: map[string]railway.StageServiceConfig{
        "service1-id": {
            Variables: map[string]railway.StageVariableConfig{
                "service1_var1": {Value: "value1"},
                "service1_var2": {Value: "value2"},
            },
        },
        "service2-id": {
            Variables: map[string]railway.StageVariableConfig{
                "service2_var1": {Value: "value3"},
            },
        },
    },
}

stageID, err := client.StageEnvironmentChanges(ctx, environmentID, multiServicePayload)
if err != nil {
    log.Fatalf("暂存多服务配置失败: %v", err)
}
```

## GraphQL 变更

### 变更定义

```graphql
mutation stageEnvironmentChanges($environmentId: String!, $payload: EnvironmentConfig!) {
  environmentStageChanges(environmentId: $environmentId, input: $payload) {
    id
  }
}
```

### 请求参数

- `environmentId`: 环境 ID（必需）
- `payload`: 环境配置（必需）

### 响应格式

```json
{
  "data": {
    "environmentStageChanges": {
      "id": "83aea0da-6633-4c66-97b7-7613e052e1dd"
    }
  }
}
```

## 注意事项

1. **暂存操作**: 暂存操作不会立即应用变更，需要后续提交
2. **暂存ID**: 返回的暂存ID用于后续的提交或回滚操作
3. **权限要求**: 需要相应的 API 权限
4. **环境ID**: 确保提供的环境ID是有效的
5. **服务ID**: 确保提供的服务ID在指定环境中存在

## 工作流程

1. **暂存变更**: 使用 `StageEnvironmentChanges` 或 `StageServiceVariables` 暂存变更
2. **获取暂存ID**: 保存返回的暂存ID
3. **提交变更**: 使用暂存ID调用提交操作来应用变更
4. **验证结果**: 检查变更是否成功应用

## 错误处理

```go
stageID, err := client.StageEnvironmentChanges(ctx, environmentID, payload)
if err != nil {
    // 处理不同类型的错误
    switch {
    case strings.Contains(err.Error(), "permission"):
        log.Fatal("权限不足，无法暂存变更")
    case strings.Contains(err.Error(), "not found"):
        log.Fatal("环境或服务不存在")
    default:
        log.Fatalf("暂存失败: %v", err)
    }
}
``` 