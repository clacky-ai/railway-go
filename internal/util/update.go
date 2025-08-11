package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
)

// GitHubRelease 表示GitHub发布信息
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdates 检查更新（后台运行）
func CheckForUpdates(currentVersion string) {
	// 简单的版本检查，实际项目中可能需要更复杂的逻辑
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://api.github.com/repos/railwayapp/cli/releases/latest")
	if err != nil {
		// 静默失败，不影响主要功能
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	// 简单的版本比较（实际项目中应该使用semver库）
	if release.TagName != "" && release.TagName != fmt.Sprintf("v%s", currentVersion) {
		yellow := color.New(color.FgYellow).SprintFunc()
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		purple := color.New(color.FgMagenta).SprintFunc()

		fmt.Printf("%s %s visit %s for more info\n",
			green("New version available:"),
			yellow(release.TagName),
			purple("https://docs.railway.com/guides/cli"),
		)
	}
}
