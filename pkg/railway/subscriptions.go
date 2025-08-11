package railway

import (
	"context"
	"encoding/json"

	iclient "github.com/railwayapp/cli/internal/client"
	igql "github.com/railwayapp/cli/internal/gql"
)

// 订阅封装
func (c *Client) SubscribeBuildLogs(ctx context.Context, deploymentID string, filter string, limit int, onLog func(timestamp, message string)) error {
	vars := map[string]any{"deploymentId": deploymentID, "filter": filter, "limit": limit}
	return iclient.Subscribe(ctx, c.cfg, igql.BuildLogsSub, vars,
		func(data json.RawMessage) {
			var pl igql.BuildLogsPayload
			if err := json.Unmarshal(data, &pl); err == nil {
				for _, l := range pl.BuildLogs {
					if onLog != nil {
						onLog(l.Timestamp, l.Message)
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
