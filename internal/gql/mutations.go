package gql

// LoginSessionCreate GraphQL变更
const LoginSessionCreateMutation = `
mutation LoginSessionCreate {
  loginSessionCreate
}
`

// LoginSessionCreateResponse 登录会话创建响应
type LoginSessionCreateResponse struct {
	LoginSessionCreate string `json:"loginSessionCreate"`
}

// LoginSessionConsume GraphQL变更
const LoginSessionConsumeMutation = `
mutation LoginSessionConsume($code: String!) {
  loginSessionConsume(code: $code)
}
`

// LoginSessionConsumeResponse 登录会话消费响应
type LoginSessionConsumeResponse struct {
	LoginSessionConsume *string `json:"loginSessionConsume"`
}

// ProjectCreate GraphQL变更（返回环境列表，与Rust版一致）
const ProjectCreateMutation = `
mutation ProjectCreate($name: String, $description: String, $teamId: String) {
  projectCreate(input: { name: $name, description: $description, teamId: $teamId }) {
    name
    id
    environments {
      edges {
        node {
          id
          name
        }
      }
    }
  }
}
`

// ProjectCreateResponse 项目创建响应
type ProjectCreateResponse struct {
	ProjectCreate struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Environments struct {
			Edges []struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"environments"`
	} `json:"projectCreate"`
}

// EnvironmentCreate GraphQL变更
const EnvironmentCreateMutation = `
mutation EnvironmentCreate($input: EnvironmentCreateInput!) {
  environmentCreate(input: $input) {
    id
    name
  }
}
`

// EnvironmentCreateInput 环境创建输入
type EnvironmentCreateInput struct {
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
}

// EnvironmentCreateResponse 环境创建响应
type EnvironmentCreateResponse struct {
	EnvironmentCreate struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"environmentCreate"`
}

// ServiceCreate GraphQL变更
const ServiceCreateMutation = `
mutation ServiceCreate($input: ServiceCreateInput!) {
  serviceCreate(input: $input) {
    id
    name
  }
}
`

// ServiceCreateInput 服务创建输入
type ServiceCreateInput struct {
	ProjectID     string             `json:"projectId"`
	Name          string             `json:"name"`
	Source        *Source            `json:"source,omitempty"`
	EnvironmentID string             `json:"environmentId,omitempty"`
	Variables     map[string]*string `json:"variables,omitempty"`
	Branch        *string            `json:"branch,omitempty"`
}

// Source 服务源配置
type Source struct {
	Repo  *string `json:"repo,omitempty"`
	Image *string `json:"image,omitempty"`
}

// ServiceCreateResponse 服务创建响应
type ServiceCreateResponse struct {
	ServiceCreate struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"serviceCreate"`
}

// ServiceInstanceDeploy GraphQL变更
const ServiceInstanceDeployMutation = `
mutation ServiceInstanceDeploy($input: ServiceInstanceDeployInput!) {
  serviceInstanceDeploy(input: $input) {
    id
    status
  }
}
`

// ServiceInstanceDeployInput 服务实例部署输入
type ServiceInstanceDeployInput struct {
	ServiceID     string `json:"serviceId"`
	EnvironmentID string `json:"environmentId"`
}

// ServiceInstanceDeployResponse 服务实例部署响应
type ServiceInstanceDeployResponse struct {
	ServiceInstanceDeploy struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"serviceInstanceDeploy"`
}

// VariableCollectionUpsert GraphQL变更
const VariableCollectionUpsertMutation = `
mutation VariableCollectionUpsert($input: VariableCollectionUpsertInput!) {
  variableCollectionUpsert(input: $input)
}
`

// VariableCollectionUpsertInput 变量集合更新输入
type VariableCollectionUpsertInput struct {
	ProjectID     string             `json:"projectId"`
	EnvironmentID string             `json:"environmentId"`
	ServiceID     *string            `json:"serviceId,omitempty"`
	Replace       *bool              `json:"replace,omitempty"`
	Variables     map[string]*string `json:"variables"`
}

// VariableCollectionUpsertResponse 变量集合更新响应
type VariableCollectionUpsertResponse struct {
	VariableCollectionUpsert bool `json:"variableCollectionUpsert"`
}

// TemplateDeploy GraphQL变更
const TemplateDeployMutation = `
mutation TemplateDeploy($projectId: String!, $environmentId: String!, $templateId: String!, $serializedConfig: SerializedTemplateConfig!) {
  templateDeployV2(input: { projectId: $projectId, environmentId: $environmentId, templateId: $templateId, serializedConfig: $serializedConfig }) {
    projectId
    workflowId
  }
}
`

// TemplateDeployInput 模板部署输入
type TemplateDeployInput struct {
	ProjectID        string                   `json:"projectId"`
	EnvironmentID    string                   `json:"environmentId"`
	TemplateID       string                   `json:"templateId"`
	SerializedConfig SerializedTemplateConfig `json:"serializedConfig"`
}

// SerializedTemplateConfig 序列化模板配置
type SerializedTemplateConfig map[string]interface{}

// TemplateDeployResponse 模板部署响应
type TemplateDeployResponse struct {
	TemplateDeployV2 struct {
		ProjectID  string `json:"projectId"`
		WorkflowID string `json:"workflowId"`
	} `json:"templateDeployV2"`
}

// DeploymentRedeploy GraphQL变更
const DeploymentRedeployMutation = `
mutation DeploymentRedeploy($id: String!) {
  deploymentRedeploy(id: $id) {
    id
    status
  }
}
`

// DeploymentRedeployResponse 部署重新部署响应
type DeploymentRedeployResponse struct {
	DeploymentRedeploy struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"deploymentRedeploy"`
}

// ServiceDelete GraphQL变更
const ServiceDeleteMutation = `
mutation ServiceDelete($id: String!) {
  serviceDelete(id: $id)
}
`

// ServiceDeleteResponse 服务删除响应
type ServiceDeleteResponse struct {
	ServiceDelete bool `json:"serviceDelete"`
}

// ProjectDelete GraphQL变更
const ProjectDeleteMutation = `
mutation ProjectDelete($id: String!) {
  projectDelete(id: $id)
}
`

// ProjectDeleteResponse 项目删除响应
type ProjectDeleteResponse struct {
	ProjectDelete bool `json:"projectDelete"`
}

// DeploymentRemove 删除部署
const DeploymentRemoveMutation = `
mutation DeploymentRemove($id: String!) {
  deploymentRemove(id: $id)
}
`

// ServiceDomainCreate 创建服务域名
const ServiceDomainCreateMutation = `
mutation ServiceDomainCreate($environmentId: String!, $serviceId: String!) {
  serviceDomainCreate(input: { environmentId: $environmentId, serviceId: $serviceId }) {
    id
    domain
  }
}
`

// CustomDomainCreate 创建自定义域名
const CustomDomainCreateMutation = `
mutation CustomDomainCreate($input: CustomDomainCreateInput!) {
  customDomainCreate(input: $input) {
    id
    domain
    status { dnsRecords { hostlabel fqdn recordType requiredValue currentValue status zone purpose } }
  }
}
`

// CustomDomainAvailable 查询
const CustomDomainAvailableQuery = `
query CustomDomainAvailable($domain: String!) { customDomainAvailable(domain: $domain) { available message } }
`

// CustomDomainDelete 删除自定义域名
const CustomDomainDeleteMutation = `
mutation CustomDomainDelete($id: String!) { customDomainDelete(id: $id) }
`

// ServiceDomainDelete 删除服务域名
const ServiceDomainDeleteMutation = `
mutation ServiceDomainDelete($id: String!) { serviceDomainDelete(id: $id) }
`

// ServiceInstanceStop GraphQL变更（input 形式）
const ServiceInstanceStopMutation = `
mutation ServiceInstanceStop($input: ServiceInstanceStopInput!) {
  serviceInstanceStop(input: $input)
}
`

// ServiceInstanceStopInput 服务实例停止输入
type ServiceInstanceStopInput struct {
	ServiceID     string `json:"serviceId"`
	EnvironmentID string `json:"environmentId"`
}

// ServiceInstanceStopResponse 服务实例停止响应
type ServiceInstanceStopResponse struct {
	ServiceInstanceStop bool `json:"serviceInstanceStop"`
}

// ServiceInstanceScale GraphQL变更（input 形式）
const ServiceInstanceScaleMutation = `
mutation ServiceInstanceScale($input: ServiceInstanceScaleInput!) {
  serviceInstanceScale(input: $input)
}
`

// ServiceInstanceScaleInput 服务实例缩放输入
type ServiceInstanceScaleInput struct {
	ServiceID     string `json:"serviceId"`
	EnvironmentID string `json:"environmentId"`
	Replicas      int    `json:"replicas"`
}

// ServiceInstanceScaleResponse 服务实例缩放响应
type ServiceInstanceScaleResponse struct {
	ServiceInstanceScale bool `json:"serviceInstanceScale"`
}

// ServiceInstanceStopByParams GraphQL变更（参数形式）
const ServiceInstanceStopByParamsMutation = `
mutation ServiceInstanceStopByParams($serviceId: String!, $environmentId: String!) {
  serviceInstanceStop(serviceId: $serviceId, environmentId: $environmentId)
}
`

// ServiceInstanceScaleByParams GraphQL变更（参数形式）
const ServiceInstanceScaleByParamsMutation = `
mutation ServiceInstanceScaleByParams($serviceId: String!, $environmentId: String!) {
  serviceInstanceScale(serviceId: $serviceId, environmentId: $environmentId, replicas: 0)
}
`

// DeploymentStop GraphQL变更（返回对象）
const DeploymentStopMutation = `
mutation DeploymentStop($id: String!) {
  deploymentStop(id: $id) {
    id
    status
    deploymentStopped
  }
}
`

// DeploymentStopResponse 部署停止响应（对象形式）
type DeploymentStopResponse struct {
	DeploymentStop struct {
		ID                string `json:"id"`
		Status            string `json:"status"`
		DeploymentStopped bool   `json:"deploymentStopped"`
	} `json:"deploymentStop"`
}

// DeploymentStopSimple GraphQL变更（返回布尔值）
const DeploymentStopSimpleMutation = `
mutation DeploymentStopSimple($id: String!) {
  deploymentStop(id: $id)
}
`

// DeploymentStopSimpleResponse 部署停止响应（布尔形式）
type DeploymentStopSimpleResponse struct {
	DeploymentStop bool `json:"deploymentStop"`
}

// DeploymentCancel GraphQL变更
const DeploymentCancelMutation = `
mutation DeploymentCancel($id: String!) {
  deploymentCancel(id: $id)
}
`

// DeploymentCancelResponse 部署取消响应
type DeploymentCancelResponse struct {
	DeploymentCancel bool `json:"deploymentCancel"`
}

// DeploymentAbort GraphQL变更
const DeploymentAbortMutation = `
mutation DeploymentAbort($id: String!) {
  deploymentAbort(id: $id)
}
`

// DeploymentAbortResponse 部署中止响应
type DeploymentAbortResponse struct {
	DeploymentAbort bool `json:"deploymentAbort"`
}

// ProjectTokenCreate GraphQL变更（input 形式）
const ProjectTokenCreateMutation = `
mutation ProjectTokenCreate($input: ProjectTokenCreateInput!) {
  projectTokenCreate(input: $input)
}
`

// ProjectTokenCreateInput 项目令牌创建输入
type ProjectTokenCreateInput struct {
	Name          string `json:"name"`
	ProjectID     string `json:"projectId"`
	EnvironmentID string `json:"environmentId"`
}

// ProjectTokenCreateResponse 项目令牌创建响应
type ProjectTokenCreateResponse struct {
	ProjectTokenCreate string `json:"projectTokenCreate"`
}

// ProjectTokenCreateByParams GraphQL变更（参数形式）
const ProjectTokenCreateByParamsMutation = `
mutation ProjectTokenCreateByParams($projectId: String!, $environmentId: String!, $name: String!) {
  projectTokenCreate(projectId: $projectId, environmentId: $environmentId, name: $name)
}
`

// ProjectTokenDelete GraphQL变更（直接参数）
const ProjectTokenDeleteMutation = `
mutation ProjectTokenDelete($id: String!) {
  projectTokenDelete(id: $id)
}
`

// ProjectTokenDeleteResponse 项目令牌删除响应
type ProjectTokenDeleteResponse struct {
	ProjectTokenDelete bool `json:"projectTokenDelete"`
}

// ProjectTokenDeleteByInput GraphQL变更（input 形式）
const ProjectTokenDeleteByInputMutation = `
mutation ProjectTokenDeleteByInput($input: ProjectTokenDeleteInput!) {
  projectTokenDelete(input: $input)
}
`

// ProjectTokenDeleteInput 项目令牌删除输入
type ProjectTokenDeleteInput struct {
	ID string `json:"id"`
}

// VolumeInstanceBackupCreate GraphQL变更（直接参数）
const VolumeInstanceBackupCreateMutation = `
mutation VolumeInstanceBackupCreate($volumeInstanceId: String!) {
  volumeInstanceBackupCreate(volumeInstanceId: $volumeInstanceId) {
    workflowId
  }
}
`

// VolumeInstanceBackupCreateResponse 备份创建响应（返回工作流ID）
type VolumeInstanceBackupCreateResponse struct {
	VolumeInstanceBackupCreate struct {
		WorkflowID string `json:"workflowId"`
	} `json:"volumeInstanceBackupCreate"`
}

// VolumeInstanceBackupRestore GraphQL变更
const VolumeInstanceBackupRestoreMutation = `
mutation VolumeInstanceBackupRestore($volumeInstanceId: String!, $volumeInstanceBackupId: String!) {
  volumeInstanceBackupRestore(
    volumeInstanceId: $volumeInstanceId
    volumeInstanceBackupId: $volumeInstanceBackupId
  ) {
    workflowId
  }
}
`

// VolumeInstanceBackupRestoreResponse 备份恢复响应
type VolumeInstanceBackupRestoreResponse struct {
	VolumeInstanceBackupRestore struct {
		WorkflowID string `json:"workflowId"`
	} `json:"volumeInstanceBackupRestore"`
}

// VolumeInstanceBackupBatchDelete GraphQL变更
const VolumeInstanceBackupBatchDeleteMutation = `
mutation volumeInstanceBackupBatchDelete($volumeInstanceId: String!, $volumeInstanceBackupIds: [String!]!) {
  volumeInstanceBackupBatchDelete(
    volumeInstanceId: $volumeInstanceId
    volumeInstanceBackupIds: $volumeInstanceBackupIds
  ) {
    workflowId
  }
}
`

// VolumeInstanceBackupBatchDeleteResponse 备份批量删除响应
type VolumeInstanceBackupBatchDeleteResponse struct {
	VolumeInstanceBackupBatchDelete struct {
		WorkflowID string `json:"workflowId"`
	} `json:"volumeInstanceBackupBatchDelete"`
}

// EnvironmentPatchCommitStaged GraphQL变更
const EnvironmentPatchCommitStagedMutation = `
mutation environmentPatchCommitStaged($environmentId: String!, $message: String, $skipDeploys: Boolean) {
  environmentPatchCommitStaged(
    environmentId: $environmentId
    commitMessage: $message
    skipDeploys: $skipDeploys
  )
}
`

// EnvironmentPatchCommitStagedResponse 环境补丁提交阶段性变更响应
type EnvironmentPatchCommitStagedResponse struct {
	EnvironmentPatchCommitStaged string `json:"environmentPatchCommitStaged"`
}

// DeploymentRollback GraphQL变更
const DeploymentRollbackMutation = `
mutation deploymentRollback($id: String!) {
  deploymentRollback(id: $id)
}
`

// DeploymentRollbackResponse 部署回滚响应
type DeploymentRollbackResponse struct {
	DeploymentRollback bool `json:"deploymentRollback"`
}

// VolumeInstanceBackupScheduleUpdate GraphQL变更
const VolumeInstanceBackupScheduleUpdateMutation = `
mutation volumeInstanceBackupScheduleUpdate($volumeInstanceId: String!, $kinds: [VolumeInstanceBackupScheduleKind!]!) {
  volumeInstanceBackupScheduleUpdate(
    volumeInstanceId: $volumeInstanceId
    kinds: $kinds
  )
}
`

// VolumeInstanceBackupScheduleUpdateResponse 卷备份调度更新响应
type VolumeInstanceBackupScheduleUpdateResponse struct {
	VolumeInstanceBackupScheduleUpdate bool `json:"volumeInstanceBackupScheduleUpdate"`
}

// EnvironmentStageChanges GraphQL变更
const EnvironmentStageChangesMutation = `
mutation stageEnvironmentChanges($environmentId: String!, $payload: EnvironmentConfig!) {
  environmentStageChanges(environmentId: $environmentId, input: $payload) {
    id
  }
}
`

// EnvironmentStageChangesResponse 环境变更暂存响应
type EnvironmentStageChangesResponse struct {
	EnvironmentStageChanges struct {
		ID string `json:"id"`
	} `json:"environmentStageChanges"`
}
