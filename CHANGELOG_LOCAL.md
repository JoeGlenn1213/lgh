# LGH 本地开发日志

## v1.3.0 (2026-04-08) - CI/CD Integration

### New Features

- **Push 元数据增强**: push 事件新增 `changed_files` 字段
  - 自动计算两个 commit 之间的文件变更列表
  - ActionD 可用于路径过滤，只触发相关插件

- **Commit Status API**: CI 状态回写
  - `GET /api/repos/{repo}/commits/{sha}/status` - 查询 CI 状态
  - `POST /api/repos/{repo}/commits/{sha}/status` - 写入 CI 状态
  - 支持 per-plugin 状态聚合

- **Status Storage**: 基于 JSON 文件的 commit 状态存储
  - `internal/git/status.go` - CommitStatus 结构体
  - 自动聚合整体状态 (pending/success/failure/error)

### Modified Files

| 文件 | 变更 |
|------|------|
| `internal/git/refs.go` | GetChangedFiles() 函数 |
| `internal/git/backend.go` | push 事件添加 changed_files |
| `internal/git/status.go` | 新建 CommitStatus 存储 |
| `internal/server/server.go` | /api/repos 路由 + status API |

### Integration

配合 ActionD v1.2.0 实现：
- 每次 push jobs: 15 → 3-5 (profile=fast)
- CI 结果可查询、可追溯

---

## v1.2.0 (2026-01-11) - Smart Archival System

### New Features
- **`lgh up "msg"`**: 一键起飞命令
  - 自动检测项目类型并生成 `.gitignore`（Python/Go/Node/Java/Rust/AI）
  - 执行 `git add .` + `git commit` + `git push`
  - 支持 `--force` 跳过垃圾检测，`--no-ignore` 跳过自动忽略
  - 首次使用可加 `-n <name>` 指定仓库名

- **`lgh save "msg"`**: 本地存档命令
  - 与 `lgh up` 类似，但不推送到远程
  - 适合 WIP 代码的临时保存

- **Smart Ignore**: 智能 `.gitignore` 生成
  - 自动检测项目类型：Python、Go、Node/TS、Java、Rust、AI/ML
  - 根据项目类型生成对应的 `.gitignore` 模板
  - 已集成到 `lgh add`、`lgh up`、`lgh save`

- **Trash Detection**: 垃圾预警系统
  - 大文件检测（>50MB 单文件阻断）
  - 敏感文件阻断（`.env`、`*.key` 等）
  - 危险目录检测（`node_modules/`、`__pycache__/` 等）

- **`lgh log`**: 服务运行日志查看
  - 查看服务启动、错误、警告等运行时日志
  - 支持 `--level ERROR` 按级别过滤
  - 支持 `--json` 输出 JSON 格式（供 AI/MCP 使用）
  - 支持 `--watch` 实时监控

- **`lgh mcp`**: MCP 服务器（AI Agent 集成）
  - 支持 stdio 传输模式（Cursor、Claude Desktop）
  - 9 个工具：lgh_status, lgh_list, lgh_add, lgh_remove, lgh_up, lgh_save, lgh_serve_start, lgh_serve_stop, lgh_log
  - 3 个资源：lgh://config, lgh://repos, lgh://server/status

- **Skill SDK (`pkg/skill/`)**: 能力接口
  - 可被外部项目 import 使用
  - 3 个内置 Skill：lgh.backup, lgh.status, lgh.list
  - 简洁接口：Skill.Meta() + Skill.Execute()

### New Files
- `internal/ignore/detect.go` - 项目类型检测
- `internal/ignore/templates.go` - Gitignore 模板
- `internal/ignore/trash.go` - 垃圾预警检测
- `internal/slog/slog.go` - 服务日志记录器
- `internal/mcp/server.go` - MCP 服务器核心
- `internal/mcp/tools.go` - MCP 工具处理器
- `cmd/lgh/up.go` - `lgh up` 命令
- `cmd/lgh/save.go` - `lgh save` 命令
- `cmd/lgh/log.go` - `lgh log` 命令
- `cmd/lgh/mcp.go` - `lgh mcp` 命令

### Modified Files
- `cmd/lgh/add.go` - 集成 Smart Ignore，新增 `--no-ignore` 参数
- `internal/server/server.go` - 集成 slog 服务日志

---

## v1.1.1 (2025-12-29) - Pending Release

### Fixes
- **Clone Directory Naming**: Fixed `lgh clone` creating directories with `.git` suffix.
  - Previously, `lgh clone ActionD` would create a directory named `ActionD.git` because git uses the URL's last path component directly.
  - Now we strip the `.git` suffix to create `ActionD` as expected.
  - Affected file: `internal/git/repo.go` - `CloneRepo()` function.

---

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
