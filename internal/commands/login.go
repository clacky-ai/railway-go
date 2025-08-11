package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/pkg/browser"
	"github.com/railwayapp/cli/internal/client"
	"github.com/railwayapp/cli/internal/config"
	"github.com/railwayapp/cli/internal/gql"
	"github.com/railwayapp/cli/internal/util"
	"github.com/spf13/cobra"
)

// NewLoginCommand 创建登录命令
func NewLoginCommand(cfg *config.Config) *cobra.Command {
	var browserless bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "登录到你的Railway账户",
		Long:  "登录到你的Railway账户以访问你的项目和服务。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cfg, browserless)
		},
	}

	cmd.Flags().BoolVarP(&browserless, "browserless", "b", false, "无浏览器登录")

	return cmd
}

func runLogin(cfg *config.Config, browserless bool) error {
	// 检查是否已有RAILWAY_TOKEN环境变量
	if token := config.GetRailwayAPIToken(); token != nil {
		gqlClient, err := client.NewAuthorized(cfg)
		if err == nil {
			if user, err := getUserInfo(gqlClient); err == nil {
				green := color.New(color.FgGreen, color.Bold).SprintFunc()
				fmt.Printf("%s found\n", green("RAILWAY_TOKEN"))
				printUser(user)
				return nil
			}
		}
		util.PrintError("Found invalid RAILWAY_TOKEN")
		return fmt.Errorf("无效的RAILWAY_TOKEN")
	}

	if browserless {
		return browserlessLogin(cfg)
	}

	// 询问是否打开浏览器
	openBrowser, err := util.PromptConfirm("打开浏览器登录?")
	if err != nil {
		return err
	}

	if !openBrowser {
		return browserlessLogin(cfg)
	}

	return browserLogin(cfg)
}

func browserLogin(cfg *config.Config) error {
	// 生成随机端口
	port := rand.Intn(10000) + 50000

	// 创建HTTP服务器
	router := mux.NewRouter()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	tokenChan := make(chan string, 1)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("https://%s", cfg.GetHost()))
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, PUT, PATCH, POST, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == "GET" {
			query := r.URL.Query()
			if token := query.Get("token"); token != "" {
				tokenChan <- token

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("https://%s", cfg.GetHost()))
				fmt.Fprintf(w, `{"status":"Ok","error":""}`)
				return
			}
		}

		http.Error(w, "Bad Request", http.StatusBadRequest)
	})

	// 启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务器错误: %v\n", err)
		}
	}()

	// 生成登录URL
	loginURL, err := generateLoginURL(cfg, port)
	if err != nil {
		return err
	}

	// 打开浏览器
	if err := browser.OpenURL(loginURL); err != nil {
		return browserlessLogin(cfg)
	}

	// 显示等待消息
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " 等待登录..."
	s.Start()

	// 等待令牌
	select {
	case token := <-tokenChan:
		s.Stop()
		server.Shutdown(context.Background())

		// 保存令牌
		if err := cfg.SetAuthToken(token); err != nil {
			return err
		}

		// 获取用户信息
		gqlClient, err := client.NewAuthorized(cfg)
		if err != nil {
			return err
		}

		user, err := getUserInfo(gqlClient)
		if err != nil {
			return err
		}

		printUser(user)
		return nil

	case <-time.After(5 * time.Minute):
		s.Stop()
		server.Shutdown(context.Background())
		return fmt.Errorf("登录超时")
	}
}

func browserlessLogin(cfg *config.Config) error {
	util.PrintInfo("无浏览器登录")

	// 创建登录会话
	gqlClient, err := client.NewUnauthorized(cfg)
	if err != nil {
		return err
	}

	var response gql.LoginSessionCreateResponse
	err = gqlClient.Mutate(context.Background(), gql.LoginSessionCreateMutation, nil, &response)
	if err != nil {
		return err
	}

	wordCode := response.LoginSessionCreate

	// 生成登录URL
	payload := fmt.Sprintf("wordCode=%s&hostname=%s", wordCode, getHostname())
	encodedPayload := base64.URLEncoding.EncodeToString([]byte(payload))
	loginURL := fmt.Sprintf("https://%s/cli-login?d=%s", cfg.GetHost(), encodedPayload)

	fmt.Printf("请访问:\n  %s\n", color.New(color.FgCyan, color.Bold, color.Underline).Sprint(loginURL))
	fmt.Printf("你的配对代码是: %s\n", color.New(color.FgMagenta, color.Bold).Sprint(wordCode))

	// 显示等待消息
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " 等待登录..."
	s.Start()

	// 轮询登录状态
	for {
		time.Sleep(1 * time.Second)

		variables := map[string]interface{}{
			"code": wordCode,
		}

		var consumeResponse gql.LoginSessionConsumeResponse
		err := gqlClient.Mutate(context.Background(), gql.LoginSessionConsumeMutation, variables, &consumeResponse)
		if err != nil {
			continue
		}

		if consumeResponse.LoginSessionConsume != nil {
			s.Stop()
			token := *consumeResponse.LoginSessionConsume

			// 保存令牌
			if err := cfg.SetAuthToken(token); err != nil {
				return err
			}

			// 获取用户信息
			authorizedClient, err := client.NewAuthorized(cfg)
			if err != nil {
				return err
			}

			user, err := getUserInfo(authorizedClient)
			if err != nil {
				return err
			}

			printUser(user)
			return nil
		}
	}
}

func getUserInfo(client *client.Client) (*gql.UserMetaResponse, error) {
	var response gql.UserMetaResponse
	err := client.Query(context.Background(), gql.UserMetaQuery, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func printUser(user *gql.UserMetaResponse) {
	if user.Me.Name != nil {
		fmt.Printf("已登录为 %s (%s)\n",
			color.New(color.FgGreen, color.Bold).Sprint(*user.Me.Name),
			user.Me.Email)
	} else {
		fmt.Printf("已登录为 %s\n", user.Me.Email)
	}
}

func generateLoginURL(cfg *config.Config, port int) (string, error) {
	code := generateRandomCode(32)
	hostname := getHostname()

	payload := fmt.Sprintf("port=%d&code=%s&hostname=%s", port, code, hostname)
	encodedPayload := base64.URLEncoding.EncodeToString([]byte(payload))

	return fmt.Sprintf("https://%s/cli-login?d=%s", cfg.GetHost(), encodedPayload), nil
}

func generateRandomCode(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "unknown"
}
