# LGH v1.0.4 发布准备

**日期**: 2025-12-26
## v1.0.5 (Unreleased)

### New Features
- **Event System**: Introduced internal event bus and logging infrastructure.
- **Log Rotation**: Automatically rotates event logs > 10MB to ensure stability.
- **lgh events**: New command to view system activity with `--type` filtering and efficient reverse reading.
- **Git Push Tracking**: Server now explicitly captures push operations and logs reference changes with Commit IDs.

### Fixes
- **UI**: `lgh repo status` now correctly identifies the active remote based on upstream configuration.
- **Security**: Enforced Safe Bind (requires `--allow-unsafe` for public access). Config file permissions set to 0600.
- **Performance**: Event logging is now asynchronous to avoid blocking Git operations.
- **Reliability**: Improved error handling for repository reference tracking.
- **Reliability**: Guaranteed event flushing on CLI command exit.
- **Reliability**: Graceful shutdown ensures all events are logged.
- **Documentation**: Comprehensive README rewrite covering all v1.0.5 features and security guidelines.

## v1.0.9 (2025-12-26)

### New Features
- **One-Step Setup**: `lgh add . --push` now handles **everything**.
  - If the directory has no commits (or is fresh), it automatically performs `git add .` and `git commit -m "Initial commit by LGH"`.
  - Turns a folder of files into a hosted global repo in literally one command.
  - Warns if `.gitignore` might be missing (implicit warning via git output).

## v1.0.8 (2025-12-26)

### Improvements
- **Workflow**: `lgh add --push` now defaults to pushing `HEAD` (safer than guessing branch name).
- **UX**: Suppressed duplicate/confusing manual push instructions when auto-push is active.
- **Fixes**: Cleaned up internal instruction logic.

## v1.0.7 (2025-12-26)

### New Features
- **Workflow**: Added `--push` flag to `lgh add`.
  - `lgh add . --push`: Automatically pushes current branch to LGH remote after adding.
  - `lgh add . --push-branch <name>`: Pushes a specific branch.
  - Improves "out-of-the-box" experience by eliminating the manual `git push` step.

## v1.0.6 (2025-12-26)

### New Features
- **Routing**: Added **Virtual Owner Support**. LGH now explicitly supports URLs in the format `http://host/lgh/:repo.git` to satisfy tool requirements for `owner/repo` structure (e.g. Cursor, Terraform). The `/lgh/` prefix is automatically routed to the correct local repository. Note: Only `/lgh/` is supported as a virtual owner for security and consistency.

## v1.0.5 (2025-12-26)

## ✅ 新功能 (v1.0.4)

### 1. 核心仓库状态工具 (`lgh repo`)
- **lgh repo status**: 在任何 git 项目目录中，清晰展示本地与远程的连接状态
- **lgh repo inspect**: 查看 LGH 内部裸仓的详细信息 (HEAD, 分支, 最近提交)
- **lgh repo set-default**: 修改裸仓的默认分支 (HEAD symbolic-ref)

### 2. 远程切换器 (`lgh remote`)
- **lgh remote use**: 快速切换当前分支的上游 (upstream)，例如在 `lgh` 和 `origin` 之间切换

### 3. 便捷工具
- **lgh clone**: 语法糖，`lgh clone my-repo` 直接克隆本地仓库
- **lgh doctor**: 系统健康检查，检测环境、配置和端口问题

## 📝 变更文件
- `cmd/lgh/repo.go` (新增)
- `cmd/lgh/remote.go` (新增)
- `cmd/lgh/clone.go` (新增)
- `cmd/lgh/doctor.go` (新增)
- `cmd/lgh/main.go` (注册新命令，更新版本号)
- `internal/git/repo.go` (增强 git 功能支持)
- `pkg/ui/output.go` (增强 UI 支持)
- `Makefile` (版本号 1.0.4)
- `README.md` / `README.zh-CN.md` (文档更新)

---

## 🚀 发布步骤

### Step 1: 构建 release 二进制
```bash
make release
```

### Step 2: 提交代码
```bash
git add -A
git commit -m "v1.0.4: Add repo status/inspect, remote switcher, doctor, and clone commands"
git tag v1.0.4
git push origin main --tags
```

### Step 3: 创建 GitHub Release
- Tag: `v1.0.4`
- Title: `LGH v1.0.4 - The Repository Management Update`
- Upload binaries
- Copy SHA256

### Step 4: 更新 homebrew-tap
- 更新 `Formula/lgh.rb` 中的 URL 和 SHA256

---
