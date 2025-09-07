# Enable App Sleep 示例

这个示例演示了如何使用 Railway Go SDK 的应用休眠功能。应用休眠允许服务在无流量时自动休眠以节省资源，并在有新请求时自动唤醒。

## 功能特性

- **启用/禁用应用休眠**: 为指定服务开启或关闭自动休眠功能
- **灵活的提交控制**: 可以选择仅暂存变更或立即提交
- **项目关联校验**: 支持可选的项目ID参数进行关联校验
- **幂等操作**: 提供幂等方法确保达到期望状态
- **配置验证**: 读取并验证应用休眠配置是否生效

## 支持的接口

### 1. 便捷方法
- `EnableAppSleep(ctx, environmentID, serviceID, commit)` - 启用应用休眠
- `DisableAppSleep(ctx, environmentID, serviceID, commit)` - 禁用应用休眠
- `EnableAppSleepInProject(ctx, projectID, environmentID, serviceID, commit)` - 在指定项目中启用
- `DisableAppSleepInProject(ctx, projectID, environmentID, serviceID, commit)` - 在指定项目中禁用

### 2. 高级配置
```go
type SetServiceSleepOptions struct {
    ProjectID     *string  // 可选：项目ID，用于关联校验
    EnvironmentID string   // 必需：环境ID
    ServiceID     string   // 必需：服务ID
    Enable        bool     // 必需：是否启用休眠
    Commit        *bool    // 可选：是否立即提交，nil表示仅暂存
    CommitMessage *string  // 可选：提交消息
    SkipDeploys   *bool    // 可选：是否跳过部署
}

func SetServiceSleepApplication(ctx, opts) (stageID, commitID, error)
```

### 3. 幂等方法
```go
func EnsureServiceSleepApplication(ctx, environmentID, serviceID, enable, commit) (changed, stageID, commitID, error)
```

## 使用方法

### 1. 设置环境变量

```bash
export RAILWAY_API_TOKEN="your_railway_api_token"
export ENVIRONMENT_ID="your_environment_id"
export SERVICE_ID="your_service_id"
export PROJECT_ID="your_project_id"  # 可选
```

### 2. 运行示例

```bash
cd examples/enable_app_sleep
go run main.go
```

### 3. 输出示例

```
正在为服务 svc_123 启用应用休眠功能...
环境ID: env_456
项目ID: proj_789

=== 示例 1: 启用应用休眠 ===
✅ 启用成功! 暂存ID: 83aea0da-6633-4c66-97b7-7613e052e1dd, 提交ID: 0a77a1c7c9e1...

=== 示例 2: 验证配置 ===
从 staged changes 中读取到 sleepApplication: true
✅ 验证成功: 服务 svc_123 的应用休眠已启用

=== 示例 3: 使用高级选项禁用应用休眠（仅暂存） ===
✅ 暂存成功! 暂存ID: 74bea1cb-5644-4c77-98c8-8614f053f2ee (未提交)

=== 示例 4: 使用幂等方法确保应用休眠已启用 ===
ℹ️  状态未变更，应用休眠已处于期望状态

=== 使用说明 ===
应用休眠功能说明:
  - 启用后，服务在无流量时会自动休眠以节省资源
  - 有新请求时会自动唤醒服务
  - 适用于开发环境或低流量的生产环境

支持的操作:
  - EnableAppSleep: 启用应用休眠
  - DisableAppSleep: 禁用应用休眠
  - SetServiceSleepApplication: 高级配置选项
  - EnsureServiceSleepApplication: 幂等操作，确保达到期望状态
```

## API 说明

### EnableAppSleep 方法

```go
func (c *Client) EnableAppSleep(ctx context.Context, environmentID, serviceID string, commit bool) (stageID string, commitID *string, err error)
```

**参数：**
- `ctx`: 上下文
- `environmentID`: 环境ID
- `serviceID`: 服务ID
- `commit`: 是否立即提交变更

**返回值：**
- `stageID`: 暂存ID
- `commitID`: 提交ID（如果 commit=true）
- `error`: 错误信息

### SetServiceSleepApplication 方法

```go
func (c *Client) SetServiceSleepApplication(ctx context.Context, opts SetServiceSleepOptions) (stageID string, commitID *string, err error)
```

这是最灵活的方法，支持所有配置选项。

### EnsureServiceSleepApplication 方法

```go
func (c *Client) EnsureServiceSleepApplication(ctx context.Context, environmentID, serviceID string, enable bool, commit bool) (changed bool, stageID *string, commitID *string, err error)
```

幂等方法，会先检查当前状态，只有在需要变更时才执行操作。

## GraphQL 底层实现

该功能基于以下 GraphQL 操作：

1. **读取环境配置**:
```graphql
query environmentConfig($environmentId: String!) {
  environment(id: $environmentId) {
    config
    serviceInstances { edges { node { id serviceId environmentId } } }
  }
  environmentStagedChanges(environmentId: $environmentId) {
    patch
  }
}
```

2. **暂存变更**:
```graphql
mutation stageEnvironmentChanges($environmentId: String!, $payload: EnvironmentConfig!) {
  environmentStageChanges(environmentId: $environmentId, input: $payload) { id }
}
```

3. **提交变更**:
```graphql
mutation environmentPatchCommitStaged($environmentId: String!, $message: String, $skipDeploys: Boolean) {
  environmentPatchCommitStaged(environmentId: $environmentId, message: $message, skipDeploys: $skipDeploys)
}
```

## 注意事项

1. **权限要求**: 需要相应的 Railway API 权限
2. **环境ID**: 确保提供的环境ID是有效的
3. **服务ID**: 确保提供的服务ID在指定环境中存在
4. **暂存与提交**: 
   - 暂存操作不会立即应用变更
   - 需要提交操作才能使变更生效
   - 可以选择在同一次调用中完成暂存和提交
5. **配置生效**: 变更提交后可能需要一些时间才能完全生效

## 错误处理

```go
stageID, commitID, err := client.EnableAppSleep(ctx, envID, svcID, true)
if err != nil {
    // 处理不同类型的错误
    switch {
    case strings.Contains(err.Error(), "权限"):
        log.Fatal("权限不足，无法启用应用休眠")
    case strings.Contains(err.Error(), "not found"):
        log.Fatal("环境或服务不存在")
    default:
        log.Fatalf("启用失败: %v", err)
    }
}
```

## 最佳实践

1. **先暂存后提交**: 对于重要配置，建议先暂存变更，确认无误后再提交
2. **使用幂等方法**: 在自动化脚本中使用 `EnsureServiceSleepApplication` 避免重复操作
3. **验证结果**: 变更后建议读取配置验证是否生效
4. **错误处理**: 始终检查返回的错误信息并进行适当处理
5. **环境隔离**: 在开发环境中测试配置变更，确认无误后再应用到生产环境
