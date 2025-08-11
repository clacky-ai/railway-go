package railway

import (
	"context"
	"fmt"

	igql "github.com/railwayapp/cli/internal/gql"
)

type ServiceDomain struct {
	ID     string
	Domain string
}

type CustomDomain struct {
	ID     string
	Domain string
}

type Domains struct {
	ServiceDomains []ServiceDomain
	CustomDomains  []CustomDomain
}

// ListDomains 列出服务的域名（服务域名与自定义域名）
func (c *Client) ListDomains(ctx context.Context, projectID, environmentID, serviceID string) (*Domains, error) {
	var resp struct {
		Domains struct {
			ServiceDomains []struct {
				ID     string `json:"id"`
				Domain string `json:"domain"`
			} `json:"serviceDomains"`
			CustomDomains []struct {
				ID     string `json:"id"`
				Domain string `json:"domain"`
			} `json:"customDomains"`
		} `json:"domains"`
	}
	if err := c.gqlClient.Query(ctx, igql.DomainsQuery, map[string]any{
		"projectId":     projectID,
		"environmentId": environmentID,
		"serviceId":     serviceID,
	}, &resp); err != nil {
		return nil, err
	}
	out := &Domains{}
	for _, d := range resp.Domains.ServiceDomains {
		out.ServiceDomains = append(out.ServiceDomains, ServiceDomain{ID: d.ID, Domain: d.Domain})
	}
	for _, d := range resp.Domains.CustomDomains {
		out.CustomDomains = append(out.CustomDomains, CustomDomain{ID: d.ID, Domain: d.Domain})
	}
	return out, nil
}

// CreateServiceDomain 为服务创建默认域名
func (c *Client) CreateServiceDomain(ctx context.Context, environmentID, serviceID string) (ServiceDomain, error) {
	var resp struct {
		ServiceDomainCreate struct {
			ID     string `json:"id"`
			Domain string `json:"domain"`
		} `json:"serviceDomainCreate"`
	}
	if err := c.gqlClient.Mutate(ctx, igql.ServiceDomainCreateMutation, map[string]any{
		"environmentId": environmentID,
		"serviceId":     serviceID,
	}, &resp); err != nil {
		return ServiceDomain{}, err
	}
	return ServiceDomain{ID: resp.ServiceDomainCreate.ID, Domain: resp.ServiceDomainCreate.Domain}, nil
}

// CheckCustomDomainAvailable 检查自定义域名可用性
func (c *Client) CheckCustomDomainAvailable(ctx context.Context, domain string) (available bool, message string, err error) {
	var resp struct {
		CustomDomainAvailable struct {
			Available bool   `json:"available"`
			Message   string `json:"message"`
		} `json:"customDomainAvailable"`
	}
	if err := c.gqlClient.Query(ctx, igql.CustomDomainAvailableQuery, map[string]any{"domain": domain}, &resp); err != nil {
		return false, "", err
	}
	return resp.CustomDomainAvailable.Available, resp.CustomDomainAvailable.Message, nil
}

// CreateCustomDomain 绑定自定义域名，可选映射到指定端口
func (c *Client) CreateCustomDomain(ctx context.Context, projectID, environmentID, serviceID, domain string, targetPort *int) (CustomDomain, error) {
	input := map[string]any{
		"domain":        domain,
		"projectId":     projectID,
		"environmentId": environmentID,
		"serviceId":     serviceID,
	}
	if targetPort != nil && *targetPort > 0 {
		input["targetPort"] = *targetPort
	}
	var resp struct {
		CustomDomainCreate struct {
			ID     string `json:"id"`
			Domain string `json:"domain"`
		} `json:"customDomainCreate"`
	}
	if err := c.gqlClient.Mutate(ctx, igql.CustomDomainCreateMutation, map[string]any{"input": input}, &resp); err != nil {
		return CustomDomain{}, err
	}
	return CustomDomain{ID: resp.CustomDomainCreate.ID, Domain: resp.CustomDomainCreate.Domain}, nil
}

// DeleteDomain 删除服务域名或自定义域名（传入 ID，内部尝试两种删除）
func (c *Client) DeleteDomain(ctx context.Context, id string) error {
	var raw1 struct {
		ServiceDomainDelete bool `json:"serviceDomainDelete"`
	}
	if err := c.gqlClient.Mutate(ctx, igql.ServiceDomainDeleteMutation, map[string]any{"id": id}, &raw1); err == nil {
		if raw1.ServiceDomainDelete {
			return nil
		}
	}
	var raw2 struct {
		CustomDomainDelete bool `json:"customDomainDelete"`
	}
	if err := c.gqlClient.Mutate(ctx, igql.CustomDomainDeleteMutation, map[string]any{"id": id}, &raw2); err == nil {
		if raw2.CustomDomainDelete {
			return nil
		}
	}
	return fmt.Errorf("failed to delete domain: %s", id)
}
