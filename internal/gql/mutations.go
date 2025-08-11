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
mutation TemplateDeploy($input: TemplateDeployInput!) {
  templateDeploy(input: $input) {
    id
    name
  }
}
`

// TemplateDeployInput 模板部署输入
type TemplateDeployInput struct {
	ProjectID        string      `json:"projectId"`
	EnvironmentID    string      `json:"environmentId"`
	TemplateID       string      `json:"templateId"`
	SerializedConfig interface{} `json:"serializedConfig"`
}

// TemplateDeployResponse 模板部署响应
type TemplateDeployResponse struct {
	TemplateDeploy struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"templateDeploy"`
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
