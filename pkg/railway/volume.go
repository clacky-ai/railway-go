package railway

import (
	"context"

	igql "github.com/railwayapp/cli/internal/gql"
)

type Backup struct {
	ID           string
	Name         string
	CreatedAt    string
	ExpiresAt    *string
	ExternalID   *string
	UsedMB       *int64
	ReferencedMB *int64
	ScheduleID   *string
}

// RollbackResult 回滚结果
type RollbackResult struct {
	ID           string
	Status       string
	DeploymentID string
}

// GetAllVolumeBackups 获取所有卷备份
func (c *Client) GetAllVolumeBackups(ctx context.Context, volumeInstanceID string) ([]Backup, error) {
	var resp igql.VolumeInstanceBackupListResponse
	if err := c.gqlClient.Query(ctx, igql.VolumeInstanceBackupListQuery, map[string]any{"volumeInstanceId": volumeInstanceID}, &resp); err != nil {
		return nil, err
	}

	backups := make([]Backup, len(resp.VolumeInstanceBackupList))
	for i, backup := range resp.VolumeInstanceBackupList {
		backups[i] = Backup{
			ID:           backup.ID,
			Name:         backup.Name,
			CreatedAt:    backup.CreatedAt,
			ExpiresAt:    backup.ExpiresAt,
			ExternalID:   backup.ExternalID,
			UsedMB:       backup.UsedMB,
			ReferencedMB: backup.ReferencedMB,
			ScheduleID:   backup.ScheduleID,
		}
	}
	return backups, nil
}

// CreateVolumeBackup 通过卷实例ID创建备份，返回工作流ID
func (c *Client) CreateVolumeBackup(ctx context.Context, volumeInstanceID string) (string, error) {
	var resp igql.VolumeInstanceBackupCreateResponse
	if err := c.gqlClient.Mutate(ctx, igql.VolumeInstanceBackupCreateMutation, map[string]any{"volumeInstanceId": volumeInstanceID}, &resp); err != nil {
		return "", err
	}
	return resp.VolumeInstanceBackupCreate.WorkflowID, nil
}

// RestoreVolumeBackup 通过卷实例ID和备份ID恢复备份，返回工作流ID
func (c *Client) RestoreVolumeBackup(ctx context.Context, volumeInstanceID, volumeInstanceBackupID string) (string, error) {
	var resp igql.VolumeInstanceBackupRestoreResponse
	if err := c.gqlClient.Mutate(ctx, igql.VolumeInstanceBackupRestoreMutation, map[string]any{
		"volumeInstanceId":       volumeInstanceID,
		"volumeInstanceBackupId": volumeInstanceBackupID,
	}, &resp); err != nil {
		return "", err
	}
	return resp.VolumeInstanceBackupRestore.WorkflowID, nil
}
