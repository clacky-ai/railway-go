package railway

import (
	"context"
	"encoding/json"

	iclient "github.com/railwayapp/cli/internal/client"
	igql "github.com/railwayapp/cli/internal/gql"
)

// 订阅封装
func (c *Client) SubscribeBuildLogs(ctx context.Context, deploymentID string, filter string, limit int, onLog func(timestamp, message string, attributes map[string]string)) error {
	vars := map[string]any{"deploymentId": deploymentID, "filter": filter, "limit": limit}
	return iclient.Subscribe(ctx, c.cfg, igql.BuildLogsSub, vars,
		func(data json.RawMessage) {
			var pl igql.BuildLogsPayload
			if err := json.Unmarshal(data, &pl); err == nil {
				for _, l := range pl.BuildLogs {
					attrs := map[string]string{}
					for _, a := range l.Attributes {
						attrs[a.Key] = a.Value
					}

					if onLog != nil {
						onLog(l.Timestamp, l.Message, attrs)
					}
				}
			}
		}, func(err error) {},
	)
}

func (c *Client) SubscribeDeploymentLogs(ctx context.Context, deploymentID string, filter string, limit int, onLog func(timestamp, message string, attributes map[string]string)) error {
	vars := map[string]any{"deploymentId": deploymentID, "filter": filter, "limit": limit}
	return iclient.Subscribe(ctx, c.cfg, igql.DeploymentLogsSub, vars,
		func(data json.RawMessage) {
			var pl igql.DeploymentLogsPayload
			if err := json.Unmarshal(data, &pl); err == nil {
				for _, l := range pl.DeploymentLogs {
					attrs := map[string]string{}
					for _, a := range l.Attributes {
						attrs[a.Key] = a.Value
					}
					if onLog != nil {
						onLog(l.Timestamp, l.Message, attrs)
					}
				}
			}
		}, func(err error) {},
	)
}

func (c *Client) SubscribeDeploymentStatus(ctx context.Context, deploymentID string, onStatus func(id, status string, stopped bool)) error {
	vars := map[string]any{"id": deploymentID}
	return iclient.Subscribe(ctx, c.cfg, igql.DeploymentStatusSub, vars,
		func(data json.RawMessage) {
			var st igql.DeploymentStatusPayload
			if err := json.Unmarshal(data, &st); err == nil {
				if onStatus != nil {
					onStatus(st.Deployment.ID, st.Deployment.Status, st.Deployment.DeploymentStopped)
				}
			}
		}, func(err error) {},
	)
}

// SubscribeEnvironmentLogs 订阅环境日志
func (c *Client) SubscribeEnvironmentLogs(ctx context.Context, environmentID string, filter string, beforeLimit int, beforeDate, anchorDate, afterDate string, afterLimit *int, onLog func(timestamp, message, severity string, tags map[string]*string, attributes map[string]string)) error {
	vars := map[string]interface{}{
		"environmentId": environmentID,
		"filter":        filter,
		"beforeLimit":   beforeLimit,
		"beforeDate":    beforeDate,
		"anchorDate":    anchorDate,
		"afterDate":     afterDate,
		"afterLimit":    afterLimit,
	}
	return iclient.Subscribe(ctx, c.cfg, igql.EnvironmentLogsSub, vars,
		func(data json.RawMessage) {
			var pl igql.EnvironmentLogsPayload
			if err := json.Unmarshal(data, &pl); err == nil {
				for _, l := range pl.EnvironmentLogs {
					// 转换tags为map
					tags := map[string]*string{
						"projectId":            l.Tags.ProjectID,
						"environmentId":        l.Tags.EnvironmentID,
						"pluginId":             l.Tags.PluginID,
						"serviceId":            l.Tags.ServiceID,
						"deploymentId":         l.Tags.DeploymentID,
						"deploymentInstanceId": l.Tags.DeploymentInstanceID,
						"snapshotId":           l.Tags.SnapshotID,
					}

					// 转换attributes为map
					attrs := map[string]string{}
					for _, a := range l.Attributes {
						attrs[a.Key] = a.Value
					}

					if onLog != nil {
						onLog(l.Timestamp, l.Message, l.Severity, tags, attrs)
					}
				}
			}
		}, func(err error) {},
	)
}
