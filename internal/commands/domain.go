package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/spf13/cobra"
)

// NewDomainCommand 域名管理：创建服务域名或自定义域名
func NewDomainCommand(cfg *config.Config) *cobra.Command {
	var (
		port    int
		service string
		asJSON  bool
	)

	cmd := &cobra.Command{
		Use:   "domain [custom-domain]",
		Short: "为服务创建域名，或绑定自定义域名",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var customDomain string
			if len(args) > 0 {
				customDomain = args[0]
			}
			if strings.TrimSpace(customDomain) != "" {
				return runCreateCustomDomain(cfg, customDomain, port, service, asJSON)
			}
			return runCreateServiceDomain(cfg, service, asJSON)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 0, "将自定义域名映射到服务的端口")
	cmd.Flags().StringVarP(&service, "service", "s", "", "服务ID或名称（默认使用已链接服务）")
	cmd.Flags().BoolVar(&asJSON, "json", false, "以JSON格式输出")
	// 子命令: 删除域名
	cmd.AddCommand(newDomainDeleteCmd(cfg))
	return cmd
}

func newDomainDeleteCmd(cfg *config.Config) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "删除域名（支持服务域名或自定义域名）",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(id) == "" {
				return fmt.Errorf("请使用 -i/--id 指定要删除的域名ID")
			}
			return runDomainDelete(cfg, id)
		},
	}
	cmd.Flags().StringVarP(&id, "id", "i", "", "域名ID（service/custom domain 的ID）")
	return cmd
}

func runDomainDelete(cfg *config.Config, id string) error {
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}
	// 尝试两种删除（服务域名 / 自定义域名），任一成功即可
	var raw1 struct {
		ServiceDomainDelete bool `json:"serviceDomainDelete"`
	}
	if err := gqlClient.Mutate(context.Background(), gql.ServiceDomainDeleteMutation, map[string]any{"id": id}, &raw1); err == nil {
		if raw1.ServiceDomainDelete {
			fmt.Println("已删除服务域名")
			return nil
		}
	}
	var raw2 struct {
		CustomDomainDelete bool `json:"customDomainDelete"`
	}
	if err := gqlClient.Mutate(context.Background(), gql.CustomDomainDeleteMutation, map[string]any{"id": id}, &raw2); err == nil {
		if raw2.CustomDomainDelete {
			fmt.Println("已删除自定义域名")
			return nil
		}
	}
	return fmt.Errorf("删除域名失败：请确认id正确且当前账号有权限")
}

func runCreateServiceDomain(cfg *config.Config, serviceArg string, asJSON bool) error {
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}
	linked, err := cfg.GetLinkedProject()
	if err != nil {
		return err
	}
	var proj gql.ProjectResponse
	if err := gqlClient.Query(context.Background(), gql.ProjectQuery, map[string]any{"id": linked.Project}, &proj); err != nil {
		return err
	}

	// 解析服务
	var serviceID string
	if s := strings.TrimSpace(serviceArg); s != "" {
		for _, se := range proj.Project.Services.Edges {
			if eq(se.Node.ID, s) || eq(se.Node.Name, s) {
				serviceID = se.Node.ID
				break
			}
		}
		if serviceID == "" {
			return fmt.Errorf("未找到服务: %s", s)
		}
	} else if linked.Service != nil && *linked.Service != "" {
		serviceID = *linked.Service
	} else {
		return fmt.Errorf("未链接服务，请使用 -s/--service 指定服务ID或名称")
	}

	// 若已有域名则打印并退出
	var domainsResp struct {
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
	if err := gqlClient.Query(context.Background(), gql.DomainsQuery, map[string]any{
		"projectId":     linked.Project,
		"environmentId": linked.Environment,
		"serviceId":     serviceID,
	}, &domainsResp); err != nil {
		return err
	}
	total := len(domainsResp.Domains.ServiceDomains) + len(domainsResp.Domains.CustomDomains)
	if total > 0 {
		if asJSON {
			b, _ := json.MarshalIndent(domainsResp.Domains, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		fmt.Println("Domains already exists on the service:")
		if total == 1 {
			var d string
			if len(domainsResp.Domains.ServiceDomains) == 1 {
				d = domainsResp.Domains.ServiceDomains[0].Domain
			} else {
				d = domainsResp.Domains.CustomDomains[0].Domain
			}
			fmt.Printf("🚀 https://%s\n", d)
			return nil
		}
		for _, d := range domainsResp.Domains.CustomDomains {
			fmt.Printf("- https://%s\n", d.Domain)
		}
		for _, d := range domainsResp.Domains.ServiceDomains {
			fmt.Printf("- https://%s\n", d.Domain)
		}
		return nil
	}

	// 创建服务域名
	var resp struct {
		ServiceDomainCreate struct {
			ID     string `json:"id"`
			Domain string `json:"domain"`
		} `json:"serviceDomainCreate"`
	}
	if err := gqlClient.Mutate(context.Background(), gql.ServiceDomainCreateMutation, map[string]any{
		"environmentId": linked.Environment,
		"serviceId":     serviceID,
	}, &resp); err != nil {
		return err
	}
	if asJSON {
		out := map[string]any{"domain": fmt.Sprintf("https://%s", resp.ServiceDomainCreate.Domain)}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	fmt.Printf("Service Domain created:\n🚀 https://%s\n", resp.ServiceDomainCreate.Domain)
	return nil
}

func runCreateCustomDomain(cfg *config.Config, domain string, port int, serviceArg string, asJSON bool) error {
	gqlClient, err := client.NewAuthorized(cfg)
	if err != nil {
		return fmt.Errorf("请先登录: %w", err)
	}
	linked, err := cfg.GetLinkedProject()
	if err != nil {
		return err
	}
	var proj gql.ProjectResponse
	if err := gqlClient.Query(context.Background(), gql.ProjectQuery, map[string]any{"id": linked.Project}, &proj); err != nil {
		return err
	}
	// 解析服务
	var serviceID string
	if s := strings.TrimSpace(serviceArg); s != "" {
		for _, se := range proj.Project.Services.Edges {
			if eq(se.Node.ID, s) || eq(se.Node.Name, s) {
				serviceID = se.Node.ID
				break
			}
		}
		if serviceID == "" {
			return fmt.Errorf("未找到服务: %s", s)
		}
	} else if linked.Service != nil && *linked.Service != "" {
		serviceID = *linked.Service
	} else {
		return fmt.Errorf("未链接服务，请使用 -s/--service 指定服务ID或名称")
	}

	// 可用性检查
	var avail struct {
		CustomDomainAvailable struct {
			Available bool   `json:"available"`
			Message   string `json:"message"`
		} `json:"customDomainAvailable"`
	}
	if err := gqlClient.Query(context.Background(), gql.CustomDomainAvailableQuery, map[string]any{"domain": domain}, &avail); err != nil {
		return err
	}
	if !avail.CustomDomainAvailable.Available {
		return fmt.Errorf("domain is not available: %s", domain)
	}

	// 创建自定义域名
	var resp struct {
		CustomDomainCreate struct {
			ID     string `json:"id"`
			Domain string `json:"domain"`
			Status struct {
				DNSRecords []struct {
					Hostlabel     string `json:"hostlabel"`
					RecordType    string `json:"recordType"`
					RequiredValue string `json:"requiredValue"`
					Zone          string `json:"zone"`
				} `json:"dnsRecords"`
			} `json:"status"`
		} `json:"customDomainCreate"`
	}
	input := map[string]any{
		"domain":        domain,
		"environmentId": linked.Environment,
		"projectId":     linked.Project,
		"serviceId":     serviceID,
	}
	if port > 0 {
		input["targetPort"] = port
	}
	if err := gqlClient.Mutate(context.Background(), gql.CustomDomainCreateMutation, map[string]any{"input": input}, &resp); err != nil {
		return err
	}
	if asJSON {
		b, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	fmt.Printf("Domain created: %s\n", resp.CustomDomainCreate.Domain)
	if len(resp.CustomDomainCreate.Status.DNSRecords) == 0 {
		return nil
	}
	fmt.Printf("要完成自定义域名设置，请在 %s 添加以下DNS记录：\n\n", resp.CustomDomainCreate.Status.DNSRecords[0].Zone)
	// 打印简表
	fmt.Printf("\t%-8s%-16s%-s\n", "Type", "Name", "Value")
	for _, r := range resp.CustomDomainCreate.Status.DNSRecords {
		name := r.Hostlabel
		if strings.TrimSpace(name) == "" {
			name = "@"
		}
		fmt.Printf("\t%-8s%-16s%-s\n", r.RecordType, name, r.RequiredValue)
	}
	fmt.Println("\n注意：若 Name 为 \"@\"，表示根域名；DNS 变更可能需要最长 72 小时生效。")
	return nil
}
