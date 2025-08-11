package railway

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

	"github.com/railwayapp/cli/internal/config"
	ignore "github.com/sabhiram/go-gitignore"
)

// UpParams 控制 Up 行为
type UpParams struct {
	ProjectID     string
	EnvironmentID string
	ServiceID     string

	ProjectRoot string // 项目根目录（用于 ignore 规则）
	Path        string // 需要部署的子路径（可为空）
	NoGitignore bool
	PathAsRoot  bool
	Verbose     bool
	Detach      bool
	CI          bool

	OnBuildLog      func(line string)
	OnDeploymentLog func(line string)
	OnStatus        func(status string)
}

// Up 打包上传并可选跟随日志，返回 (deploymentID, logsURL)
func (c *Client) Up(ctx context.Context, p UpParams) (string, string, error) {
	if strings.TrimSpace(p.ProjectRoot) == "" {
		return "", "", fmt.Errorf("ProjectRoot is required")
	}
	deployRoot := p.ProjectRoot
	if p.Path != "" {
		if !filepath.IsAbs(p.Path) {
			deployRoot = filepath.Join(p.ProjectRoot, p.Path)
		} else {
			deployRoot = p.Path
		}
	}

	// ignore 规则
	var gi *ignore.GitIgnore
	if !p.NoGitignore {
		var patterns []string
		if b, err := os.ReadFile(filepath.Join(p.ProjectRoot, ".railwayignore")); err == nil {
			patterns = append(patterns, strings.Split(string(b), "\n")...)
		}
		if b, err := os.ReadFile(filepath.Join(p.ProjectRoot, ".gitignore")); err == nil {
			patterns = append(patterns, strings.Split(string(b), "\n")...)
		}
		if len(patterns) > 0 {
			gi = ignore.CompileIgnoreLines(patterns...)
		}
	}

	if p.Verbose {
		fmt.Println("Indexing & archiving...")
	}

	// 打包 tar.gz
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
	rootForPrefix := p.ProjectRoot
	if p.PathAsRoot {
		rootForPrefix = deployRoot
	}
	if err := filepath.Walk(deployRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relFromRoot := path
		r, _ := filepath.Rel(rootForPrefix, path)
		relFromRoot = r
		if relFromRoot == "." {
			return nil
		}
		base := filepath.Base(path)
		if base == ".git" || base == "node_modules" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if gi != nil {
			relForIgnore, _ := filepath.Rel(p.ProjectRoot, path)
			relForIgnore = filepath.ToSlash(relForIgnore)
			if gi.MatchesPath(relForIgnore) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if info.Mode().IsRegular() {
			arcPath := filepath.ToSlash(filepath.Join(".", relFromRoot))
			return addFile(path, arcPath, info)
		}
		return nil
	}); err != nil {
		return "", "", fmt.Errorf("archive failed: %w", err)
	}
	if err := tw.Close(); err != nil {
		return "", "", err
	}
	if err := gz.Close(); err != nil {
		return "", "", err
	}

	if p.Verbose {
		fmt.Printf("archive bytes: %d\n", buf.Len())
	}

	// 上传
	host := c.cfg.GetHost()
	uploadURL := fmt.Sprintf("https://backboard.%s/project/%s/environment/%s/up?serviceId=%s", host, p.ProjectID, p.EnvironmentID, p.ServiceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if t := config.GetRailwayToken(); t != nil {
		req.Header.Set("project-access-token", *t)
	} else if t := c.cfg.GetRailwayAuthToken(); t != nil {
		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", *t))
	}
	req.Header.Set("x-source", fmt.Sprintf("railway-cli/%s", "4.6.1"))
	req.Header.Set("user-agent", fmt.Sprintf("railway-cli/%s", "4.6.1"))

	httpClient := &http.Client{Timeout: 300 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("upload failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	var raw map[string]any
	_ = json.Unmarshal(bodyBytes, &raw)
	deploymentID := getString(raw, "deployment_id", "deploymentId")
	logsURL := getString(raw, "logs_url", "logsUrl")

	if p.Detach {
		return deploymentID, logsURL, nil
	}

	// 如果需要，跟随日志（交由调用方通过订阅接口实现）
	if deploymentID == "" {
		return deploymentID, logsURL, nil
	}
	return deploymentID, logsURL, nil
}
