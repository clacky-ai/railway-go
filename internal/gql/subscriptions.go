package gql

// BuildLogs subscription
const BuildLogsSub = `
subscription BuildLogs($deploymentId: String!, $filter: String, $limit: Int) {
  buildLogs(deploymentId: $deploymentId, filter: $filter, limit: $limit) {
    timestamp
    message
  }
}
`

type BuildLogsPayload struct {
	BuildLogs []struct {
		Timestamp string `json:"timestamp"`
		Message   string `json:"message"`
	} `json:"buildLogs"`
}

// DeploymentLogs subscription
const DeploymentLogsSub = `
subscription DeploymentLogs($deploymentId: String!, $filter: String, $limit: Int) {
  deploymentLogs(deploymentId: $deploymentId, filter: $filter, limit: $limit) {
    timestamp
    message
    attributes { key value }
  }
}
`

type DeploymentLogsPayload struct {
	DeploymentLogs []struct {
		Timestamp  string `json:"timestamp"`
		Message    string `json:"message"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"attributes"`
	} `json:"deploymentLogs"`
}

// Deployment status subscription
const DeploymentStatusSub = `
subscription Deployment($id: String!) {
  deployment(id: $id) {
    id
    status
    deploymentStopped
  }
}
`

type DeploymentStatusPayload struct {
	Deployment struct {
		ID                string `json:"id"`
		Status            string `json:"status"`
		DeploymentStopped bool   `json:"deploymentStopped"`
	} `json:"deployment"`
}
