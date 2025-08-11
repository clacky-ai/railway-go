package commands

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"
)

// NewUpCommand 创建上传/部署命令
func NewUpCommand(cfg *config.Config) *cobra.Command {
	var (
		path        string
		detach      bool
		ci          bool
		service     string
		environment string
		noGitignore bool
		pathAsRoot  bool
		verbose     bool
	)

	cmd := &cobra.Command{
		Use:   "up",
		Short: "部署当前项目",
		Long:  "将当前项目打包上传到Railway并触发部署。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUp(cfg, path, detach, ci, service, environment, noGitignore, pathAsRoot, verbose)
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "部署路径（默认当前项目目录）")
	cmd.Flags().BoolVarP(&detach, "detach", "d", false, "不附加日志流")
	cmd.Flags().BoolVarP(&ci, "ci", "c", false, "仅流式构建日志后退出（等价于CI模式）")
	cmd.Flags().StringVarP(&service, "service", "s", "", "要部署到的服务ID（默认使用已链接服务）")
	cmd.Flags().StringVarP(&environment, "environment", "e", "", "要部署到的环境ID（默认使用已链接环境）")
	cmd.Flags().BoolVar(&noGitignore, "no-gitignore", false, "不要读取 .gitignore 规则")
	cmd.Flags().BoolVar(&pathAsRoot, "path-as-root", false, "使用 --path 作为归档前缀（默认为项目根）")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "详细输出")

	return cmd
}

func runUp(cfg *config.Config, path string, detach, ci bool, service, environment string, noGitignore, pathAsRoot, verbose bool) error {
	httpClient := &http.Client{Timeout: 300 * time.Second}

	linked, err := cfg.GetLinkedProject()
	if err != nil {
		return fmt.Errorf("未找到已链接的项目: %w", err)
	}

	if environment == "" {
		environment = linked.Environment
	}
	if service == "" && linked.Service != nil {
		service = *linked.Service
	}

	projectDir := linked.ProjectPath
	deployRoot := projectDir
	if path != "" {
		if !filepath.IsAbs(path) {
			deployRoot = filepath.Join(projectDir, path)
		} else {
			deployRoot = path
		}
	}

	// 加载 ignore 规则
	var gi *ignore.GitIgnore
	if !noGitignore {
		var patterns []string
		// .railwayignore
		if b, err := os.ReadFile(filepath.Join(projectDir, ".railwayignore")); err == nil {
			lines := strings.Split(string(b), "\n")
			patterns = append(patterns, lines...)
		}
		// .gitignore
		if b, err := os.ReadFile(filepath.Join(projectDir, ".gitignore")); err == nil {
			lines := strings.Split(string(b), "\n")
			patterns = append(patterns, lines...)
		}
		if len(patterns) > 0 {
			gi = ignore.CompileIgnoreLines(patterns...)
		}
	}

	if verbose {
		fmt.Println("Indexing & archiving...")
	}

	// 创建 tar.gz 归档
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)

	addFile := func(absPath, arcPath string, info os.FileInfo) error {
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = arcPath
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		f, err := os.Open(absPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}
		return nil
	}

	rootForPrefix := projectDir
	if pathAsRoot {
		rootForPrefix = deployRoot
	}

	err = filepath.Walk(deployRoot, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relFromRoot := p
		if !pathAsRoot {
			// 归档前缀统一从项目根开始
			r, _ := filepath.Rel(rootForPrefix, p)
			relFromRoot = r
		} else {
			// 使用 path 作为归档前缀
			r, _ := filepath.Rel(rootForPrefix, p)
			relFromRoot = r
		}
		if relFromRoot == "." {
			return nil
		}

		// 忽略 .git 与 node_modules
		base := filepath.Base(p)
		if base == ".git" || base == "node_modules" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 应用 ignore 规则（相对项目根）
		if gi != nil {
			relForIgnore, _ := filepath.Rel(projectDir, p)
			relForIgnore = filepath.ToSlash(relForIgnore)
			if gi.MatchesPath(relForIgnore) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.Mode().IsRegular() {
			// 归档路径以 "." 为根
			arcPath := filepath.ToSlash(filepath.Join(".", relFromRoot))
			return addFile(p, arcPath, info)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("打包失败: %w", err)
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("archive bytes: %d\n", buf.Len())
	}

	host := cfg.GetHost()
	uploadURL := fmt.Sprintf("https://backboard.%s/project/%s/environment/%s/up?serviceId=%s", host, linked.Project, environment, service)

	// 直接发送 tar.gz 数据
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, uploadURL, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if t := config.GetRailwayToken(); t != nil {
		req.Header.Set("project-access-token", *t)
	} else if t := cfg.GetRailwayAuthToken(); t != nil {
		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", *t))
	}
	req.Header.Set("x-source", fmt.Sprintf("railway-cli/%s", "4.6.1"))
	req.Header.Set("user-agent", fmt.Sprintf("railway-cli/%s", "4.6.1"))

	sp := spinner.New(spinner.CharSets[14], 90*time.Millisecond)
	sp.Suffix = " Uploading"
	sp.Start()
	resp, err := httpClient.Do(req)
	if err != nil {
		sp.Stop()
		return fmt.Errorf("上传失败: %w", err)
	}
	defer resp.Body.Close()
	sp.Stop()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("上传失败，状态码 %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	// 解析响应体（兼容下划线和驼峰）
	bodyBytes, _ := io.ReadAll(resp.Body)
	var raw map[string]any
	_ = json.Unmarshal(bodyBytes, &raw)

	getStr := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := raw[k]; ok {
				if s, ok := v.(string); ok {
					return s
				}
			}
		}
		return ""
	}

	deploymentID := getStr("deployment_id", "deploymentId")
	logsURL := getStr("logs_url", "logsUrl")

	if logsURL != "" {
		fmt.Printf("  Build Logs: %s\n", logsURL)
	} else {
		fmt.Println("  Build Logs: (响应未返回 logsUrl)")
	}

	if detach {
		return nil
	}

	ciMode := ci || config.IsCI()
	if !isTerminalStdout() && !ciMode {
		return nil
	}

	if deploymentID == "" {
		fmt.Println("警告：未获取到deployment_id，日志订阅可能无法开始")
		return nil
	}

	// 并发启动日志订阅与状态订阅
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 构建日志
	go func() {
		vars := map[string]interface{}{"deploymentId": deploymentID, "filter": "", "limit": 500}
		_ = client.Subscribe(ctx, cfg, gql.BuildLogsSub, vars, func(data json.RawMessage) {
			var pl gql.BuildLogsPayload
			if err := json.Unmarshal(data, &pl); err == nil {
				for _, l := range pl.BuildLogs {
					fmt.Println(l.Message)
					if ciMode && strings.HasPrefix(l.Message, "No changed files matched patterns") {
						cancel()
						os.Exit(0)
					}
				}
			}
		}, func(err error) { fmt.Fprintf(os.Stderr, "构建日志订阅错误: %v\n", err) })
	}()

	// 部署日志（非CI模式）
	if !ciMode {
		go func() {
			vars := map[string]interface{}{"deploymentId": deploymentID, "filter": "", "limit": 500}
			_ = client.Subscribe(ctx, cfg, gql.DeploymentLogsSub, vars, func(data json.RawMessage) {
				var pl gql.DeploymentLogsPayload
				if err := json.Unmarshal(data, &pl); err == nil {
					for _, l := range pl.DeploymentLogs {
						fmt.Println(formatAttrLog(l.Message, l.Attributes))
					}
				}
			}, func(err error) { fmt.Fprintf(os.Stderr, "部署日志订阅错误: %v\n", err) })
		}()
	}

	// 状态订阅
	statusDone := make(chan struct{})
	go func() {
		vars := map[string]interface{}{"id": deploymentID}
		_ = client.Subscribe(ctx, cfg, gql.DeploymentStatusSub, vars, func(data json.RawMessage) {
			var st gql.DeploymentStatusPayload
			if err := json.Unmarshal(data, &st); err == nil {
				switch strings.ToUpper(st.Deployment.Status) {
				case "SUCCESS":
					fmt.Println("Deploy complete")
					if ciMode {
						os.Exit(0)
					}
					close(statusDone)
				case "FAILED":
					fmt.Println("Deploy failed")
					os.Exit(1)
				case "CRASHED":
					fmt.Println("Deploy crashed")
					os.Exit(1)
				}
			}
		}, func(err error) { fmt.Fprintf(os.Stderr, "状态订阅错误: %v\n", err) })
	}()

	<-statusDone
	return nil
}

func formatAttrLog(message string, attrs []struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}) string {
	if len(attrs) == 0 {
		return message
	}
	var b strings.Builder
	b.WriteString(message)
	b.WriteString(" ")
	for i, a := range attrs {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(a.Key)
		b.WriteString("=")
		b.WriteString(a.Value)
	}
	return b.String()
}

func isTerminalStdout() bool {
	return true
}
