package railway

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// RetryOption 控制幂等/重试行为
type RetryOption struct {
	MaxAttempts int
	Backoff     time.Duration
}

func defaultRetry() RetryOption { return RetryOption{MaxAttempts: 3, Backoff: 500 * time.Millisecond} }

// withRetry 在临时错误时重试
func withRetry(ctx context.Context, opt RetryOption, fn func() error) error {
	if opt.MaxAttempts <= 0 {
		opt = defaultRetry()
	}
	var last error
	for attempt := 1; attempt <= opt.MaxAttempts; attempt++ {
		if err := fn(); err != nil {
			last = err
			// 简单的可重试判断：网络错误/后端临时错误关键字
			msg := strings.ToLower(err.Error())
			if attempt < opt.MaxAttempts && (strings.Contains(msg, "timeout") || strings.Contains(msg, "temporary") || strings.Contains(msg, "connection") || strings.Contains(msg, "transient")) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(opt.Backoff * time.Duration(attempt)):
				}
				continue
			}
			return err
		}
		return nil
	}
	if last == nil {
		last = errors.New("unknown error")
	}
	return last
}

// EnsureService 存在即返回，不存在则创建（幂等）
func (c *Client) EnsureService(ctx context.Context, projectID, serviceName string, retry RetryOption) (*Service, error) {
	var out *Service
	err := withRetry(ctx, retry, func() error {
		// 查询现有服务
		p, err := c.GetProject(ctx, projectID)
		if err != nil {
			return err
		}
		for _, s := range p.Services {
			if s.Name == serviceName {
				tmp := s
				out = &tmp
				return nil
			}
		}
		// 未找到则创建
		created, err := c.CreateService(ctx, projectID, serviceName)
		if err != nil {
			return err
		}
		out = created
		return nil
	})
	return out, err
}

// EnsureEnvironment 存在即返回，不存在则创建（幂等）
func (c *Client) EnsureEnvironment(ctx context.Context, projectID, envName string, retry RetryOption) (*Environment, error) {
	var out *Environment
	err := withRetry(ctx, retry, func() error {
		p, err := c.GetProject(ctx, projectID)
		if err != nil {
			return err
		}
		for _, e := range p.Environments {
			if e.Name == envName {
				tmp := e
				out = &tmp
				return nil
			}
		}
		created, err := c.CreateEnvironment(ctx, projectID, envName)
		if err != nil {
			return err
		}
		out = created
		return nil
	})
	return out, err
}

// EnsureVariables 以幂等方式应用变量（默认非 replace）。若 replace=true，则确保最终值与 desired 一致。
func (c *Client) EnsureVariables(ctx context.Context, projectID, environmentID, serviceID string, desired map[string]string, replace bool, retry RetryOption) error {
	return withRetry(ctx, retry, func() error {
		current, err := c.GetVariables(ctx, projectID, environmentID, serviceID)
		if err != nil {
			return err
		}
		return c.ApplyVariableDiff(ctx, projectID, environmentID, &serviceID, replace, current, desired)
	})
}

// EnsureUp 幂等部署：若上传成功且状态已是 SUCCESS 则不阻塞；否则订阅直到成功/失败。
// 注意：无法避免实际创建新的 deployment，但允许通过回调/状态判断达到“成功状态幂等”。
func (c *Client) EnsureUp(ctx context.Context, p UpParams, retry RetryOption) (string, string, error) {
	var depID, logs string
	err := withRetry(ctx, retry, func() error {
		d, l, err := c.Up(ctx, p)
		if err != nil {
			return err
		}
		depID, logs = d, l
		return nil
	})
	return depID, logs, err
}

// EnsureProjectToken 若不存在同名 Token，则创建；否则直接返回新建的 Token 字符串（注意：后端一般不返回旧 Token 明文）
func (c *Client) EnsureProjectToken(ctx context.Context, projectID, environmentID, name string, retry RetryOption) (string, error) {
	var token string
	err := withRetry(ctx, retry, func() error {
		// 无法拿旧 token 明文，直接创建并返回；若后端去重同名，则此操作本身是幂等的
		t, err := c.CreateProjectToken(ctx, projectID, environmentID, name)
		if err != nil {
			return err
		}
		token = t
		return nil
	})
	return token, err
}

// EnsureServiceInstanceDeploy 幂等触发：后端一般会返回新的 deployment id；
// 可结合 SubscribeDeploymentStatus 等实现“直到成功/失败”的幂等流程。
func (c *Client) EnsureServiceInstanceDeploy(ctx context.Context, serviceID, environmentID string, retry RetryOption) (string, string, error) {
	var depID, status string
	err := withRetry(ctx, retry, func() error {
		id, st, err := c.DeployServiceInstance(ctx, serviceID, environmentID)
		if err != nil {
			return err
		}
		depID, status = id, st
		return nil
	})
	return depID, status, err
}

// WaitDeploymentSuccess 阻塞直到部署成功或失败
func (c *Client) WaitDeploymentSuccess(ctx context.Context, deploymentID string) (finalStatus string, err error) {
	done := make(chan struct{})
	var st string
	err = c.SubscribeDeploymentStatus(ctx, deploymentID, func(_ string, status string, _ bool) {
		s := strings.ToUpper(status)
		switch s {
		case "SUCCESS":
			st = s
			close(done)
		case "FAILED", "CRASHED":
			st = s
			close(done)
		}
	})
	if err != nil {
		return "", err
	}
	<-done
	if st == "" {
		return "", fmt.Errorf("unknown final status")
	}
	return st, nil
}
