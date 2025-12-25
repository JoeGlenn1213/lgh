# LGH 安全指南

本文档介绍如何安全地部署和使用 LGH。

## 🔒 安全模型

LGH 默认绑定到 `127.0.0.1`，这意味着只有本机可以访问。当需要暴露到网络时，必须采取额外的安全措施。

## 安全级别

### 级别 1: 本地使用（默认，最安全）

```bash
lgh serve  # 默认 127.0.0.1:9418
```

- ✅ 只有本机可访问
- ✅ 无需额外配置
- ❌ 无法远程访问

### 级别 2: 内置 Basic Auth（推荐快速使用）

```bash
# 1. 设置认证
lgh auth setup

# 2. 启动服务（可绑定到网络）
lgh serve --bind 0.0.0.0

# 3. 客户端使用
git clone http://username:password@192.168.1.100:9418/repo.git
```

配置文件示例 (`~/.localgithub/config.yaml`):
```yaml
port: 9418
bind_address: "0.0.0.0"
read_only: true  # 推荐：只读模式
auth_enabled: true
auth_user: "git-user"
auth_password_hash: "salt:hash..."
```

### 级别 3: 反向代理 + TLS（推荐生产环境）

这是**最安全的方案**，LGH 仍绑定到 127.0.0.1，由成熟的反向代理处理认证和 TLS。

#### Caddy 配置（推荐）

```caddyfile
# Caddyfile
git.example.com {
    # 自动 HTTPS
    basicauth * {
        git-user $2a$14$hashhere...
    }
    reverse_proxy localhost:9418
}
```

```bash
# 启动 LGH（仅本地）
lgh serve

# 启动 Caddy
caddy run
```

#### Nginx 配置

```nginx
server {
    listen 443 ssl;
    server_name git.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    auth_basic "Git Access";
    auth_basic_user_file /etc/nginx/.htpasswd;

    location / {
        proxy_pass http://127.0.0.1:9418;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        
        # Git 需要的大超时
        proxy_read_timeout 3600;
        proxy_send_timeout 3600;
        client_max_body_size 0;  # 无限制
    }
}
```

生成 htpasswd:
```bash
htpasswd -c /etc/nginx/.htpasswd git-user
```

### 级别 4: 隧道服务（适合临时共享）

#### Cloudflare Tunnel + Access

```bash
# 1. 安装 cloudflared
brew install cloudflare/cloudflare/cloudflared

# 2. 创建隧道
cloudflared tunnel create lgh

# 3. 配置 Access 策略（在 Cloudflare 面板）
# 4. 运行隧道
cloudflared tunnel --url http://localhost:9418
```

#### ngrok + Basic Auth

```bash
# ngrok 支持内置认证
ngrok http 9418 --auth="user:password"
```

## 🛡️ 安全检查清单

### 部署前

- [ ] 使用强密码（至少 12 字符）
- [ ] 配置文件权限为 0600
- [ ] 考虑使用只读模式
- [ ] 检查防火墙规则

### 部署后

- [ ] 定期更新 LGH
- [ ] 监控访问日志
- [ ] 定期轮换密码

## 🚫 不推荐的做法

```bash
# ❌ 错误：直接暴露无认证
lgh serve --bind 0.0.0.0

# ❌ 错误：使用隧道但无认证
ngrok http 9418

# ❌ 错误：弱密码
lgh auth hash "123456"
```

## ✅ 推荐做法

```bash
# ✅ 正确：本地使用
lgh serve

# ✅ 正确：网络暴露 + 认证 + 只读
lgh auth setup
lgh serve --bind 0.0.0.0 --read-only

# ✅ 正确：反向代理（最佳）
lgh serve  # 只监听 localhost
caddy run  # 处理 TLS 和认证
```

## 📋 配置模板

### 最小安全配置

```yaml
# ~/.localgithub/config.yaml
port: 9418
bind_address: "127.0.0.1"
read_only: false
```

### 内网共享配置

```yaml
port: 9418
bind_address: "0.0.0.0"
read_only: true
auth_enabled: true
auth_user: "team"
auth_password_hash: "your-hash-here"
```

### 生产配置

```yaml
port: 9418
bind_address: "127.0.0.1"  # 只本地，反向代理处理网络
read_only: false
# 认证由反向代理处理
auth_enabled: false
```

## 🔐 密码哈希

LGH 使用 HMAC-SHA256 加盐哈希存储密码：

```bash
# 生成密码哈希
lgh auth hash

# 哈希格式：salt:hash
# 例如：a1b2c3d4e5:f6a7b8c9d0e1f2...
```

## 📞 报告安全问题

如发现安全漏洞，请发送邮件至 security@example.com，不要公开披露。
