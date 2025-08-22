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

// BackupSchedule 备份调度
type BackupSchedule struct {
	ID               string
	Name             string
	Cron             string
	Kind             string
	RetentionSeconds int64
	CreatedAt        string
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
