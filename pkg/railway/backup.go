package railway

import (
	"context"
	"fmt"

	igql "github.com/railwayapp/cli/internal/gql"
)

// Backup 备份信息
type Backup struct {
	ID        string
	Name      string
	CreatedAt string
	Status    string
	Size      *int64
	Service   struct {
		ID   string
		Name string
	}
}

// BackupList 备份列表（支持分页）
type BackupList struct {
	Backups   []Backup
	HasNext   bool
	EndCursor *string
}

// ListBackups 获取服务的备份列表
// serviceID: 服务ID
// after: 分页游标，可选
func (c *Client) ListBackups(ctx context.Context, serviceID string, after *string) (*BackupList, error) {
	params := map[string]any{"serviceId": serviceID}
	if after != nil {
		params["after"] = *after
	}

	var resp igql.BackupsResponse
	if err := c.gqlClient.Query(ctx, igql.BackupsQuery, params, &resp); err != nil {
		return nil, err
	}

	backups := make([]Backup, 0, len(resp.Backups.Edges))
	for _, edge := range resp.Backups.Edges {
		backup := Backup{
			ID:        edge.Node.ID,
			Name:      edge.Node.Name,
			CreatedAt: edge.Node.CreatedAt,
			Status:    edge.Node.Status,
			Size:      edge.Node.Size,
		}
		backup.Service.ID = edge.Node.Service.ID
		backup.Service.Name = edge.Node.Service.Name
		backups = append(backups, backup)
	}

	return &BackupList{
		Backups:   backups,
		HasNext:   resp.Backups.PageInfo.HasNextPage,
		EndCursor: resp.Backups.PageInfo.EndCursor,
	}, nil
}

// CreateBackup 创建服务备份
// serviceID: 服务ID
// name: 备份名称，可选
func (c *Client) CreateBackup(ctx context.Context, serviceID, name string) (*Backup, error) {
	input := igql.BackupCreateInput{
		ServiceID: serviceID,
		Name:      name,
	}

	var resp igql.BackupCreateResponse
	if err := c.gqlClient.Mutate(ctx, igql.BackupCreateMutation, map[string]any{"input": input}, &resp); err != nil {
		return nil, err
	}

	return &Backup{
		ID:        resp.BackupCreate.ID,
		Name:      resp.BackupCreate.Name,
		CreatedAt: resp.BackupCreate.CreatedAt,
		Status:    resp.BackupCreate.Status,
	}, nil
}

// RollbackBackup 回滚到指定备份
// backupID: 备份ID
func (c *Client) RollbackBackup(ctx context.Context, backupID string) (*RollbackResult, error) {
	input := igql.BackupRollbackInput{
		BackupID: backupID,
	}

	var resp igql.BackupRollbackResponse
	if err := c.gqlClient.Mutate(ctx, igql.BackupRollbackMutation, map[string]any{"input": input}, &resp); err != nil {
		return nil, err
	}

	return &RollbackResult{
		ID:           resp.BackupRollback.ID,
		Status:       resp.BackupRollback.Status,
		DeploymentID: resp.BackupRollback.DeploymentID,
	}, nil
}

// RollbackResult 回滚结果
type RollbackResult struct {
	ID           string
	Status       string
	DeploymentID string
}

// GetAllBackups 获取所有备份（自动处理分页）
func (c *Client) GetAllBackups(ctx context.Context, serviceID string) ([]Backup, error) {
	var allBackups []Backup
	var after *string

	for {
		list, err := c.ListBackups(ctx, serviceID, after)
		if err != nil {
			return nil, fmt.Errorf("获取备份列表失败: %w", err)
		}

		allBackups = append(allBackups, list.Backups...)

		if !list.HasNext {
			break
		}
		after = list.EndCursor
	}

	return allBackups, nil
}
