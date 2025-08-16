package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

// 模板配置相关数据结构
type DeserializedServiceNetworking struct {
	ServiceDomains map[string]interface{} `json:"serviceDomains,omitempty"`
	TCPProxies     map[string]interface{} `json:"tcpProxies,omitempty"`
}

type DeserializedServiceVolumeMount struct {
	MountPath string `json:"mountPath"`
}

type DeserializedServiceVariable struct {
	DefaultValue *string `json:"defaultValue,omitempty"`
	Value        *string `json:"value,omitempty"`
	Description  *string `json:"description,omitempty"`
	IsOptional   *bool   `json:"isOptional,omitempty"`
}

type DeserializedServiceDeploy struct {
	HealthcheckPath *string `json:"healthcheckPath,omitempty"`
	StartCommand    *string `json:"startCommand,omitempty"`
}

type DeserializedServiceSource struct {
	Image         *string `json:"image,omitempty"`
	Repo          *string `json:"repo,omitempty"`
	RootDirectory *string `json:"rootDirectory,omitempty"`
}

type DeserializedTemplateService struct {
	Deploy       *DeserializedServiceDeploy                 `json:"deploy,omitempty"`
	Icon         *string                                    `json:"icon,omitempty"`
	Name         string                                     `json:"name"`
	Networking   *DeserializedServiceNetworking             `json:"networking,omitempty"`
	Source       *DeserializedServiceSource                 `json:"source,omitempty"`
	Variables    map[string]*DeserializedServiceVariable    `json:"variables,omitempty"`
	VolumeMounts map[string]*DeserializedServiceVolumeMount `json:"volumeMounts,omitempty"`
}

type DeserializedTemplateConfig struct {
	Services map[string]*DeserializedTemplateService `json:"services,omitempty"`
}

// NewDeployCommand 创建部署模板命令（对齐 Rust deploy.rs）
func NewDeployCommand(cfg *config.Config) *cobra.Command {
	var templates []string
	var variablePairs []string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "部署模板到项目",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(cfg, templates, variablePairs)
		},
	}

	cmd.Flags().StringArrayVarP(&templates, "template", "t", []string{}, "模板代码，可多次传入")
	// 注意: 不能使用 -v，已被全局 verbose 占用
	cmd.Flags().StringArrayVar(&variablePairs, "variable", []string{}, "模板变量，支持 KEY=VALUE 或 Service.Key=VALUE")

	return cmd
}

func runDeploy(cfg *config.Config, templates []string, variablePairs []string) error {
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}
	linked, err := cfg.GetLinkedProject()
	if err != nil {
		return err
	}

	// 若未指定模板，交互输入一个
	if len(templates) == 0 {
		t, err := util.PromptText("Select template to deploy")
		if err != nil {
			return fmt.Errorf("no template specified")
		}
		t = strings.TrimSpace(t)
		if t == "" {
			return fmt.Errorf("no template selected")
		}
		templates = []string{t}
	}

	// 解析变量对（支持 Service.Key 和 Key）
	userVars := parseTemplateVars(variablePairs)

	// 遍历每个模板进行部署
	for _, templateCode := range templates {
		templateCode = strings.TrimSpace(templateCode)
		if templateCode == "" {
			continue
		}

		// 使用新的 fetchAndCreate 方法
		if err := fetchAndCreate(gqlClient, cfg, templateCode, linked, userVars); err != nil {
			return fmt.Errorf("部署模板 %s 失败: %w", templateCode, err)
		}
	}

	return nil
}

func parseTemplateVars(pairs []string) map[string]string {
	vars := map[string]string{}
	for _, p := range pairs {
		idx := strings.IndexByte(p, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(p[:idx])
		val := strings.TrimSpace(p[idx+1:])
		if key == "" || val == "" {
			continue
		}
		vars[key] = val
	}
	return vars
}

// fetchAndCreate 获取模板详情并创建部署（对应 Rust 版本的 fetch_and_create）
func fetchAndCreate(
	gqlClient *client.Client,
	cfg *config.Config,
	templateCode string,
	linkedProject *config.LinkedProject,
	vars map[string]string,
) error {
	ctx := context.Background()

	// 1. 获取模板详情
	var templateDetail gql.TemplateDetailResponse
	if err := gqlClient.Query(ctx, gql.TemplateDetailQuery, map[string]interface{}{
		"code": templateCode,
	}, &templateDetail); err != nil {
		return fmt.Errorf("获取模板详情失败: %w", err)
	}

	// 2. 反序列化模板配置
	var templateConfig DeserializedTemplateConfig
	if len(templateDetail.Template.SerializedConfig) > 0 {
		if err := json.Unmarshal(templateDetail.Template.SerializedConfig, &templateConfig); err != nil {
			return fmt.Errorf("解析模板配置失败: %w", err)
		}
	}

	// 3. 确保项目和环境存在（这里假设已经通过 linkedProject 验证）
	if linkedProject.Project == "" || linkedProject.Environment == "" {
		return fmt.Errorf("项目或环境未正确链接")
	}

	// 4. 处理服务变量
	if templateConfig.Services != nil {
		for _, service := range templateConfig.Services {
			if service.Variables == nil {
				continue
			}

			for key, variable := range service.Variables {
				var value string
				var found bool

				// 优先级1: 服务特定变量 (Service.Key)
				if val, exists := vars[service.Name+"."+key]; exists {
					value = strings.TrimSpace(val)
					found = true
				} else if val, exists := vars[key]; exists {
					// 优先级2: 全局变量 (Key)
					value = strings.TrimSpace(val)
					found = true
				} else if variable.DefaultValue != nil && strings.TrimSpace(*variable.DefaultValue) != "" {
					// 优先级3: 默认值
					value = strings.TrimSpace(*variable.DefaultValue)
					found = true
				} else if variable.IsOptional != nil && *variable.IsOptional {
					// 优先级4: 可选变量，跳过
					continue
				} else {
					// 优先级5: 必需变量，提示用户输入
					description := ""
					if variable.Description != nil {
						description = fmt.Sprintf("   *%s*\n", *variable.Description)
					}

					prompt := fmt.Sprintf("Environment Variable %s for service %s is required, please set a value:\n%s",
						key, service.Name, description)

					inputValue, err := util.PromptText(prompt)
					if err != nil {
						return fmt.Errorf("获取环境变量 %s 失败: %w", key, err)
					}
					value = strings.TrimSpace(inputValue)
					found = true
				}

				if found && value != "" {
					variable.Value = &value
				}
			}
		}
	}

	// 5. 显示进度指示器
	sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	sp.Suffix = fmt.Sprintf(" Creating %s...", templateCode)
	sp.Start()

	// 6. 部署模板
	var deployResult gql.TemplateDeployResponse

	// 将 DeserializedTemplateConfig 转换为 SerializedTemplateConfig
	serializedConfig := make(gql.SerializedTemplateConfig)
	configBytes, err := json.Marshal(templateConfig)
	if err != nil {
		sp.Stop()
		return fmt.Errorf("序列化模板配置失败: %w", err)
	}
	if err := json.Unmarshal(configBytes, &serializedConfig); err != nil {
		sp.Stop()
		return fmt.Errorf("转换模板配置失败: %w", err)
	}

	deployInput := gql.TemplateDeployInput{
		ProjectID:        linkedProject.Project,
		EnvironmentID:    linkedProject.Environment,
		TemplateID:       templateDetail.Template.ID,
		SerializedConfig: serializedConfig,
	}

	if err := gqlClient.Mutate(ctx, gql.TemplateDeployMutation, map[string]interface{}{
		"projectId":        deployInput.ProjectID,
		"environmentId":    deployInput.EnvironmentID,
		"templateId":       deployInput.TemplateID,
		"serializedConfig": deployInput.SerializedConfig,
	}, &deployResult); err != nil {
		sp.Stop()
		return fmt.Errorf("部署模板失败: %w", err)
	}

	// 7. 显示成功消息
	sp.Stop()
	util.PrintSuccess(fmt.Sprintf("🎉 Added %s to project", templateDetail.Template.Name))

	return nil
}
