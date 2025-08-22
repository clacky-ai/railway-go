# 卷备份删除示例

这个示例演示了如何使用 Railway Go SDK 删除卷备份。

## 功能说明

- 支持批量删除多个备份
- 返回工作流ID用于跟踪删除进度
- 删除操作是异步的

## 使用方法

### 1. 设置环境变量

```bash
export RAILWAY_API_TOKEN="your_api_token_here"
```

### 2. 运行示例

```bash
# 删除单个备份
go run main.go <volumeInstanceID> <backupID>

# 删除多个备份
go run main.go <volumeInstanceID> <backupID1> <backupID2> <backupID3>
```

### 3. 示例

```bash
# 删除单个备份
go run main.go d9e8972e-5757-447a-a095-dcf40865d227 851408aa-579d-4d57-9075-e3e86f8c45a4

# 删除多个备份
go run main.go d9e8972e-5757-447a-a095-dcf40865d227 \
  851408aa-579d-4d57-9075-e3e86f8c45a4 \
  851408aa-579d-4d57-9075-e3e86f8c45a5 \
  851408aa-579d-4d57-9075-e3e86f8c45a6
```

## API 说明

### DeleteVolumeBackups 方法

```go
func (c *Client) DeleteVolumeBackups(ctx context.Context, volumeInstanceID string, volumeInstanceBackupIDs []string) (string, error)
```

**参数：**
- `ctx`: 上下文
- `volumeInstanceID`: 卷实例ID
- `volumeInstanceBackupIDs`: 要删除的备份ID列表

**返回值：**
- `workflowID`: 工作流ID，用于跟踪删除进度
- `error`: 错误信息

## 注意事项

1. 删除操作是异步的，API调用成功后立即返回工作流ID
2. 可以通过工作流ID查询删除操作的进度
3. 删除操作不可逆，请谨慎操作
4. 确保有足够的权限执行删除操作

## 完整示例

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

	// 删除备份
	ctx := context.Background()
	volumeInstanceID := "d9e8972e-5757-447a-a095-dcf40865d227"
	backupIDs := []string{"851408aa-579d-4d57-9075-e3e86f8c45a4"}

	workflowID, err := client.DeleteVolumeBackups(ctx, volumeInstanceID, backupIDs)
	if err != nil {
		log.Fatalf("删除备份失败: %v", err)
	}

	fmt.Printf("删除备份成功！工作流ID: %s\n", workflowID)
}
```

## GraphQL Mutation

底层使用的GraphQL mutation：

```graphql
mutation volumeInstanceBackupBatchDelete($volumeInstanceId: String!, $volumeInstanceBackupIds: [String!]!) {
  volumeInstanceBackupBatchDelete(
    volumeInstanceId: $volumeInstanceId
    volumeInstanceBackupIds: $volumeInstanceBackupIds
  ) {
    workflowId
  }
}
``` 