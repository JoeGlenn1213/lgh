# LGH - LocalGitHub

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Platform-macOS%20|%20Linux%20|%20Windows-lightgrey" alt="Platform">
</p>

<p align="center">
  <a href="README.zh-CN.md">中文文档</a>
</p>

**LGH (LocalGitHub)** is a lightweight local Git hosting service. It wraps `git http-backend` to provide GitHub-like HTTP access, running entirely on localhost - turning your local directory into a Git server.

## ✨ Features

- 🚀 **Lightweight** - Single binary, no external dependencies
- 🔧 **Easy to Use** - Intuitive CLI commands, one-click repository setup
- 🌐 **HTTP Access** - Standard Git HTTP protocol, compatible with all Git clients
- 🔒 **Authentication** - Built-in Basic Auth with salted password hashing
- 🛡️ **Read-Only Mode** - Optional read-only mode to protect repositories
- 📡 **mDNS Discovery** - Automatic LAN discovery for team collaboration
- 🌍 **Tunnel Support** - One-click expose to internet (ngrok, cloudflared)

## 📦 Installation

### Option 1: Download Pre-built Binary (Recommended)

Download the pre-built binary for your system:

| OS | Architecture | Download |
|------|------|------|
| macOS | Apple Silicon (M1/M2/M3) | [lgh-darwin-arm64](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.1/lgh-darwin-arm64) |
| macOS | Intel | [lgh-darwin-amd64](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.1/lgh-darwin-amd64) |
| Linux | x86_64 | [lgh-linux-amd64](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.1/lgh-linux-amd64) |
| Linux | ARM64 | [lgh-linux-arm64](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.1/lgh-linux-arm64) |
| Windows | x86_64 | [lgh-windows-amd64.exe](https://github.com/JoeGlenn1213/lgh/releases/download/v1.0.1/lgh-windows-amd64.exe) |

```bash
# Install after download (macOS ARM64 example)
chmod +x lgh-darwin-arm64
sudo mv lgh-darwin-arm64 /usr/local/bin/lgh
```

#### Windows Installation

1. Download `lgh-windows-amd64.exe`
2. Rename to `lgh.exe`
3. Move to a folder in your `%PATH%` (e.g., `C:\Program Files\lgh\`)
4. Run in PowerShell or Command Prompt


### Option 2: Install Script

```bash
# Install
curl -sSL https://raw.githubusercontent.com/JoeGlenn1213/lgh/main/install.sh | bash

# Uninstall
curl -sSL https://raw.githubusercontent.com/JoeGlenn1213/lgh/main/uninstall.sh | bash
```

### Option 3: Homebrew (macOS)

```bash
# Add tap
brew tap JoeGlenn1213/tap

# Install
brew install lgh

# Uninstall
brew uninstall lgh
```

### Option 4: Build from Source

```bash
git clone https://github.com/JoeGlenn1213/lgh.git
cd lgh
make build
sudo make install

# Or manually
go build -o lgh ./cmd/lgh/
sudo mv lgh /usr/local/bin/
```

### Option 5: Go Install

```bash
go install github.com/JoeGlenn1213/lgh/cmd/lgh@latest
```

## 🚀 Quick Start

### 1. Initialize LGH Environment

```bash
lgh init
```

This creates the necessary directories and config files in `~/.localgithub/`.

### 2. Start the Server

```bash
lgh serve
```

Server listens on `http://127.0.0.1:9418` by default.

### 3. Add a Repository

```bash
# Add current directory
cd your-project
lgh add .

# Or specify a path
lgh add /path/to/your/project

# Use custom name
lgh add . --name my-awesome-project
```

### 4. Push Code

```bash
git push lgh main
```

### 5. Clone from Elsewhere

```bash
git clone http://127.0.0.1:9418/your-project.git
```

## 📖 Command Reference

| Command | Description | Example |
|------|------|------|
| `lgh init` | Initialize LGH environment | `lgh init` |
| `lgh serve` | Start HTTP server | `lgh serve --port 8080` |
| `lgh add` | Add repository to LGH | `lgh add . --name my-repo` |
| `lgh list` | List all repositories | `lgh list` |
| `lgh status` | View server status | `lgh status` |
| `lgh remove` | Remove repository | `lgh remove my-repo` |
| `lgh tunnel` | Expose to internet | `lgh tunnel --method ngrok` |
| `lgh auth` | Manage authentication | `lgh auth setup` |

### Server Options

```bash
# Read-only mode (disable push)
lgh serve --read-only

# Custom port
lgh serve --port 8080

# Enable mDNS for LAN discovery
lgh serve --mdns

# Bind to all interfaces (LAN access)
lgh serve --bind 0.0.0.0
```

### Add Repository Options

```bash
# Custom name
lgh add . --name custom-name

# Don't auto-add remote
lgh add . --no-remote
```

## 🔐 Authentication

Enable authentication when sharing repositories over the network:

### Setup Authentication

```bash
# Interactive setup (hidden password input)
lgh auth setup

# View auth status
lgh auth status

# Generate password hash (for manual config)
lgh auth hash

# Disable auth
lgh auth disable
```

### Client Authentication

```bash
# Method 1: URL embedded credentials
git clone http://username:password@192.168.1.100:9418/repo.git

# Method 2: Use Git credential helper
git config credential.helper store
git clone http://192.168.1.100:9418/repo.git
# Enter username/password on first access
```

### Security Best Practices

| Scenario | Recommended Config |
|------|----------|
| Local development | Default config (127.0.0.1) |
| LAN sharing | `--bind 0.0.0.0 --read-only` + `auth setup` |
| Internet exposure | Reverse proxy (Caddy/Nginx) + TLS + Auth |

> ⚠️ **Security Note**: Password must be at least 8 characters. Config file stores salted hash, not plaintext.

See [docs/SECURITY.md](docs/SECURITY.md) for detailed security guidelines.

## 🏗️ Directory Structure

```
~/.localgithub/
├── config.yaml          # Global config
├── mappings.yaml        # Repository mappings
├── lgh.pid             # Server PID file
└── repos/              # Bare repository storage
    ├── MyApp.git/
    └── ProjectB.git/
```

## ⚙️ Configuration

`~/.localgithub/config.yaml`:

```yaml
port: 9418
bind_address: "127.0.0.1"
repos_dir: "/Users/you/.localgithub/repos"
read_only: false
mdns_enabled: false

# Authentication (optional)
auth_enabled: true
auth_user: "git-user"
auth_password_hash: "salt:hash..."
```

## 🌐 Tunnel Feature

LGH supports multiple ways to expose your local service to the internet:

```bash
# Show all available tunnel methods
lgh tunnel

# Use ngrok
lgh tunnel --method ngrok

# Use Cloudflare Tunnel
lgh tunnel --method cloudflared

# SSH reverse proxy
lgh tunnel --method ssh
```

> ⚠️ **Security Note**: Always enable authentication (`lgh auth setup`) or use a reverse proxy before exposing to the internet.

## 🔧 Advanced Usage

### LAN Sharing (with Auth)

```bash
# 1. Setup auth
lgh auth setup

# 2. Bind to all interfaces with read-only mode
lgh serve --bind 0.0.0.0 --read-only --mdns

# 3. Access from other devices (requires auth)
git clone http://username:password@your-hostname.local:9418/repo.git
```

### Production Deployment (Recommended)

```bash
# 1. LGH listens on localhost only
lgh serve

# 2. Use Caddy reverse proxy + auto HTTPS
# Caddyfile:
# git.example.com {
#     basicauth * {
#         user $2a$14$...
#     }
#     reverse_proxy localhost:9418
# }
```

### CI/CD Integration

```bash
# Temporarily expose for GitHub Actions
lgh auth setup
lgh tunnel --method ngrok &
# Use ngrok URL + credentials
```

## 🧪 Testing

```bash
# Run all tests
go test ./... -v

# Run integration tests
go test ./test/... -v

# Skip long-running tests
go test ./... -v -short
```

## 📋 System Requirements

- **Go 1.23+** (for building)
- **Git** (runtime)
- **macOS**, **Linux**, or **Windows**




## 🤝 Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Create Pull Request

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file.

## 🙏 Acknowledgments

- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [spf13/viper](https://github.com/spf13/viper) - Configuration management
- [fatih/color](https://github.com/fatih/color) - Terminal coloring
- [hashicorp/mdns](https://github.com/hashicorp/mdns) - mDNS support

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/JoeGlenn1213">JoeGlenn1213</a>
</p>
