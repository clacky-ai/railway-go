package gql

import (
	"encoding/json"
)

// UserMeta GraphQL查询
const UserMetaQuery = `
query UserMeta {
  me {
    id
    name
    email
    avatar
  }
}
`

// UserMetaResponse 用户元数据响应
type UserMetaResponse struct {
	Me struct {
		ID     string  `json:"id"`
		Name   *string `json:"name"`
		Email  string  `json:"email"`
		Avatar *string `json:"avatar"`
	} `json:"me"`
}

// GitHubRepos 查询（用于解析默认分支）
const GitHubReposQuery = `
query GitHubRepos {
  githubRepos {
    fullName
    defaultBranch
  }
}
`

// GitHubReposResponse 响应
type GitHubReposResponse struct {
	GitHubRepos []struct {
		FullName      string `json:"fullName"`
		DefaultBranch string `json:"defaultBranch"`
	} `json:"githubRepos"`
}

// TemplateDetail 查询（用于模板/数据库部署）
const TemplateDetailQuery = `
query TemplateDetail($code: String!) {
  template(code: $code) {
    id
    name
    serializedConfig
  }
}
`

// TemplateDetailResponse 响应（serializedConfig 兼容字符串或对象）
type TemplateDetailResponse struct {
	Template struct {
		ID               string          `json:"id"`
		Name             string          `json:"name"`
		SerializedConfig json.RawMessage `json:"serializedConfig"`
	} `json:"template"`
}

// Projects GraphQL查询
const ProjectsQuery = `
query Projects {
  projects {
    edges {
      node {
        id
        name
        description
        createdAt
        updatedAt
      }
    }
  }
}
`

// ProjectsResponse 项目列表响应
type ProjectsResponse struct {
	Projects struct {
		Edges []struct {
			Node struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				CreatedAt   string  `json:"createdAt"`
				UpdatedAt   string  `json:"updatedAt"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"projects"`
}

// Project GraphQL查询
const ProjectQuery = `
query Project($id: String!) {
  project(id: $id) {
    id
    name
    description
    environments {
      edges {
        node {
          id
          name
        }
      }
    }
    services {
      edges {
        node {
          id
          name
          icon
          serviceInstances {
            edges {
              node {
                id
                serviceName
                environmentId
              }
            }
          }
        }
      }
    }
  }
}
`

// ProjectResponse 项目详情响应
type ProjectResponse struct {
	Project struct {
		ID           string  `json:"id"`
		Name         string  `json:"name"`
		Description  *string `json:"description"`
		Environments struct {
			Edges []struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"environments"`
		Services struct {
			Edges []struct {
				Node struct {
					ID               string  `json:"id"`
					Name             string  `json:"name"`
					Icon             *string `json:"icon"`
					ServiceInstances struct {
						Edges []struct {
							Node struct {
								ID            string `json:"id"`
								ServiceName   string `json:"serviceName"`
								EnvironmentID string `json:"environmentId"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"serviceInstances"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"services"`
	} `json:"project"`
}

// ProjectToken GraphQL查询
const ProjectTokenQuery = `
query ProjectToken {
  projectToken {
    project {
      id
      name
    }
    environment {
      id
      name
    }
  }
}
`

// ProjectTokenResponse 项目令牌响应
type ProjectTokenResponse struct {
	ProjectToken struct {
		Project struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		Environment struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"environment"`
	} `json:"projectToken"`
}

// Deployments GraphQL查询
const DeploymentsQuery = `
query Deployments($projectId: String!, $environmentId: String!, $serviceId: String) {
  deployments(
    input: {
      projectId: $projectId
      environmentId: $environmentId
      serviceId: $serviceId
    }
    first: 20
  ) {
    edges {
      node {
        id
        status
        createdAt
        updatedAt
        staticUrl
        url
        service {
          id
          name
        }
      }
    }
  }
}
`

// DeploymentsResponse 部署列表响应
type DeploymentsResponse struct {
	Deployments struct {
		Edges []struct {
			Node struct {
				ID        string  `json:"id"`
				Status    string  `json:"status"`
				CreatedAt string  `json:"createdAt"`
				UpdatedAt string  `json:"updatedAt"`
				StaticURL *string `json:"staticUrl"`
				URL       *string `json:"url"`
				Service   struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"service"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"deployments"`
}

// VariablesForServiceDeployment GraphQL查询
const VariablesForServiceDeploymentQuery = `
query VariablesForServiceDeployment($projectId: String!, $environmentId: String!, $serviceId: String!) {
  variables(
    projectId: $projectId
    environmentId: $environmentId
    serviceId: $serviceId
  )
}
`

// VariablesResponse 环境变量响应
type VariablesResponse struct {
	Variables map[string]*string `json:"variables"`
}

// Domains 查询
const DomainsQuery = `
query Domains($environmentId: String!, $projectId: String!, $serviceId: String!) {
  domains(environmentId: $environmentId, projectId: $projectId, serviceId: $serviceId) {
    serviceDomains { id domain }
    customDomains { 
      id 
      domain 
      status {
        dnsRecords {
          hostlabel
          fqdn
          recordType
          requiredValue
          currentValue
          status
          zone
          purpose
        }
      }
    }
  }
}
`

// UserProjects（Workspaces）查询
const UserProjectsQuery = `
query UserProjects {
  externalWorkspaces {
    id
    name
    teamId
  }
  me {
    workspaces {
      id
      name
      team { id }
    }
  }
}
`

// UserProjectsResponse 工作区响应
type UserProjectsResponse struct {
	ExternalWorkspaces []struct {
		ID     string  `json:"id"`
		Name   string  `json:"name"`
		TeamID *string `json:"teamId"`
	} `json:"externalWorkspaces"`
	Me struct {
		Workspaces []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Team *struct {
				ID string `json:"id"`
			} `json:"team"`
		} `json:"workspaces"`
	} `json:"me"`
}

// 完整的工作区+项目详情查询（含环境与服务）
const UserProjectsFullQuery = `
query UserProjects {
  externalWorkspaces {
    id
    name
    teamId
    projects {
      id
      name
      createdAt
      updatedAt
      deletedAt
      team { id name }
      environments { edges { node { id name } } }
      services {
        edges { node {
          id
          name
          serviceInstances { edges { node { environmentId } } }
        } }
      }
    }
  }
  me {
    workspaces {
      id
      name
      team {
        id
        projects {
          edges { node {
            id
            name
            createdAt
            updatedAt
            deletedAt
            team { id name }
            environments { edges { node { id name } } }
            services {
              edges { node {
                id
                name
                serviceInstances { edges { node { environmentId } } }
              } }
            }
          } }
        }
      }
    }
  }
}
`

type UserProjectsFullResponse struct {
	ExternalWorkspaces []struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		TeamID   *string `json:"teamId"`
		Projects []struct {
			ID           string  `json:"id"`
			Name         string  `json:"name"`
			DeletedAt    *string `json:"deletedAt"`
			Environments struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"environments"`
			Services struct {
				Edges []struct {
					Node struct {
						ID               string `json:"id"`
						Name             string `json:"name"`
						ServiceInstances struct {
							Edges []struct {
								Node struct {
									EnvironmentID string `json:"environmentId"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"serviceInstances"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"services"`
		} `json:"projects"`
	} `json:"externalWorkspaces"`
	Me struct {
		Workspaces []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Team *struct {
				ID       string `json:"id"`
				Projects struct {
					Edges []struct {
						Node struct {
							ID           string  `json:"id"`
							Name         string  `json:"name"`
							DeletedAt    *string `json:"deletedAt"`
							Environments struct {
								Edges []struct {
									Node struct {
										ID   string `json:"id"`
										Name string `json:"name"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"environments"`
							Services struct {
								Edges []struct {
									Node struct {
										ID               string `json:"id"`
										Name             string `json:"name"`
										ServiceInstances struct {
											Edges []struct {
												Node struct {
													EnvironmentID string `json:"environmentId"`
												} `json:"node"`
											} `json:"edges"`
										} `json:"serviceInstances"`
									} `json:"node"`
								} `json:"edges"`
							} `json:"services"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"projects"`
			} `json:"team"`
		} `json:"workspaces"`
	} `json:"me"`
}

// ProjectTokens GraphQL查询（分页列表）
const ProjectTokensQuery = `
query ProjectTokens($projectId: String!, $after: String) {
  projectTokens(projectId: $projectId, first: 50, after: $after) {
    edges {
      cursor
      node {
        id
        name
        project {
          id
          name
        }
        environment {
          id
          name
        }
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`

// ProjectTokensResponse 项目令牌列表响应
type ProjectTokensResponse struct {
	ProjectTokens struct {
		Edges []struct {
			Cursor string `json:"cursor"`
			Node   struct {
				ID          string                    `json:"id"`
				Name        string                    `json:"name"`
				Project     struct{ ID, Name string } `json:"project"`
				Environment struct{ ID, Name string } `json:"environment"`
			} `json:"node"`
		} `json:"edges"`
		PageInfo struct {
			HasNextPage bool    `json:"hasNextPage"`
			EndCursor   *string `json:"endCursor"`
		} `json:"pageInfo"`
	} `json:"projectTokens"`
}
