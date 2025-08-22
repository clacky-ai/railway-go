package railway

import (
	"testing"
)

func TestVolumeBackupScheduleConstants(t *testing.T) {
	// 测试卷备份调度类型常量
	if VolumeBackupScheduleDaily != "DAILY" {
		t.Errorf("期望 VolumeBackupScheduleDaily 为 'DAILY'，实际为 '%s'", VolumeBackupScheduleDaily)
	}

	if VolumeBackupScheduleWeekly != "WEEKLY" {
		t.Errorf("期望 VolumeBackupScheduleWeekly 为 'WEEKLY'，实际为 '%s'", VolumeBackupScheduleWeekly)
	}

	if VolumeBackupScheduleMonthly != "MONTHLY" {
		t.Errorf("期望 VolumeBackupScheduleMonthly 为 'MONTHLY'，实际为 '%s'", VolumeBackupScheduleMonthly)
	}
}

func TestBackupScheduleStruct(t *testing.T) {
	// 测试 BackupSchedule 结构体
	schedule := BackupSchedule{
		ID:               "test-id",
		Name:             "Test Schedule",
		Cron:             "0 0 * * *",
		Kind:             VolumeBackupScheduleDaily,
		RetentionSeconds: 86400,
		CreatedAt:        "2025-01-01T00:00:00Z",
	}

	// 验证字段值
	if schedule.ID != "test-id" {
		t.Errorf("期望 ID 为 'test-id'，实际为 '%s'", schedule.ID)
	}

	if schedule.Name != "Test Schedule" {
		t.Errorf("期望 Name 为 'Test Schedule'，实际为 '%s'", schedule.Name)
	}

	if schedule.Cron != "0 0 * * *" {
		t.Errorf("期望 Cron 为 '0 0 * * *'，实际为 '%s'", schedule.Cron)
	}

	if schedule.Kind != VolumeBackupScheduleDaily {
		t.Errorf("期望 Kind 为 '%s'，实际为 '%s'", VolumeBackupScheduleDaily, schedule.Kind)
	}

	if schedule.RetentionSeconds != 86400 {
		t.Errorf("期望 RetentionSeconds 为 86400，实际为 %d", schedule.RetentionSeconds)
	}

	if schedule.CreatedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("期望 CreatedAt 为 '2025-01-01T00:00:00Z'，实际为 '%s'", schedule.CreatedAt)
	}
}
