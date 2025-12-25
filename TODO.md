# LGH TODO & Future Improvements

This file tracks planned improvements and feature requests.

## 🔴 P0 - Critical (Before 1.1)

### Windows Auth Enhancement (DONE ✓)
- [x] Use `golang.org/x/term` for cross-platform password input
- [x] File: `cmd/lgh/auth.go`

## 🟡 P1 - High Priority

### Tunnel URL Capture
- [ ] Parse ngrok/cloudflared output to extract generated URL
- [ ] Print URL prominently in console
- [ ] Consider opening URL in browser automatically
- [ ] File: `internal/tunnel/tunnel.go`

### Enhanced Status Command
- [ ] Show number of hosted repositories
- [ ] Show total repository size
- [ ] Show last push timestamp per repo
- [ ] File: `cmd/lgh/status.go`

### PID File & Daemon Mode
- [ ] Consider using `sevlyar/go-daemon` for proper daemonization
- [ ] Better signal handling (SIGTERM, SIGINT)
- [ ] Handle orphan processes
- [ ] File: `internal/server/server.go`

## 🟢 P2 - Nice to Have

### Disk Space Check
- [ ] Check available disk space on `lgh init`
- [ ] Show repository sizes in `lgh list`
- [ ] Warn if disk is nearly full

### Git Environment (DONE ✓)
- [x] `GIT_HTTP_EXPORT_ALL=1` - Already implemented in `internal/git/backend.go:97`
- [x] `GIT_PROJECT_ROOT` - Already implemented

### Configuration Improvements
- [ ] Support environment variable overrides (LGH_PORT, LGH_BIND, etc.)
- [ ] Config file validation on load
- [ ] `lgh config` command to view/edit config

### Web UI (Future)
- [ ] Simple web dashboard showing repos
- [ ] Clone URL copy button
- [ ] Repository browser

## 📝 Technical Notes

### CGI Environment Variables
The following are set in `internal/git/backend.go`:
- `GIT_PROJECT_ROOT` - Root directory for repositories
- `GIT_HTTP_EXPORT_ALL=1` - Export all repos (no git-daemon-export-ok needed)
- `REMOTE_USER` - For logging purposes

### Tunnel Architecture
Current implementation wraps CLI tools:
- ngrok: `ngrok http <port>`
- cloudflared: `cloudflared tunnel --url http://localhost:<port>`
- SSH: Manual command generation

### Platform-Specific Code
- `cmd/lgh/auth.go` - Unix only (uses stty)
- `cmd/lgh/auth_windows.go` - Windows stub
- `internal/registry/lock_unix.go` - File locking (flock)
- `internal/registry/lock_windows.go` - File locking (mutex)

## 🐛 Known Issues

1. **Windows Auth**: `lgh auth` commands show guidance but don't work interactively
2. **Tunnel URL**: Generated URLs from ngrok/cloudflared not automatically captured

## 📅 Changelog

### v1.0.0 (Current)
- Initial release
- Basic Auth support
- Multi-platform builds (including Windows)
- Security hardening
