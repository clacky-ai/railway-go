package railway

import (
	"context"
	"fmt"
	"strings"
)

// SetServiceSleepOptions 设置服务休眠选项
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

// SetServiceSleepApplication 暂存 sleepApplication 开关；可选立即提交
// 返回：stageID；若 Commit=true 则同时返回 commitID
func (c *Client) SetServiceSleepApplication(ctx context.Context, opts SetServiceSleepOptions) (stageID string, commitID *string, err error) {
	if strings.TrimSpace(opts.EnvironmentID) == "" || strings.TrimSpace(opts.ServiceID) == "" {
		return "", nil, fmt.Errorf("environmentID/serviceID required")
	}

	// 若提供了 projectID，可进行关联校验（可选）：
	// - 读取 Project -> 验证该 service 是否属于该 project
	// - （必要时）解析/校验 environmentID 与 project 的一致性
	if opts.ProjectID != nil && strings.TrimSpace(*opts.ProjectID) != "" {
		// 这里可以添加校验逻辑，暂时跳过
		// TODO: 实现项目和服务的关联校验
	}

	// 构造 payload（使用 typed 结构）
	payload := StageEnvironmentConfig{
		Services: map[string]StageServiceConfig{
			opts.ServiceID: {
				Deploy: &StageDeployConfig{
					SleepApplication: &opts.Enable,
				},
			},
		},
	}

	stageID, err = c.StageEnvironmentChanges(ctx, opts.EnvironmentID, payload)
	if err != nil {
		return "", nil, err
	}

	// 可选提交
	if opts.Commit != nil && *opts.Commit {
		commitID, err := c.EnvironmentPatchCommitStaged(ctx, opts.EnvironmentID, opts.CommitMessage, opts.SkipDeploys)
		if err != nil {
			return stageID, nil, err
		}
		return stageID, &commitID, nil
	}
	return stageID, nil, nil
}

// EnableAppSleep 启用应用休眠
func (c *Client) EnableAppSleep(ctx context.Context, environmentID, serviceID string, commit bool) (stageID string, commitID *string, err error) {
	var skipDeploys = false
	opts := SetServiceSleepOptions{
		EnvironmentID: environmentID,
		ServiceID:     serviceID,
		Enable:        true,
		Commit:        &commit,
		SkipDeploys:   &skipDeploys,
	}
	return c.SetServiceSleepApplication(ctx, opts)
}

// DisableAppSleep 禁用应用休眠
func (c *Client) DisableAppSleep(ctx context.Context, environmentID, serviceID string, commit bool) (stageID string, commitID *string, err error) {
	var skipDeploys = false
	opts := SetServiceSleepOptions{
		EnvironmentID: environmentID,
		ServiceID:     serviceID,
		Enable:        false,
		Commit:        &commit,
		SkipDeploys:   &skipDeploys,
	}
	return c.SetServiceSleepApplication(ctx, opts)
}

// EnableAppSleepInProject 在指定项目中启用应用休眠
func (c *Client) EnableAppSleepInProject(ctx context.Context, projectID, environmentID, serviceID string, commit bool) (string, *string, error) {
	opts := SetServiceSleepOptions{
		ProjectID:     &projectID,
		EnvironmentID: environmentID,
		ServiceID:     serviceID,
		Enable:        true,
		Commit:        &commit,
	}
	return c.SetServiceSleepApplication(ctx, opts)
}

// DisableAppSleepInProject 在指定项目中禁用应用休眠
func (c *Client) DisableAppSleepInProject(ctx context.Context, projectID, environmentID, serviceID string, commit bool) (string, *string, error) {
	opts := SetServiceSleepOptions{
		ProjectID:     &projectID,
		EnvironmentID: environmentID,
		ServiceID:     serviceID,
		Enable:        false,
		Commit:        &commit,
	}
	return c.SetServiceSleepApplication(ctx, opts)
}

// EnsureServiceSleepApplication 先读取 config；若已达期望值则直接返回（不发起变更）
func (c *Client) EnsureServiceSleepApplication(ctx context.Context, environmentID, serviceID string, enable bool, commit bool) (changed bool, stageID *string, commitID *string, err error) {
	// 读取当前环境配置
	config, err := c.GetEnvironmentConfig(ctx, environmentID, false, true)
	if err != nil {
		return false, nil, nil, fmt.Errorf("failed to get environment config: %v", err)
	}

	// 检查当前 sleepApplication 状态
	currentEnabled := false

	// 首先检查 staged changes 中的配置
	if config.EnvironmentStagedChanges.Patch != nil {
		if services, ok := config.EnvironmentStagedChanges.Patch["services"].(map[string]interface{}); ok {
			if serviceConfig, ok := services[serviceID].(map[string]interface{}); ok {
				if deployConfig, ok := serviceConfig["deploy"].(map[string]interface{}); ok {
					if sleepApp, ok := deployConfig["sleepApplication"].(bool); ok {
						currentEnabled = sleepApp
					}
				}
			}
		}
	}

	// 如果 staged changes 中没有，则检查当前配置
	if config.EnvironmentStagedChanges.Patch == nil || !currentEnabled {
		if config.Environment.Config != nil {
			if services, ok := config.Environment.Config["services"].(map[string]interface{}); ok {
				if serviceConfig, ok := services[serviceID].(map[string]interface{}); ok {
					if deployConfig, ok := serviceConfig["deploy"].(map[string]interface{}); ok {
						if sleepApp, ok := deployConfig["sleepApplication"].(bool); ok {
							currentEnabled = sleepApp
						}
					}
				}
			}
		}
	}

	// 如果已经是期望状态，直接返回
	if currentEnabled == enable {
		return false, nil, nil, nil
	}

	// 否则进行变更
	opts := SetServiceSleepOptions{
		EnvironmentID: environmentID,
		ServiceID:     serviceID,
		Enable:        enable,
		Commit:        &commit,
	}

	stage, commitResult, err := c.SetServiceSleepApplication(ctx, opts)
	if err != nil {
		return false, nil, nil, err
	}

	return true, &stage, commitResult, nil
}
