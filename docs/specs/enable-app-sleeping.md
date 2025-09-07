
#### 启用 App Sleeping 
官方说明：`https://docs.railway.com/reference/app-sleeping`

当前 GraphQL 未提供单一“开关”字段的 mutation；实践方式为：
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
2) 构造并暂存变更，向 `EnvironmentConfig` 的 `services[serviceId]` 合并 sleep 配置（示例结构，具体键名以读取到的 `config` 为准进行合并覆盖）：
```graphql
mutation stageEnvironmentChanges($environmentId: String!, $payload: EnvironmentConfig!) {
  environmentStageChanges(environmentId: $environmentId, input: $payload) { id }
}
```
示例 payload：
```json
{
  "services": {
    "<serviceId>": { "sleepApplication": true },
  }
}
```