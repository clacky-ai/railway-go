package railway

import (
	"context"

	igql "github.com/railwayapp/cli/internal/gql"
)

// 卷备份调度类型常量
const (
	VolumeBackupScheduleDaily   = "DAILY"
	VolumeBackupScheduleWeekly  = "WEEKLY"
	VolumeBackupScheduleMonthly = "MONTHLY"
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

// BackupSchedule 备份调度
type BackupSchedule struct {
	ID               string
	Name             string
	Cron             string
	Kind             string
	RetentionSeconds int64
	CreatedAt        string
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

// GetVolumeBackupSchedules 获取卷备份调度列表
func (c *Client) GetVolumeBackupSchedules(ctx context.Context, volumeInstanceID string) ([]BackupSchedule, error) {
	var resp igql.VolumeInstanceBackupScheduleListResponse
	if err := c.gqlClient.QueryInternal(ctx, igql.VolumeInstanceBackupScheduleListQuery, map[string]any{"volumeInstanceId": volumeInstanceID}, &resp); err != nil {
		return nil, err
	}

	schedules := make([]BackupSchedule, len(resp.VolumeInstanceBackupScheduleList))
	for i, schedule := range resp.VolumeInstanceBackupScheduleList {
		schedules[i] = BackupSchedule{
			ID:               schedule.ID,
			Name:             schedule.Name,
			Cron:             schedule.Cron,
			Kind:             schedule.Kind,
			RetentionSeconds: schedule.RetentionSeconds,
			CreatedAt:        schedule.CreatedAt,
		}
	}
	return schedules, nil
}

// UpdateVolumeBackupSchedules 更新卷备份调度
func (c *Client) UpdateVolumeBackupSchedules(ctx context.Context, volumeInstanceID string, kinds []string) (bool, error) {
	var resp igql.VolumeInstanceBackupScheduleUpdateResponse
	if err := c.gqlClient.MutateInternal(ctx, igql.VolumeInstanceBackupScheduleUpdateMutation, map[string]any{
		"volumeInstanceId": volumeInstanceID,
		"kinds":            kinds,
	}, &resp); err != nil {
		return false, err
	}
	return resp.VolumeInstanceBackupScheduleUpdate, nil
}
