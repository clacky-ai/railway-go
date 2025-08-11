package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/machinebox/graphql"
	"github.com/railwayapp/cli/internal/config"
)

// Client 表示GraphQL客户端
type Client struct {
	client *graphql.Client
	config *config.Config
}

// New 创建新的GraphQL客户端
func New(cfg *config.Config) (*Client, error) {
	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建GraphQL客户端
	client := graphql.NewClient(cfg.GetBackboardURL(), graphql.WithHTTPClient(httpClient))

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// NewAuthorized 创建带认证的GraphQL客户端
func NewAuthorized(cfg *config.Config) (*Client, error) {
	// 创建HTTP客户端并设置认证头
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	client := graphql.NewClient(cfg.GetBackboardURL(), graphql.WithHTTPClient(httpClient))

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// NewUnauthorized 创建无认证的GraphQL客户端
func NewUnauthorized(cfg *config.Config) (*Client, error) {
	return New(cfg)
}

// Query 执行GraphQL查询
func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	req := graphql.NewRequest(query)

	// 添加变量
	if variables != nil {
		for key, value := range variables {
			req.Var(key, value)
		}
	}

	// 设置认证头
	c.setAuthHeaders(req)

	return c.client.Run(ctx, req, response)
}

// Mutate 执行GraphQL变更
func (c *Client) Mutate(ctx context.Context, mutation string, variables map[string]interface{}, response interface{}) error {
	return c.Query(ctx, mutation, variables, response)
}

// setAuthHeaders 设置认证头
func (c *Client) setAuthHeaders(req *graphql.Request) {
	// 设置用户代理
	req.Header.Set("x-source", fmt.Sprintf("railway-cli/%s", "4.6.1"))
	req.Header.Set("user-agent", fmt.Sprintf("railway-cli/%s", "4.6.1"))

	// 设置认证头
	if token := config.GetRailwayToken(); token != nil {
		req.Header.Set("project-access-token", *token)
	} else if token := c.config.GetRailwayAuthToken(); token != nil {
		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", *token))
	}
}
