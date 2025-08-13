package commands

import (
	"fmt"

	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

// NewUnlinkCommand 创建取消链接命令
func NewUnlinkCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "取消链接当前目录与Railway项目",
		Long:  "取消当前目录与Railway项目的链接关系。",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.UnlinkProject(); err != nil {
				return fmt.Errorf("取消链接失败: %w", err)
			}
			util.PrintSuccess("已成功取消项目链接")
			return nil
		},
	}
	return cmd
}

// NewListCommand 创建列表命令
func NewListCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出项目、服务或部署",
		Long:  "列出你的Railway项目、服务或部署信息。",
		RunE: func(cmd *cobra.Command, args []string) error {
			util.PrintInfo("列表功能正在开发中...")
			return nil
		},
	}
	return cmd
}

// NewUpCommand 在 up.go 中实现

// NewDeployCommand 创建部署命令
func NewDeployCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "部署模板到项目",
		Long:  "部署Railway模板到你的项目中。",
		RunE: func(cmd *cobra.Command, args []string) error {
			util.PrintInfo("模板部署功能正在开发中...")
			return nil
		},
	}
	return cmd
}

// NewRedeployCommand 创建重新部署命令
func NewRedeployCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeploy",
		Short: "重新部署服务",
		Long:  "重新部署指定的服务或最新的部署。",
		RunE: func(cmd *cobra.Command, args []string) error {
			util.PrintInfo("重新部署功能正在开发中...")
			return nil
		},
	}
	return cmd
}

// NewDownCommand 在 down.go 中实现

// NewServiceCommand 创建服务管理命令
// service 命令已在 service.go 中实现

// NewStatusCommand 创建状态命令
func NewStatusCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "显示项目状态",
		Long:  "显示当前项目和服务的状态信息。",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 显示基本的链接信息
			project, err := cfg.GetLinkedProject()
			if err != nil {
				util.PrintWarning("当前目录未链接到任何项目")
				return nil
			}

			fmt.Printf("项目: %s\n", project.Project)
			if project.Name != nil {
				fmt.Printf("项目名称: %s\n", *project.Name)
			}
			fmt.Printf("环境: %s\n", project.Environment)
			if project.EnvironmentName != nil {
				fmt.Printf("环境名称: %s\n", *project.EnvironmentName)
			}
			if project.Service != nil {
				fmt.Printf("服务: %s\n", *project.Service)
			}
			fmt.Printf("项目路径: %s\n", project.ProjectPath)

			return nil
		},
	}
	return cmd
}

// NewVariablesCommand 创建环境变量命令
// variables 命令已在 variables.go 中实现

// NewRunCommand 创建运行命令
func NewRunCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [command]",
		Short: "使用Railway环境变量运行命令",
		Long:  "在本地使用Railway环境变量运行指定命令。",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			util.PrintInfo("本地运行功能正在开发中...")
			return nil
		},
	}
	return cmd
}

// NewOpenCommand 创建打开命令
func NewOpenCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open",
		Short: "在浏览器中打开项目",
		Long:  "在浏览器中打开Railway项目仪表板。",
		RunE: func(cmd *cobra.Command, args []string) error {
			util.PrintInfo("浏览器打开功能正在开发中...")
			return nil
		},
	}
	return cmd
}

// NewDocsCommand 创建文档命令
func NewDocsCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "打开Railway文档",
		Long:  "在浏览器中打开Railway官方文档。",
		RunE: func(cmd *cobra.Command, args []string) error {
			util.PrintInfo("文档功能正在开发中...")
			return nil
		},
	}
	return cmd
}

// NewCompletionCommand 创建自动补全命令
func NewCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "生成自动补全脚本",
		Long:      "生成指定shell的自动补全脚本。",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
			default:
				return fmt.Errorf("不支持的shell: %s", args[0])
			}
		},
	}
	return cmd
}
