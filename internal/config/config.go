package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

// Environment 表示Railway环境
type Environment string

const (
	EnvironmentProduction Environment = "production"
	EnvironmentStaging    Environment = "staging"
	EnvironmentDev        Environment = "dev"
)

// LinkedProject 表示链接的项目配置
type LinkedProject struct {
	ProjectPath     string  `json:"projectPath"`
	Name            *string `json:"name,omitempty"`
	Project         string  `json:"project"`
	Environment     string  `json:"environment"`
	EnvironmentName *string `json:"environmentName,omitempty"`
	Service         *string `json:"service,omitempty"`
}

// RailwayUser 表示Railway用户配置
type RailwayUser struct {
	Token *string `json:"token,omitempty"`
}

// LinkedFunction 表示链接的函数
type LinkedFunction struct {
	Path string `json:"path"`
	ID   string `json:"id"`
}

// RailwayConfig 表示Railway配置文件
type RailwayConfig struct {
	Projects        map[string]LinkedProject `json:"projects"`
	User            RailwayUser              `json:"user"`
	LinkedFunctions []LinkedFunction         `json:"linkedFunctions,omitempty"`
}

// Config 表示配置管理器
type Config struct {
	rootConfig     RailwayConfig
	rootConfigPath string
}

// New 创建新的配置实例
func New() (*Config, error) {
	env := GetEnvironment()
	var configFile string

	switch env {
	case EnvironmentProduction:
		configFile = ".railway/config.json"
	case EnvironmentStaging:
		configFile = ".railway/config-staging.json"
	case EnvironmentDev:
		configFile = ".railway/config-dev.json"
	}

	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户主目录: %w", err)
	}

	configPath := filepath.Join(homeDir, configFile)

	config := &Config{
		rootConfigPath: configPath,
		rootConfig: RailwayConfig{
			Projects: make(map[string]LinkedProject),
			User:     RailwayUser{},
		},
	}

	// 尝试读取现有配置
	if data, err := ioutil.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config.rootConfig); err != nil {
			fmt.Fprintf(os.Stderr, "无法解析配置文件，重新生成: %v\n", err)
			config.rootConfig = RailwayConfig{
				Projects: make(map[string]LinkedProject),
				User:     RailwayUser{},
			}
		}
	}

	return config, nil
}

// GetEnvironment 获取当前环境
func GetEnvironment() Environment {
	env := os.Getenv("RAILWAY_ENV")
	switch env {
	case "production":
		return EnvironmentProduction
	case "staging":
		return EnvironmentStaging
	case "dev", "develop":
		return EnvironmentDev
	default:
		return EnvironmentProduction
	}
}

// GetRailwayToken 获取Railway令牌
func GetRailwayToken() *string {
	if token := os.Getenv("RAILWAY_TOKEN"); token != "" {
		return &token
	}
	return nil
}

// GetRailwayAPIToken 获取Railway API令牌
func GetRailwayAPIToken() *string {
	if token := os.Getenv("RAILWAY_API_TOKEN"); token != "" {
		return &token
	}
	return nil
}

// IsCI 检查是否在CI环境中
func IsCI() bool {
	ci := os.Getenv("CI")
	return ci == "true"
}

// GetRailwayAuthToken 获取Railway认证令牌（环境变量或配置文件）
func (c *Config) GetRailwayAuthToken() *string {
	if token := GetRailwayAPIToken(); token != nil {
		return token
	}
	if c.rootConfig.User.Token != nil && *c.rootConfig.User.Token != "" {
		return c.rootConfig.User.Token
	}
	return nil
}

// GetHost 获取Railway主机地址
func (c *Config) GetHost() string {
	switch GetEnvironment() {
	case EnvironmentProduction:
		return "railway.com"
	case EnvironmentStaging:
		return "railway-staging.com"
	case EnvironmentDev:
		return "railway-develop.com"
	default:
		return "railway.com"
	}
}

// GetBackboardURL 获取Backboard GraphQL端点
func (c *Config) GetBackboardURL() string {
	return fmt.Sprintf("https://backboard.%s/graphql/v2", c.GetHost())
}

// GetRelayHostPath 获取中继服务器主机路径
func (c *Config) GetRelayHostPath() string {
	return fmt.Sprintf("backboard.%s/relay", c.GetHost())
}

// GetCurrentDirectory 获取当前工作目录
func (c *Config) GetCurrentDirectory() (string, error) {
	return os.Getwd()
}

// GetClosestLinkedProjectDirectory 获取最近的链接项目目录
func (c *Config) GetClosestLinkedProjectDirectory() (string, error) {
	if GetRailwayToken() != nil {
		return c.GetCurrentDirectory()
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, exists := c.rootConfig.Projects[currentDir]; exists {
			return currentDir, nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}

	return "", fmt.Errorf("未找到链接的项目")
}

// GetLinkedProject 获取链接的项目
func (c *Config) GetLinkedProject() (*LinkedProject, error) {
	path, err := c.GetClosestLinkedProjectDirectory()
	if err != nil {
		return nil, err
	}

	if project, exists := c.rootConfig.Projects[path]; exists {
		return &project, nil
	}

	return nil, fmt.Errorf("未找到项目配置")
}

// LinkProject 链接项目
func (c *Config) LinkProject(projectID, environmentID string, name, environmentName *string) error {
	currentDir, err := c.GetCurrentDirectory()
	if err != nil {
		return err
	}

	project := LinkedProject{
		ProjectPath:     currentDir,
		Name:            name,
		Project:         projectID,
		Environment:     environmentID,
		EnvironmentName: environmentName,
	}

	c.rootConfig.Projects[currentDir] = project
	return c.Save()
}

// LinkService 链接服务
func (c *Config) LinkService(serviceID string) error {
	path, err := c.GetClosestLinkedProjectDirectory()
	if err != nil {
		return err
	}

	if project, exists := c.rootConfig.Projects[path]; exists {
		project.Service = &serviceID
		c.rootConfig.Projects[path] = project
		return c.Save()
	}

	return fmt.Errorf("未找到项目配置")
}

// UnlinkProject 取消链接项目
func (c *Config) UnlinkProject() error {
	path, err := c.GetClosestLinkedProjectDirectory()
	if err != nil {
		return err
	}

	delete(c.rootConfig.Projects, path)
	return c.Save()
}

// UnlinkService 取消链接服务
func (c *Config) UnlinkService() error {
	path, err := c.GetClosestLinkedProjectDirectory()
	if err != nil {
		return err
	}

	if project, exists := c.rootConfig.Projects[path]; exists {
		project.Service = nil
		c.rootConfig.Projects[path] = project
		return c.Save()
	}

	return fmt.Errorf("未找到项目配置")
}

// SetAuthToken 设置认证令牌
func (c *Config) SetAuthToken(token string) error {
	c.rootConfig.User.Token = &token
	return c.Save()
}

// Reset 重置配置
func (c *Config) Reset() error {
	c.rootConfig = RailwayConfig{
		Projects: make(map[string]LinkedProject),
		User:     RailwayUser{},
	}
	return c.Save()
}

// Save 保存配置到文件
func (c *Config) Save() error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(c.rootConfigPath), 0755); err != nil {
		return fmt.Errorf("无法创建配置目录: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(&c.rootConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("无法序列化配置: %w", err)
	}

	// 原子写入
	tempFile := c.rootConfigPath + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("无法写入临时配置文件: %w", err)
	}

	if err := os.Rename(tempFile, c.rootConfigPath); err != nil {
		return fmt.Errorf("无法移动配置文件: %w", err)
	}

	return nil
}
