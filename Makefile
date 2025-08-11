# Railway CLI Makefile

.PHONY: build install clean test lint fmt help

# 变量
BINARY_NAME=railway
VERSION=4.6.1
BUILD_DIR=build
MAIN_PATH=cmd/railway/main.go

# 默认目标
help: ## 显示帮助信息
	@echo "Railway CLI Build Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 构建二进制文件
	@echo "构建Railway CLI..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

install: build ## 安装到系统
	@echo "安装Railway CLI到系统..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "安装完成: /usr/local/bin/$(BINARY_NAME)"

clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "清理完成"

test: ## 运行测试
	@echo "运行测试..."
	@go test -v ./...

lint: ## 运行代码检查
	@echo "运行代码检查..."
	@golangci-lint run

fmt: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...
	@goimports -w .

deps: ## 安装依赖
	@echo "安装依赖..."
	@go mod download
	@go mod tidy

dev: ## 开发模式运行
	@echo "开发模式运行..."
	@go run $(MAIN_PATH)

release: clean ## 构建发布版本
	@echo "构建发布版本..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -w -s" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -w -s" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION) -w -s" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION) -w -s" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "发布版本构建完成"
