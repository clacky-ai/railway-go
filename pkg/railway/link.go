package railway

import (
	"fmt"

	"github.com/railwayapp/cli/internal/config"
)

// LinkProjectToPath 将当前工作目录（由内部 Config 决定）链接到指定项目与环境
// 注意：此方法会写入用户配置文件（~/.railway/config*.json）
func (c *Client) LinkProjectToPath(projectID, environmentID string, projectName, environmentName *string) error {
	return c.cfg.LinkProject(projectID, environmentID, projectName, environmentName)
}

// LinkServiceToPath 将当前工作目录链接到指定服务
func (c *Client) LinkServiceToPath(serviceID string) error {
	return c.cfg.LinkService(serviceID)
}

// UnlinkProjectFromPath 取消当前工作目录的项目链接
func (c *Client) UnlinkProjectFromPath() error { return c.cfg.UnlinkProject() }

// UnlinkServiceFromPath 取消当前工作目录的服务链接
func (c *Client) UnlinkServiceFromPath() error { return c.cfg.UnlinkService() }

// GetLinkedContext 读取与当前目录最近的链接上下文
func (c *Client) GetLinkedContext() (*config.LinkedProject, error) {
	lp, err := c.cfg.GetLinkedProject()
	if err != nil {
		return nil, fmt.Errorf("no linked project: %w", err)
	}
	return lp, nil
}
