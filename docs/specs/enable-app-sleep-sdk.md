## Enable App Sleep SDK 设计文档（@railway-go/）

### 背景
- 能力：为指定服务开启/关闭自动休眠（sleepApplication）。
- 约束：当前 GraphQL 未提供单一“开关”字段的 mutation；需通过对 `EnvironmentConfig` 发起阶段性变更（stage changes），向 `services[serviceId]` 合并 `sleepApplication: true/false`，参考 `docs/specs/enable-app-sleeping.md`。

### GraphQL 方案回顾
1) 读取环境配置（含 staged patch）：
```graphql
query environmentConfig(
  $environmentId: String!
  $decryptVariables: Boolean
  $decryptPatchVariables: Boolean
) {
  environment(id: $environmentId) {
    id
    config(decryptVariables: $decryptVariables)
    serviceInstances { edges { node { id serviceId environmentId } } }
  }
  environmentStagedChanges(environmentId: $environmentId) {
    id
    patch(decryptVariables: $decryptPatchVariables)
  }
}
```

2) 暂存变更，将 `sleepApplication` 合并到 `services[serviceId]`：
```graphql
mutation stageEnvironmentChanges($environmentId: String!, $payload: EnvironmentConfig!) {
  environmentStageChanges(environmentId: $environmentId, input: $payload) { id }
}
```

示例 payload：
```json
{
  "services": {
    "<serviceId>": { "sleepApplication": true }
  }
}
```

（可选）提交暂存的变更：
```graphql
mutation environmentPatchCommitStaged($environmentId: String!, $message: String, $skipDeploys: Boolean) {
  environmentPatchCommitStaged(environmentId: $environmentId, message: $message, skipDeploys: $skipDeploys)
}
```

### SDK 接口设计

文件位置：`pkg/railway/sleep.go`

1) 基础开关接口（支持开启/关闭）：
```go
// SetServiceSleepApplication 暂存 sleepApplication 开关；可选立即提交。
type SetServiceSleepOptions struct {
    // 可选：为与 RestoreBackup 形态对齐，提供 ProjectID 以便做关联校验/审计
    ProjectID     *string
    EnvironmentID string
    ServiceID     string
    Enable        bool

    // 提交控制（可选）：若为 nil 则仅 stage 不提交
    Commit        *bool
    CommitMessage *string
    SkipDeploys   *bool
}

// 返回：stageID；若 Commit=true 则同时返回 commitID
func (c *Client) SetServiceSleepApplication(ctx context.Context, opts SetServiceSleepOptions) (stageID string, commitID *string, err error)
```

2) 语义化便捷方法：
```go
func (c *Client) EnableAppSleep(ctx context.Context, environmentID, serviceID string, commit bool) (stageID string, commitID *string, err error)
func (c *Client) DisableAppSleep(ctx context.Context, environmentID, serviceID string, commit bool) (stageID string, commitID *string, err error)
// 可选：含 ProjectID 的便捷方法
func (c *Client) EnableAppSleepInProject(ctx context.Context, projectID, environmentID, serviceID string, commit bool) (string, *string, error)
func (c *Client) DisableAppSleepInProject(ctx context.Context, projectID, environmentID, serviceID string, commit bool) (string, *string, error)
```

3) 幂等保障（可选）：
```go
// EnsureServiceSleepApplication 先读取 config；若已达期望值则直接返回（不发起变更）。
func (c *Client) EnsureServiceSleepApplication(ctx context.Context, environmentID, serviceID string, enable bool, commit bool) (changed bool, stageID *string, commitID *string, err error)
```

### 与现有结构的对齐与扩展
- 已有：`StageEnvironmentChanges(ctx, environmentID, payload)` 与相关结构定义位于 `pkg/railway/variables.go`。
- 扩展建议：为 `StageServiceConfig` 增加可选字段，以支持 typed payload（也可退回 raw map 实现）。

建议的结构变更：
```go
// variables.go
type StageServiceConfig struct {
    Variables        map[string]*StageVariableConfig `json:"variables,omitempty"`
    SleepApplication *bool                           `json:"sleepApplication,omitempty"`
}
```

若不改结构，亦可在实现中构造：
```go
payload := map[string]any{
    "services": map[string]any{
        serviceID: map[string]any{"sleepApplication": enable},
    },
}
```

### 参考实现要点（伪代码）
```go
func (c *Client) SetServiceSleepApplication(ctx context.Context, opts SetServiceSleepOptions) (string, *string, error) {
    if strings.TrimSpace(opts.EnvironmentID) == "" || strings.TrimSpace(opts.ServiceID) == "" {
        return "", nil, fmt.Errorf("environmentID/serviceID 不能为空")
    }

    // 若提供了 projectID，可进行关联校验（可选）：
    // - 读取 Project -> 验证该 service 是否属于该 project
    // - （必要时）解析/校验 environmentID 与 project 的一致性

    // 构造 payload（优先 typed，其次 raw map）
    payload := StageEnvironmentConfig{Services: map[string]StageServiceConfig{opts.ServiceID: {SleepApplication: &opts.Enable}}}
    stageID, err := c.StageEnvironmentChanges(ctx, opts.EnvironmentID, payload)
    if err != nil { return "", nil, err }

    // 可选提交
    if opts.Commit != nil && *opts.Commit {
        commitID, err := c.EnvironmentPatchCommitStaged(ctx, opts.EnvironmentID, opts.CommitMessage, opts.SkipDeploys)
        if err != nil { return stageID, nil, err }
        return stageID, &commitID, nil
    }
    return stageID, nil, nil
}
```

### 示例与验证（新增 @examples/enable_app_sleep）
文件：`examples/enable_app_sleep/main.go`

用途：演示开启 sleepApplication 并验证。

运行前置：
```bash
export RAILWAY_API_TOKEN="<your token>"
export ENVIRONMENT_ID="<env-id>"
export SERVICE_ID="<service-id>"
```

核心示例代码：
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "github.com/railwayapp/cli/pkg/railway"
)

func main() {
    token := os.Getenv("RAILWAY_API_TOKEN")
    envID := os.Getenv("ENVIRONMENT_ID")
    svcID := os.Getenv("SERVICE_ID")
    if token == "" || envID == "" || svcID == "" { log.Fatal("缺少必需环境变量") }

    cli, err := railway.New(railway.WithAPIToken(token))
    if err != nil { log.Fatal(err) }

    ctx := context.Background()

    // 开启并提交
    stageID, commitID, err := cli.EnableAppSleep(ctx, envID, svcID, true)
    if err != nil { log.Fatal(err) }
    fmt.Println("stage:", stageID, "commit:", deref(commitID))

    // 读取配置校验
    cfg, err := cli.GetEnvironmentConfig(ctx, envID, false, true)
    if err != nil { log.Fatal(err) }
    // 简单断言：从 staged patch 或 config 中读取 services[svcID].sleepApplication == true
    // （可按项目中类似 examples/variable 的 map 访问方式进行解析）
}

func deref(s *string) string { if s == nil { return "" }; return *s }
```

预期输出（示例）：
```
stage: 83aea0da-6633-4c66-97b7-7613e052e1dd commit: 0a77a1c7c9e1...
验证：services[<serviceId>].sleepApplication == true
```

### 错误与边界场景
- 非法 `environmentID/serviceID`：返回明确错误。
- 已是目标态：`EnsureServiceSleepApplication` 返回 `changed=false`，不触发写入。
- 提交失败：返回 stageID 便于追踪，同时返回提交错误。

### 兼容性与迁移
- 仅新增文件与方法，不影响现有 API。
- 若采用 typed 结构扩展，仅向 `StageServiceConfig` 增加可选字段，向下兼容。

### 测试计划
- 单测：构造 payload JSON 序列化校验；Ensure 分支幂等验证；提交分支错误回传。
- 集成：使用演示用例在实际环境验证启用与读取结果。

### 交付项一览
- 代码：
  - `pkg/railway/sleep.go` 新增 Set/Enable/Disable/Ensure 方法
  - （可选）`pkg/railway/variables.go` 的 `StageServiceConfig` 增加 `SleepApplication *bool`
- 示例：
  - `examples/enable_app_sleep/main.go`
- 文档：
  - 本文档（`docs/specs/enable-app-sleep-sdk.md`）


