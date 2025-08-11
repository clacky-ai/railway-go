package railway

import (
	"context"

	igql "github.com/railwayapp/cli/internal/gql"
)

// User 用户信息
type User struct {
	ID     string
	Name   *string
	Email  string
	Avatar *string
}

// WhoAmI 返回当前认证用户信息
func (c *Client) WhoAmI(ctx context.Context) (*User, error) {
	var resp igql.UserMetaResponse
	if err := c.gqlClient.Query(ctx, igql.UserMetaQuery, nil, &resp); err != nil {
		return nil, err
	}
	return &User{ID: resp.Me.ID, Name: resp.Me.Name, Email: resp.Me.Email, Avatar: resp.Me.Avatar}, nil
}
