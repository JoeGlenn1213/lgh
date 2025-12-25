# LGH Security Guide

[中文](SECURITY.zh-CN.md)

This document describes how to securely deploy and use LGH.

## 🔒 Security Model

LGH binds to `127.0.0.1` by default, meaning only the local machine can access it. When exposing to a network, additional security measures must be taken.

## Security Levels

### Level 1: Local Use (Default, Most Secure)

```bash
lgh serve  # Default 127.0.0.1:9418
```

- ✅ Only accessible from local machine
- ✅ No additional configuration needed
- ❌ No remote access

### Level 2: Built-in Basic Auth (Quick Setup)

```bash
# 1. Setup authentication
lgh auth setup

# 2. Start server (can bind to network)
lgh serve --bind 0.0.0.0

# 3. Client usage
git clone http://username:password@192.168.1.100:9418/repo.git
```

Configuration example (`~/.localgithub/config.yaml`):
```yaml
port: 9418
bind_address: "0.0.0.0"
read_only: true  # Recommended: read-only mode
auth_enabled: true
auth_user: "git-user"
auth_password_hash: "salt:hash..."
```

### Level 3: Reverse Proxy + TLS (Production Recommended)

This is the **most secure approach**, with LGH still binding to 127.0.0.1 and a mature reverse proxy handling authentication and TLS.

#### Caddy Configuration (Recommended)

```caddyfile
# Caddyfile
git.example.com {
    # Automatic HTTPS
    basicauth * {
        git-user $2a$14$hashhere...
    }
    reverse_proxy localhost:9418
}
```

```bash
# Start LGH (local only)
lgh serve

# Start Caddy
caddy run
```

#### Nginx Configuration

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
        
        # Large timeouts needed for Git
        proxy_read_timeout 3600;
        proxy_send_timeout 3600;
        client_max_body_size 0;  # Unlimited
    }
}
```

Generate htpasswd:
```bash
htpasswd -c /etc/nginx/.htpasswd git-user
```

### Level 4: Tunnel Services (Temporary Sharing)

#### Cloudflare Tunnel + Access

```bash
# 1. Install cloudflared
brew install cloudflare/cloudflare/cloudflared

# 2. Create tunnel
cloudflared tunnel create lgh

# 3. Configure Access policy (in Cloudflare dashboard)
# 4. Run tunnel
cloudflared tunnel --url http://localhost:9418
```

#### ngrok + Basic Auth

```bash
# ngrok supports built-in auth
ngrok http 9418 --auth="user:password"
```

## 🛡️ Security Checklist

### Before Deployment

- [ ] Use strong passwords (at least 12 characters)
- [ ] Config file permissions set to 0600
- [ ] Consider using read-only mode
- [ ] Check firewall rules

### After Deployment

- [ ] Regularly update LGH
- [ ] Monitor access logs
- [ ] Rotate passwords periodically

## 🚫 Not Recommended

```bash
# ❌ Wrong: Direct exposure without auth
lgh serve --bind 0.0.0.0

# ❌ Wrong: Tunnel without auth
ngrok http 9418

# ❌ Wrong: Weak password
lgh auth hash "123456"
```

## ✅ Recommended Practices

```bash
# ✅ Correct: Local use
lgh serve

# ✅ Correct: Network exposure + auth + read-only
lgh auth setup
lgh serve --bind 0.0.0.0 --read-only

# ✅ Correct: Reverse proxy (best)
lgh serve  # Listen only on localhost
caddy run  # Handle TLS and auth
```

## 📋 Configuration Templates

### Minimal Secure Config

```yaml
# ~/.localgithub/config.yaml
port: 9418
bind_address: "127.0.0.1"
read_only: false
```

### Internal Network Sharing

```yaml
port: 9418
bind_address: "0.0.0.0"
read_only: true
auth_enabled: true
auth_user: "team"
auth_password_hash: "your-hash-here"
```

### Production Config

```yaml
port: 9418
bind_address: "127.0.0.1"  # Local only, reverse proxy handles network
read_only: false
# Authentication handled by reverse proxy
auth_enabled: false
```

## 🔐 Password Hashing

LGH uses HMAC-SHA256 with salt for password hashing:

```bash
# Generate password hash
lgh auth hash

# Hash format: salt:hash
# Example: a1b2c3d4e5:f6a7b8c9d0e1f2...
```

## 📞 Reporting Security Issues

If you discover a security vulnerability, please email security@example.com. Do not disclose publicly.
