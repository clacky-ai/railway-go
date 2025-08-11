package railway

import (
	"os"
	"strings"

	iclient "github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
)

// Option 用于配置 Client
type Option func(*options)

type options struct {
	apiToken     *string
	projectToken *string
	environment  *string
}

// WithAPIToken 使用 API Token（优先级：RAILWAY_TOKEN > RAILWAY_API_TOKEN > 配置文件 token）
// 注意：此实现通过进程环境变量注入，避免写入本地配置文件。
func WithAPIToken(token string) Option {
	return func(o *options) { o.apiToken = &token }
}

// WithProjectToken 使用项目访问 Token（将通过环境变量 RAILWAY_TOKEN 注入）
func WithProjectToken(token string) Option {
	return func(o *options) { o.projectToken = &token }
}

// WithEnvironment 指定运行环境，可选值："production"、"staging"、"dev"
func WithEnvironment(env string) Option {
	return func(o *options) { o.environment = &env }
}

// Client 面向外部使用者的 Railway 客户端
type Client struct {
	cfg       *config.Config
	gqlClient *iclient.Client
}

// New 创建 Client。若提供 WithAPIToken，将通过环境变量注入 token。
func New(opts ...Option) (*Client, error) {
	var o options
	for _, fn := range opts {
		fn(&o)
	}

	if o.apiToken != nil && strings.TrimSpace(*o.apiToken) != "" {
		_ = os.Setenv("RAILWAY_API_TOKEN", *o.apiToken)
	}
	if o.projectToken != nil && strings.TrimSpace(*o.projectToken) != "" {
		_ = os.Setenv("RAILWAY_TOKEN", *o.projectToken)
	}
	if o.environment != nil && strings.TrimSpace(*o.environment) != "" {
		_ = os.Setenv("RAILWAY_ENV", strings.TrimSpace(*o.environment))
	}

	cfg, err := config.New()
	if err != nil {
		return nil, err
	}
	gqlc, err := iclient.NewAuthorized(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{cfg: cfg, gqlClient: gqlc}, nil
}

func getString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
