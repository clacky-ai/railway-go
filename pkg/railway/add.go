package railway

import (
	"context"
	"fmt"
	"sort"
	"strings"

	igql "github.com/railwayapp/cli/internal/gql"
)

// AddOptions 用于高层 add API
type AddOptions struct {
	ProjectID     string
	EnvironmentID string

	// 数据库模板代码（如: postgres/mysql/redis/mongo）。若非空，则按模板部署数据库
	Databases []string

	// 服务创建参数（互斥优先级：Repo > Image > Empty）
	ServiceName string
	Repo        *string
	Image       *string

	// 初始环境变量（仅服务创建时使用）
	Variables map[string]string
}

// AddResult 封装 add 的结果
type AddResult struct {
	// 当创建服务时返回
	CreatedService *Service
	// 当通过模板添加数据库时返回一个或多个工作流
	TemplateResults []*TemplateDeployResult
}

// Add 高层 API：参考 CLI 行为
// - 若提供 Databases，则按模板部署数据库
// - 否则根据 Repo/Image/Empty 创建服务，并设置变量与分支
func (c *Client) Add(ctx context.Context, opts AddOptions) (*AddResult, error) {
	if strings.TrimSpace(opts.ProjectID) == "" || strings.TrimSpace(opts.EnvironmentID) == "" {
		return nil, fmt.Errorf("project/environment is required")
	}

	// 1) 数据库（经由模板部署）
	if len(opts.Databases) > 0 {
		results, err := c.addDatabasesViaTemplates(ctx, opts.ProjectID, opts.EnvironmentID, opts.Databases)
		if err != nil {
			return nil, err
		}
		return &AddResult{TemplateResults: results}, nil
	}

	// 2) 服务创建
	created, err := c.createServiceAdvanced(ctx, createServiceParams{
		projectID:     opts.ProjectID,
		environmentID: opts.EnvironmentID,
		serviceName:   strings.TrimSpace(opts.ServiceName),
		repo:          opts.Repo,
		image:         opts.Image,
		variables:     opts.Variables,
	})
	if err != nil {
		return nil, err
	}
	return &AddResult{CreatedService: created}, nil
}

// addDatabasesViaTemplates 通过模板部署数据库（参考 deploy.go 的 fetchAndCreate 流程）
func (c *Client) addDatabasesViaTemplates(ctx context.Context, projectID, environmentID string, databases []string) ([]*TemplateDeployResult, error) {
	if len(databases) == 0 {
		return nil, nil
	}
	// 规范化并排序
	normalized := make([]string, 0, len(databases))
	for _, db := range databases {
		code := strings.ToLower(strings.TrimSpace(db))
		if code == "postgresql" {
			code = "postgres"
		}
		if code == "mongodb" {
			code = "mongo"
		}
		switch code {
		case "postgres", "mysql", "redis", "mongo":
			normalized = append(normalized, code)
		default:
			// 与 CLI 一致：忽略未知并继续
			continue
		}
	}
	sort.Strings(normalized)
	if len(normalized) == 0 {
		return nil, fmt.Errorf("no supported database types provided")
	}

	results := make([]*TemplateDeployResult, 0, len(normalized))
	for _, code := range normalized {
		// 数据库模板不需要额外变量
		res, err := c.DeployTemplateWithConfig(ctx, TemplateDeployOptions{
			ProjectID:     projectID,
			EnvironmentID: environmentID,
			TemplateCode:  code,
			Variables:     map[string]string{},
		})
		if err != nil {
			return nil, fmt.Errorf("database create failed (%s): %w", code, err)
		}
		results = append(results, res)
	}
	return results, nil
}

// 内部参数结构，避免长参数列表
type createServiceParams struct {
	projectID     string
	environmentID string
	serviceName   string
	repo          *string
	image         *string
	variables     map[string]string
}

// createServiceAdvanced 创建服务，支持 repo/image/empty 与变量、自动解析默认分支
func (c *Client) createServiceAdvanced(ctx context.Context, p createServiceParams) (*Service, error) {
	var source *igql.Source
	// 优先级：Repo > Image > Empty
	if p.repo != nil && strings.TrimSpace(*p.repo) != "" {
		source = &igql.Source{Repo: p.repo}
	} else if p.image != nil && strings.TrimSpace(*p.image) != "" {
		source = &igql.Source{Image: p.image}
	}

	// 解析默认分支（仅当 repo 存在时）
	var branch *string
	if source != nil && source.Repo != nil && strings.TrimSpace(*source.Repo) != "" {
		if b := c.resolveDefaultBranch(ctx, *source.Repo); b != nil {
			branch = b
		}
	}

	// 构建变量输入
	vars := make(map[string]*string, len(p.variables))
	for k, v := range p.variables {
		vv := v
		vars[k] = &vv
	}

	// 发起 ServiceCreate 变更
	input := igql.ServiceCreateInput{
		ProjectID:     p.projectID,
		Name:          strings.TrimSpace(p.serviceName),
		Source:        source,
		EnvironmentID: p.environmentID,
		Variables:     vars,
		Branch:        branch,
	}

	var resp igql.ServiceCreateResponse
	if err := c.gqlClient.Mutate(ctx, igql.ServiceCreateMutation, map[string]any{"input": input}, &resp); err != nil {
		// 回退：不带 variables/environmentId/branch
		fallback := igql.ServiceCreateInput{ProjectID: input.ProjectID, Name: input.Name, Source: input.Source}
		if e := c.gqlClient.Mutate(ctx, igql.ServiceCreateMutation, map[string]any{"input": fallback}, &resp); e != nil {
			return nil, fmt.Errorf("service create failed: %w", err)
		}
		// 若有变量，则单独 upsert
		if len(vars) > 0 {
			up := igql.VariableCollectionUpsertInput{ProjectID: p.projectID, EnvironmentID: p.environmentID, ServiceID: &resp.ServiceCreate.ID, Variables: vars}
			var upResp igql.VariableCollectionUpsertResponse
			if e := c.gqlClient.Mutate(ctx, igql.VariableCollectionUpsertMutation, map[string]any{"input": up}, &upResp); e != nil {
				return nil, fmt.Errorf("set initial variables failed: %w", e)
			}
		}
	}

	return &Service{ID: resp.ServiceCreate.ID, Name: resp.ServiceCreate.Name}, nil
}

// resolveDefaultBranch 查询 repo 的默认分支
func (c *Client) resolveDefaultBranch(ctx context.Context, fullRepo string) *string {
	var repos igql.GitHubReposResponse
	if err := c.gqlClient.Query(ctx, igql.GitHubReposQuery, nil, &repos); err != nil {
		return nil
	}
	for _, r := range repos.GitHubRepos {
		if r.FullName == fullRepo {
			b := r.DefaultBranch
			return &b
		}
	}
	return nil
}
