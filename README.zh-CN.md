# LGH - LocalGitHub

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Platform-macOS%20|%20Linux%20|%20Windows-lightgrey" alt="Platform">
</p>

<p align="center">
  <a href="README.md">English</a>
</p>

**LGH (LocalGitHub)** 是一个轻量级本地 Git 托管服务。它通过封装 `git http-backend`，提供类似 GitHub 的 HTTP 访问能力，但完全运行在 localhost，实现"本地目录即 Git 服务"。

## ✨ 特性

- 🚀 **轻量高效** - 单一二进制文件，无需额外依赖
- 🔧 **简单易用** - 直观的 CLI 命令，一键添加仓库
- 🌐 **HTTP 访问** - 标准 Git HTTP 协议，兼容所有 Git 客户端
- 🔒 **安全认证** - 内置 Basic Auth 认证，密码加盐哈希存储
- 🛡️ **只读模式** - 可选的只读模式保护仓库安全
- 📡 **mDNS 发现** - 局域网自动发现，方便团队协作
- 🌍 **隧道支持** - 一键暴露到外网（支持 ngrok、cloudflared）

## 📦 安装

### 方式 1: 直接下载预编译版本 (推荐)

下载适合你系统的预编译二进制文件：

| 系统 | 架构 | 下载 |
|------|------|------|
| macOS | Apple Silicon (M1/M2/M3) | [lgh-1.0.3-darwin-arm64](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.3/lgh-1.0.3-darwin-arm64) |
| macOS | Intel | [lgh-1.0.3-darwin-amd64](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.3/lgh-1.0.3-darwin-amd64) |
| Linux | x86_64 | [lgh-1.0.3-linux-amd64](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.3/lgh-1.0.3-linux-amd64) |
| Linux | ARM64 | [lgh-1.0.3-linux-arm64](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.3/lgh-1.0.3-linux-arm64) |
| Windows | x86_64 | [lgh-1.0.3-windows-amd64.exe](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.3/lgh-1.0.3-windows-amd64.exe) |

```bash
# 下载后安装（以 macOS ARM64 为例）
chmod +x lgh-1.0.3-darwin-arm64
sudo mv lgh-1.0.3-darwin-arm64 /usr/local/bin/lgh
```

#### Windows 安装

1. 下载 `lgh-1.0.3-windows-amd64.exe`
2. 重命名为 `lgh.exe`
3. 移动到系统 `%PATH%` 路径下的文件夹中 (例如 `C:\Program Files\lgh\`)
4. 在 PowerShell 或 CMD 中运行


### 方式 2: 一键安装脚本

```bash
# 安装
curl -sSL https://raw.githubusercontent.com/JoeGlenn1213/lgh/main/install.sh | bash

# 卸载
curl -sSL https://raw.githubusercontent.com/JoeGlenn1213/lgh/main/uninstall.sh | bash
```

### 方式 3: Homebrew (macOS)

```bash
# 添加 tap
brew tap JoeGlenn1213/tap

# 安装
brew install lgh

# 卸载
brew uninstall lgh
```

### 方式 4: 从源码编译

```bash
git clone https://github.com/JoeGlenn1213/lgh.git
cd lgh
make build
sudo make install

# 或者手动
go build -o lgh ./cmd/lgh/
sudo mv lgh /usr/local/bin/
```

### 方式 5: Go Install

```bash
go install github.com/JoeGlenn1213/lgh/cmd/lgh@latest
```

## 🚀 快速开始

### 1. 初始化 LGH 环境

```bash
lgh init
```

这将在 `~/.localgithub/` 创建必要的目录和配置文件。

### 2. 启动服务器

```bash
# 前台启动
lgh serve

# 后台启动（守护进程模式）
lgh serve -d

# 查看服务器状态
lgh status

# 停止服务器
lgh stop
```

服务器默认监听 `http://127.0.0.1:9418`

### 3. 添加仓库

```bash
# 添加当前目录
cd your-project
lgh add .

# 或指定路径
lgh add /path/to/your/project

# 使用自定义名称
lgh add . --name my-awesome-project
```

### 4. 推送代码

```bash
git push lgh main
```

### 5. 在其他地方克隆

```bash
git clone http://127.0.0.1:9418/your-project.git
```

## 📖 命令参考

| 命令 | 说明 | 示例 |
|------|------|------|
| `lgh init` | 初始化 LGH 环境 | `lgh init` |
| `lgh serve` | 启动 HTTP 服务器 | `lgh serve -d` |
| `lgh stop` | 停止服务器 | `lgh stop` |
| `lgh add` | 添加仓库到 LGH | `lgh add . --name my-repo` |
| `lgh list` | 列出所有仓库（详细信息） | `lgh list` |
| `lgh status` | 查看服务状态和仓库列表 | `lgh status` |
| `lgh remove` | 移除仓库（先用 status 或 list 查看名称） | `lgh remove my-repo` |
| `lgh tunnel` | 暴露到外网 | `lgh tunnel --method ngrok` |
| `lgh auth` | 管理认证设置 | `lgh auth setup` |
| `lgh -v` | 显示版本 | `lgh -v` |
| `lgh doctor` | 检查系统健康状况 | `lgh doctor` |
| `lgh repo status` | 查看仓库连接状态 | `lgh repo status` |
| `lgh remote use` | 切换当前使用的远程 | `lgh remote use lgh` |
| `lgh clone` | 快速克隆 | `lgh clone repo-name` |

### 仓库管理工具 (v1.0.4+)

LGH 提供了一套工具来管理本地仓库状态，无需复杂的 git 命令。

#### 查看连接状态
一眼看清当前分支连接的是哪个远程服务：
```bash
lgh repo status
```

#### 切换远程
在 LGH 和 origin (如 GitHub) 之间快速切换：
```bash
lgh remote use lgh      # 切换上游到 LGH
lgh remote use origin   # 切换上游到 Origin
```

#### 其他实用工具
```bash
# 快速克隆 (无需完整 URL)
lgh clone my-project

# 检查裸仓详情 (HEAD, 分支等)
lgh repo inspect my-project

# 设置裸仓默认分支
lgh repo set-default my-project main

# 系统自检
lgh doctor
```

### 服务器选项

```bash
# 后台模式（守护进程）
lgh serve -d

# 只读模式（禁止 push）
lgh serve --read-only

# 自定义端口
lgh serve --port 8080

# 启用 mDNS 局域网发现
lgh serve --mdns

# 绑定到所有网卡（局域网访问）
lgh serve --bind 0.0.0.0
```

### 添加仓库选项

```bash
# 自定义名称
lgh add . --name custom-name

# 不自动添加 remote
lgh add . --no-remote
```

## 🔐 认证功能

当需要在网络上共享仓库时，建议启用认证保护：

### 设置认证

```bash
# 交互式设置（密码隐藏输入）
lgh auth setup

# 查看认证状态
lgh auth status

# 生成密码哈希（用于手动配置）
lgh auth hash

# 禁用认证
lgh auth disable
```

### 客户端认证

```bash
# 方式 1: URL 嵌入认证
git clone http://username:password@192.168.1.100:9418/repo.git

# 方式 2: 使用 Git 凭据助手
git config credential.helper store
git clone http://192.168.1.100:9418/repo.git
# 首次访问时输入用户名密码
```

### 安全最佳实践

| 场景 | 推荐配置 |
|------|----------|
| 本地开发 | 默认配置（127.0.0.1）|
| 内网共享 | `--bind 0.0.0.0 --read-only` + `auth setup` |
| 外网暴露 | 反向代理 (Caddy/Nginx) + TLS + 认证 |

> ⚠️ **安全提示**：密码必须至少 8 个字符。配置文件中存储的是加盐哈希，不是明文密码。

详细安全指南请参阅 [docs/SECURITY.md](docs/SECURITY.md)

## 🏗️ 目录结构

```
~/.localgithub/
├── config.yaml          # 全局配置
├── mappings.yaml        # 仓库映射
├── lgh.pid             # 服务 PID 文件
└── repos/              # 裸仓存储区
    ├── MyApp.git/
    └── ProjectB.git/
```

## ⚙️ 配置文件

`~/.localgithub/config.yaml`:

```yaml
port: 9418
bind_address: "127.0.0.1"
repos_dir: "/Users/you/.localgithub/repos"
read_only: false
mdns_enabled: false

# 认证配置（可选）
auth_enabled: true
auth_user: "git-user"
auth_password_hash: "salt:hash..."
```

## 🌐 隧道功能

LGH 支持多种方式将本地服务暴露到外网：

```bash
# 显示所有可用的隧道方法
lgh tunnel

# 使用 ngrok
lgh tunnel --method ngrok

# 使用 Cloudflare Tunnel
lgh tunnel --method cloudflared

# SSH 反向代理
lgh tunnel --method ssh
```

> ⚠️ **安全提示**：暴露到外网前，请务必启用认证 (`lgh auth setup`) 或使用反向代理。

## 🔧 高级用法

### 局域网共享（带认证）

```bash
# 1. 设置认证
lgh auth setup

# 2. 绑定到所有网卡并启用只读模式
lgh serve --bind 0.0.0.0 --read-only --mdns

# 3. 其他设备访问（需要认证）
git clone http://username:password@your-hostname.local:9418/repo.git
```

### 生产环境部署（推荐）

```bash
# 1. LGH 仅监听本地
lgh serve

# 2. 使用 Caddy 反向代理 + 自动 HTTPS
# Caddyfile:
# git.example.com {
#     basicauth * {
#         user $2a$14$...
#     }
#     reverse_proxy localhost:9418
# }
```

### 与 CI/CD 集成

```bash
# 临时暴露给 GitHub Actions
lgh auth setup
lgh tunnel --method ngrok &
# 使用 ngrok URL + 凭据
```

## ⚖️ 与其他方案对比

| 特性 | LGH | GitLab | Gitea | git daemon | 文件共享 |
|------|-----|--------|-------|------------|----------|
| 安装复杂度 | ⭐ 单文件 | ❌ 需要数据库 | ⚠️ 需要配置 | ⭐ 简单 | ⭐ 无需安装 |
| HTTP 协议 | ✅ | ✅ | ✅ | ❌ | ❌ |
| 身份验证 | ✅ 可选 | ✅ 必须 | ✅ 必须 | ❌ | ❌ |
| Web UI | ❌ | ✅ | ✅ | ❌ | ❌ |
| 资源占用 | ⭐ <10MB | ❌ >1GB | ⚠️ ~100MB | ⭐ <5MB | ⭐ 无 |
| 启动速度 | ⭐ <1s | ❌ >30s | ⚠️ ~10s | ⭐ <1s | ⭐ 即时 |
| 局域网发现 | ✅ mDNS | ❌ | ❌ | ❌ | ✅ |
| 适用场景 | 本地/临时 | 企业级 | 团队级 | 简单共享 | 文件传输 |

**LGH 的定位**：填补"简单文件共享"和"完整 Git 平台"之间的空白。

## 🧪 测试

```bash
# 运行所有测试
go test ./... -v

# 运行集成测试
go test ./test/... -v

# 跳过长时间运行的测试
go test ./... -v -short
```

## 📋 系统要求

- **Go 1.23+** (编译)
- **Git** (运行时)
- **macOS**、**Linux** 或 **Windows**




## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解如何参与。

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 License

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [spf13/cobra](https://github.com/spf13/cobra) - CLI 框架
- [spf13/viper](https://github.com/spf13/viper) - 配置管理
- [fatih/color](https://github.com/fatih/color) - 终端着色
- [hashicorp/mdns](https://github.com/hashicorp/mdns) - mDNS 支持

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/JoeGlenn1213">JoeGlenn1213</a>
</p>
