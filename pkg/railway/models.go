package railway

// 更丰富的数据模型（基于现有 GraphQL 响应映射）

// ServiceInfo 服务详情
type ServiceInfo struct {
	ID                     string
	Name                   string
	Icon                   *string
	InstanceEnvironmentIDs []string
}

// ProjectInfo 项目详情
type ProjectInfo struct {
	ID           string
	Name         string
	Description  *string
	Environments []Environment
	Services     []ServiceInfo
}

// DeploymentInfo 部署详情
type DeploymentInfo struct {
	ID        string
	Status    string
	CreatedAt string
	UpdatedAt string
	StaticURL *string
	URL       *string
	Service   Service
}
