package main

import (
	"fmt"
	"github.com/railwayapp/cli/internal/config"
	"os"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置初始化失败: %v\n", err)
		os.Exit(1)
	}

}
