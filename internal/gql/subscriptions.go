package gql

// BuildLogs subscription
const BuildLogsSub = `
subscription BuildLogs($deploymentId: String!, $filter: String, $limit: Int) {
  buildLogs(deploymentId: $deploymentId, filter: $filter, limit: $limit) {
    timestamp
    message
attributes {
    key
    value
  }
  }
}
`

type BuildLogsPayload struct {
	BuildLogs []struct {
		Timestamp  string `json:"timestamp"`
		Message    string `json:"message"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"attributes"`
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

// EnvironmentLogs subscription
const EnvironmentLogsSub = `
subscription streamEnvironmentLogs($environmentId: String!, $filter: String, $beforeLimit: Int!, $beforeDate: String, $anchorDate: String, $afterDate: String, $afterLimit: Int) {
  environmentLogs(
    environmentId: $environmentId
    filter: $filter
    beforeDate: $beforeDate
    anchorDate: $anchorDate
    afterDate: $afterDate
    beforeLimit: $beforeLimit
    afterLimit: $afterLimit
  ) {
    ...LogFields
  }
}

fragment LogFields on Log {
  timestamp
  message
  severity
  tags {
    projectId
    environmentId
    pluginId
    serviceId
    deploymentId
    deploymentInstanceId
    snapshotId
  }
  attributes {
    key
    value
  }
}
`

type EnvironmentLogsPayload struct {
	EnvironmentLogs []struct {
		Timestamp string `json:"timestamp"`
		Message   string `json:"message"`
		Severity  string `json:"severity"`
		Tags      struct {
			ProjectID            *string `json:"projectId"`
			EnvironmentID        *string `json:"environmentId"`
			PluginID             *string `json:"pluginId"`
			ServiceID            *string `json:"serviceId"`
			DeploymentID         *string `json:"deploymentId"`
			DeploymentInstanceID *string `json:"deploymentInstanceId"`
			SnapshotID           *string `json:"snapshotId"`
		} `json:"tags"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"attributes"`
	} `json:"environmentLogs"`
}
