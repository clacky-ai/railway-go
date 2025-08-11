# Railway CLI (Go版本)

[![CI](https://github.com/railwayapp/cli/actions/workflows/ci.yml/badge.svg)](https://github.com/railwayapp/cli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/railwayapp/cli)](https://goreportcard.com/report/github.com/railwayapp/cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

这是Railway CLI的Go语言重写版本，提供与原始Rust版本相同的功能。Railway CLI允许你从命令行与Railway平台交互，无需担心复杂的配置。

## ✨ 特性

- 🚀 **项目管理**: 创建、链接和管理Railway项目
- 🔐 **认证**: 安全的浏览器和无浏览器登录选项
- 📦 **部署**: 将应用部署到Railway平台
- 🌍 **环境变量**: 管理和使用环境变量
- 📊 **监控**: 查看部署状态和日志
- 🛠️ **服务管理**: 创建和管理服务
- 🎯 **模板部署**: 使用Railway模板快速部署

## 📦 安装

### 使用Go安装
```bash
go install github.com/railwayapp/cli/cmd/railway@latest
```

### 使用Homebrew (计划中)
```bash
brew install railway
```

### 从源码构建
```bash
git clone https://github.com/railwayapp/cli.git
cd cli/go
make build
```

### 使用Docker
```bash
docker pull ghcr.io/railwayapp/cli:latest
docker run --rm -it ghcr.io/railwayapp/cli:latest --help
```

## 🚀 快速开始

### 1. 登录到Railway
```bash
railway login
```

### 2. 初始化新项目
```bash
railway init my-awesome-project
```

### 3. 或链接现有项目
```bash
railway link
```

### 4. 部署应用
```bash
railway up
```

### 5. 查看状态
```bash
railway status
```

## 📚 命令参考

| 命令 | 描述 |
|------|------|
| `railway login` | 登录到Railway账户 |
| `railway logout` | 登出当前账户 |
| `railway whoami` | 显示当前用户信息 |
| `railway init` | 创建新项目 |
| `railway link` | 链接现有项目 |
| `railway unlink` | 取消项目链接 |
| `railway up` | 部署当前项目 |
| `railway deploy` | 部署模板 |
| `railway status` | 显示项目状态 |
| `railway logs` | 查看服务日志 |
| `railway variables` | 管理环境变量 |
| `railway run` | 使用环境变量运行命令 |
| `railway service` | 管理服务 |

## 🛠️ 开发

### 环境要求
- Go 1.21+
- Make (可选)

### 构建项目
```bash
# 使用Make
make build

# 或直接使用Go
go build -o railway cmd/railway/main.go
```

### 运行测试
```bash
make test
```

### 代码格式化
```bash
make fmt
```

### 代码检查
```bash
make lint
```

## 🏗️ 项目结构

```
go/
├── cmd/railway/           # 主程序入口
├── internal/
│   ├── client/           # GraphQL客户端
│   ├── config/           # 配置管理
│   ├── commands/         # CLI命令实现
│   ├── gql/             # GraphQL查询和变更
│   └── util/            # 工具函数
├── build/               # 构建输出
├── .github/workflows/   # GitHub Actions
├── Dockerfile          # Docker配置
├── Makefile           # 构建脚本
└── README.md          # 项目文档
```

## 🤝 贡献

我们欢迎所有形式的贡献！请查看我们的贡献指南。

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开Pull Request

## 📄 许可证

本项目使用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [Railway平台](https://railway.com)
- [官方文档](https://docs.railway.com)
- [原始Rust版本](https://github.com/railwayapp/cli)
- [问题反馈](https://github.com/railwayapp/cli/issues)

## 💬 支持

如有问题或建议，请：
- 提交 [Issue](https://github.com/railwayapp/cli/issues)
- 加入我们的 [Discord](https://discord.gg/railway)
- 查看 [文档](https://docs.railway.com)
