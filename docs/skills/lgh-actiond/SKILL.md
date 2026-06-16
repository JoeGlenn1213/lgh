---
name: lgh-actiond
description: >
  Working with LGH (LocalGitHub) + ActionD + ActionD-Web — Joe's local Git hosting and CI/CD stack.
  Architecture, MCP tools, REST API, plugin system, development workflow, known pitfalls.
trigger:
  - User asks about LGH, ActionD, lgh CLI, local CI/CD, local Git hosting
  - User wants to push code with `lgh up` and see CI results
  - Working in /Users/fenge1222/neil/LocalGitHub/ (lgh/, ActionD/, actiond-web/)
  - Debugging ActionD plugin dispatch, event flow, or MCP tools
  - User says "lgh up", "actiond", "local CI", "CI pipeline"
---

# LGH + ActionD — 本地 CI/CD 全栈文档

## 一句话定位

**LGH** = 本地 GitHub（Git 托管，端口 9418）
**ActionD** = 本地 CI/CD 引擎（监听 push 事件，跑插件，端口 3000）
**ActionD-Web** = Web 控制台（由 ActionD 静态服务）

核心工作流：`lgh up "msg"` → git push → LGH 事件 → ActionD 跑 CI → 终端显示结果

## 快速参考

| 组件 | 源码路径 | 二进制 | 端口 |
|------|----------|--------|------|
| LGH | `/Users/fenge1222/neil/LocalGitHub/lgh/` | `bin/lgh` | 9418 |
| ActionD | `/Users/fenge1222/neil/LocalGitHub/ActionD/` | `~/.local/bin/actiond` | 3000 |
| Web | `/Users/fenge1222/neil/LocalGitHub/actiond-web/` | 静态导出到 `~/.localgithub/actiond-web/out/` | (由 ActionD 服务) |

数据目录：`~/.localgithub/`

| 路径 | 内容 |
|------|------|
| `~/.localgithub/repos/` | 裸仓库 |
| `~/.localgithub/events/events.jsonl` | 事件日志（JSONL） |
| `~/.localgithub/actions/actiond.db` | ActionD SQLite 数据库 |
| `~/.localgithub/actions/` | 构建产物 |
| `~/.localgithub/plugins/` | CI 插件（35+ 个） |
| `~/.localgithub/lgh.sock` | IPC Unix Socket |

## 架构 — 事件流

```
lgh up "msg"
  → git push → LGH backend.ServeHTTP
    → pre/post ref diff → event.Publish(GitPush) with UUID event_id
    → Bus → Broker → Unix Socket (lgh.sock)
    → ActionD SocketSource → Dispatcher (trigger→language→profile→plugin.Match)
    → Worker (串行队列, cap=50) → plugin.Run() → ActionResult
    → StatusCallback → LGH /api/repos/{repo}/commits/{sha}/status
    → lgh up 终端渲染 CI 结果表格
```

关键数据链路：LGH `event.ID` (UUID) → ActionD `ActionJob.EventID` → 通过 `GET /api/actions/by-event/{event_id}` 精确查询

## CLI 命令速查

```bash
# LGH
lgh serve -d          # 启动 LGH（自动拉起 ActionD）
lgh stop              # 停止 LGH
lgh status            # 查看状态
lgh up "msg"          # 一键提交+推送+等 CI 结果
lgh save "msg"        # 仅本地 commit，不推送
lgh add .             # 注册当前目录到 LGH
lgh list              # 列出所有仓库
lgh log               # 查看日志

# ActionD
actiond start -d      # 启动 ActionD
actiond stop          # 停止 ActionD
actiond status        # 查看状态
actiond list          # 列出插件
actiond plugins       # 管理插件配置 (enable/disable/create)
```

---

## MCP 工具 — LGH（`lgh mcp` 启动）

### lgh_status
```
描述：获取 LGH 服务器状态（运行中/已停止、PID、仓库数量）
参数：无
返回：{server_running, pid, address, repos_count, repos_dir, read_only}
用法：检查 LGH 是否在运行，启动前先调这个
```

### lgh_list
```
描述：列出 LGH 上所有注册的仓库，返回本地路径和 clone URL
参数：无
返回：[{name, source_path, bare_path, clone_url, created_at}]
用法：查看有哪些仓库已注册
```

### lgh_add
```
描述：把本地 Git 仓库注册到 LGH。创建裸仓库 + 添加 'lgh' remote
参数：
  path (必填) — 本地工作目录的绝对路径
  name (可选) — 仓库名称（默认取目录名）
返回：命令输出
用法：首次使用 lgh up 前，如果仓库未注册会自动调这个
```

### lgh_remove
```
描述：从 LGH 移除仓库
参数：
  name (必填) — 仓库名称
返回：命令输出
```

### lgh_up ⭐ 最常用
```
描述：一键备份到 LGH：自动 .gitignore → git add → git commit → git push。
     推送到 localhost LGH（不是 GitHub/GitLab）。
     如果 ActionD 在运行，会自动等待 CI 完成并在终端显示结果表格。
     返回值包含 triggered_job_ids（ActionD 触发的任务 ID 列表）。

参数：
  message (必填) — Git commit message
  path (可选) — 本地工作目录绝对路径（默认当前目录）
  force (可选) — 跳过大文件检测，强制推送

返回：{success, output, commit, event_id, triggered_job_ids, project_type}

用法：
  完成代码修改后调用，等待 CI 结果再告诉用户。
  triggered_job_ids 可用于 actiond_action_get 查看详情或 actiond_job_retry 重试。
```

### lgh_save
```
描述：仅本地保存：git add + git commit，不推送到 LGH
参数：
  message (必填) — commit message
  path (可选) — 工作目录
用法：中间保存，不触发 CI
```

### lgh_serve_start
```
描述：后台启动 LGH HTTP 服务器
参数：port (可选, 默认 9418)
```

### lgh_serve_stop
```
描述：停止 LGH HTTP 服务器
参数：无
```

### lgh_log
```
描述：查看 LGH 服务器运行日志
参数：
  limit (可选, 默认 20) — 返回条数
  level (可选) — 过滤级别 (DEBUG/INFO/WARN/ERROR)
```

### lgh_rollback
```
描述：回滚到之前的 commit。执行 git reset --hard，可选 force push 到 LGH
参数：
  path (可选) — 工作目录
  steps (可选, 默认 1) — 回滚几个 commit
  push (可选) — 是否 force push 到 LGH
用法：CI 失败需要回滚时使用
```

---

## MCP 工具 — ActionD（`actiond mcp` 启动）

### actiond_status
```
描述：获取 ActionD 服务器状态（运行状态、版本、统计信息）
参数：无
用法：检查 ActionD 是否在运行
```

### actiond_plugins_list
```
描述：列出所有 CI/CD 插件。显示插件名、触发条件 (git.push/git.tag)、支持的语言、启用状态
参数：无
用法：查看有哪些插件可用，哪些已启用
```

### actiond_actions_list
```
描述：列出最近执行的 CI/CD 任务。显示 job ID、仓库、插件、状态 (done/failed/running)、耗时
参数：
  limit (可选, 默认 20) — 返回数量
用法：查看最近的 CI 执行历史
```

### actiond_action_get
```
描述：获取单个 CI/CD 任务的详细信息，包括 commit 信息、进度、执行时间
参数：
  id (必填) — 任务 ID
用法：lgh_up 返回 triggered_job_ids 后，用这个查看具体结果
```

### actiond_plugins_reload
```
描述：热重载插件（不重启 ActionD）。扫描插件目录的 manifest.json 并更新注册
参数：无
用法：添加或修改插件后刷新
```

### actiond_log
```
描述：查看 ActionD 服务器运行日志
参数：limit (可选, 默认 20)
```

### actiond_server_start / stop / restart
```
描述：启动/停止/重启 ActionD 服务器（需要环境变量 ACTIOND_MCP_ALLOW_LIFECYCLE=1）
参数：force (可选) — 有任务运行时强制操作
```

### actiond_job_cancel
```
描述：取消正在运行的 CI/CD 任务（只能取消 pending/running 状态的）
参数：id (必填) — 任务 ID
```

### actiond_job_retry
```
描述：重试失败或已取消的 CI/CD 任务
参数：id (必填) — 任务 ID
```

### actiond_plugin_enable / disable
```
描述：启用/禁用指定插件。禁用后即使事件条件满足也不会触发
参数：name (必填) — 插件名（如 'go-lint', 'security_scan'）
用法：按项目需求选择性启用插件
```

### actiond_plugins_recommend
```
描述：智能推荐插件配置。分析项目特征（语言、类型）推荐应该启用/禁用哪些插件
参数：path (可选) — 项目路径
返回：推荐列表 + 推理依据
```

### actiond_profile_get / set
```
描述：获取/设置 CI 执行 profile
Profile 控制每次 push 跑多少插件：
  fast    — 最小 CI，只跑核心 lint + test（2-3 个任务，开发时推荐）
  full    — 完整 CI，加安全扫描、覆盖率、格式化（6-10 个任务）
  release — 完整 CI/CD，加 build、deploy、release notes（10-15 个任务）
参数：profile (set 时必填) — "fast" / "full" / "release"
用法：开发时用 fast 快速反馈，合并前切 full
```

### actiond_diagnose
```
描述：诊断失败的 CI 任务，提供修复建议。
     分析失败日志，识别根因类别（build/test/lint/dependency/permission），
     提供具体的修复 hints 和相关文件。
     这是 AI 调试 CI 失败的首选工具。
参数：
  job_id (可选) — 指定任务 ID（不传则分析最近的失败任务）
  limit (可选, 默认 5) — 最多分析几个失败任务
返回：{failures: [{job_id, plugin, category, type, summary, hints, files}]}
用法：CI 失败时调用，把 hints 告诉用户
```

### dev_cycle_run ⭐ 端到端
```
描述：端到端开发循环：提交代码 → 触发 CI → 等待结果 → 返回汇总。
     聚合工具，内部自动完成 lgh up + 等待 + 收集结果。
     使用 event_id 精确追踪 CI 任务（不再按 repo 名模糊匹配）。
参数：
  message (必填) — commit message
  path (可选) — 仓库路径（默认当前目录）
  timeout (可选, 默认 300) — 等待超时秒数
  auto_rollback (可选) — 失败时自动回滚（默认 false）
  profile (可选) — 切换执行 profile (fast/full/release)
返回：{success, commit, jobs: [{plugin, status, summary, duration_ms}]}
用法：AI 修改代码后一键完成提交、测试、验证
```

### actiond_job_wait
```
描述：等待指定 CI/CD 任务完成，返回最终状态
参数：
  id (必填) — 任务 ID
  timeout (可选, 默认 300) — 超时秒数
用法：lgh_up 返回 triggered_job_ids 后等待结果
```

### actiond_cancel
```
描述：取消正在运行的 CI/CD 任务
参数：id (必填) — 任务 ID
```

---

## REST API（ActionD :3000）

| 端点 | 方法 | 用途 |
|------|------|------|
| `/api/actions` | GET | 列出最近任务 |
| `/api/actions/{id}` | GET | 任务详情 |
| `/api/actions/{id}/stream` | GET | SSE 实时日志流 |
| `/api/actions/by-event/{event_id}` | GET | 按 LGH event_id 查询任务（P1 新增） |
| `/api/actions/{id}/retry` | POST | 重试失败任务 |
| `/api/actions/{id}/cancel` | POST | 取消运行中任务 |
| `/api/actions/{id}/approve` | POST | 审批待批准任务 |
| `/api/plugins` | GET | 列出插件 |
| `/api/plugins/{name}/toggle` | POST | 启用/禁用插件 |
| `/api/profile` | GET/POST | 获取/设置执行 profile |

## 插件系统

### 插件发现（3 层优先级）
1. 内置 Go 插件（echo, deepwiki）
2. 目录扫描 `~/.localgithub/plugins/*/manifest.json`
3. 配置文件 `~/.localgithub/actions/config.json`

### manifest.json 格式
```json
{
  "apiVersion": "actiond.dev/v1",
  "name": "go-lint",
  "command": "python3",
  "args": ["run.py"],
  "triggers": ["git.push"],
  "languages": ["go"],
  "timeout": "5m",
  "refFilter": "refs/heads/*",
  "repoFilter": "*.git",
  "supported_profiles": ["fast", "full"]
}
```

### 插件 I/O 协议
- **输入** (stdin JSON): `{event, repo_path, artifact_dir}`
- **输出** (stdout JSON): `{status: "success"|"failure", summary, hints, artifacts}`

### 已安装插件（35+）
go-lint, go-test-fast, go-build, python-pytest, python-ruff, python-build,
java-quicktest, java-checkstyle, web-lint, web-test, web-build,
security_scan, coverage_report, benchmark, env-check, affected_scope,
deepwiki-action, formatter, policy_gate, approval-gate, deploy,
release-note, dependency-graph, container-package, integration-test,
observability-export, artifact_manifest, town-runtime-test ...

---

## 典型工作流

### 1. 日常开发（AI 改代码后验证）
```
1. 修改代码
2. lgh_up(message="fix: xxx")  →  等 CI 结果
3. 如果全部 ✅ → 告诉用户成功
4. 如果有 ❌ → actiond_diagnose(job_id=...) → 给用户修复建议
```

### 2. 快速迭代（只跑核心检查）
```
1. actiond_profile_set(profile="fast")
2. lgh_up(message="wip: xxx")
3. 只跑 lint + test，2-3 个任务，秒级反馈
```

### 3. 发布前全量检查
```
1. actiond_profile_set(profile="full")
2. lgh_up(message="release: xxx")
3. 跑全量 CI（安全扫描、覆盖率、格式化等）
```

### 4. CI 失败排查
```
1. actiond_actions_list(limit=5) → 找到失败的 job_id
2. actiond_diagnose(job_id="xxx") → 获取根因 + hints
3. 修复代码 → lgh_up(message="fix: xxx") → 验证
```

---

## 开发指南

### 验证栈状态
运行 `scripts/verify-stack.sh` 一键检查 LGH+ActionD 全链路健康状态。

### 构建 & 安装
```bash
# LGH
cd /Users/fenge1222/neil/LocalGitHub/lgh
go build -o lgh ./cmd/lgh/
cp lgh bin/lgh    # ⚠️ 必须同步！daemon 用 os.Executable() 可能解析到 bin/lgh
go test ./...

# ActionD
cd /Users/fenge1222/neil/LocalGitHub/ActionD
go build -o actiond ./cmd/actiond/
cp actiond ~/.local/bin/actiond
go test ./...
```

### P1 改动文件清单
| 文件 | 改动 |
|------|------|
| `lgh/cmd/lgh/serve.go` | `tryStartActionD()` — OnReady 回调自动拉起 ActionD |
| `lgh/internal/server/server.go` | `SetOnReady()` 回调，IPC socket 创建后触发 |
| `lgh/cmd/lgh/up.go` | `waitAndShowCIResults()` — event_id 轮询 + 终端 CI 结果表 |
| `lgh/internal/mcp/tools.go` | `findEventIDForCommit()` + `pollActionDByEventID()` 替代 sleep |
| `ActionD/internal/store/memory.go` | Store 接口加 `ListJobsByEventID` |
| `ActionD/internal/store/sqlite.go` | SQLite 加 `ListJobsByEventID` SQL |
| `ActionD/internal/server/server.go` | `GET /api/actions/by-event/{event_id}` 端点 |

---

## 踩坑记录

### 1. macOS 二进制签名 (SIGKILL/137) ⚠️
`go build` 产出的二进制 copy 到 `/usr/local/bin/` 会被 macOS kill。
**解决**：在项目目录内运行 `./lgh`，或用 `go install`，或 `codesign --force --sign -`

### 2. Daemon 子进程二进制路径 ⚠️
`os.Executable()` 返回实际运行的二进制路径。如果项目有 `bin/lgh`（旧版），daemon 子进程会跑旧版。
**解决**：build 后 `cp lgh bin/lgh` 同步两个位置

### 3. Socket 竞争
ActionD 启动时连接 LGH 的 Unix Socket，如果 socket 还没创建就失败（不重试）。
**解决**：用 `server.SetOnReady()` 回调，在 `startIPC()` 之后触发

### 4. bufio.Scanner 大文件截断
默认 64KB buffer，event log 超过会静默截断。
**解决**：用 `bufio.Reader.ReadString('\n')` 代替 Scanner

### 5. event_id 追踪（已修复）
旧方案：`time.Sleep(3s)` + 子串匹配 → 经常匹配不上
新方案：读 event log 找 event_id → poll `/api/actions/by-event/{event_id}` → 精确匹配

### 6. daemon PATH 找不到 actiond ⚠️
`exec.LookPath("actiond")` 在 daemon 子进程中可能找不到 `~/.local/bin/actiond`
（daemon 的 PATH 不含 `~/.local/bin`）。
**解决**：LookPath 失败后 fallback 到 `filepath.Join(home, ".local", "bin", "actiond")`。
参考 `serve.go:tryStartActionD()`。

### 7. 文档与实际不符 ⚠️
写文档时不能只看代码——必须验证：
- CLI 命令：跑 `<tool> --help` 确认实际子命令名（如 `actiond list` ≠ `actiond plugins`）
- API 端点：实际 curl 测试（如 ActionD 没有 `/health`，返回 404）
- 版本号：检查 CLI 版本（`cmd/*/main.go`）和 MCP 版本（`internal/mcp/server.go`）是否一致

### 9. 二进制候选验证（不要只靠 os.Stat）⚠️
macOS 会 SIGKILL 未签名的二进制（exit 137），`os.Stat` 返回成功但执行必死。
**解决**：验证候选二进制时用实际执行测试（如 `actiond version` 退出 0），
非零退出的候选直接跳过。`tryStartActionD()` 按 PATH → `~/.local/bin` → repo binary
顺序逐个验证，第一个 `version` 退出 0 的才用。失败写 stderr 而非静默吞掉。

### 10. 合入前质量门禁
每次改动后必须跑：
```bash
gofmt -w <changed_files>   # 格式化
go build ./...              # 编译
go test ./...               # 测试
```
三者全绿才算"可合入"。Joe 的 review 标准：找具体行号、分 P1/P2/P3 优先级、自己验证修复。

### 8. event_id 代码路径分散
event_id 查找逻辑现在存在于 3 个文件：
- `lgh/cmd/lgh/up.go` → `findEventIDFromLog()`
- `lgh/internal/mcp/tools.go` → `findEventIDForCommit()`
- `ActionD/internal/mcp/workflow.go` → `findEventIDFromLog()`
三处逻辑相同但各自实现。如果修改 event log 格式，三处都要改。
**建议**：未来提取为共享包或统一走 HTTP API。

## Codex 集成

Skill 已 symlink 到 `~/.codex/skills/lgh-actiond`，Codex 可自动触发。
