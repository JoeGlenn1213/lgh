Always respond in Chinese-simplified

# LGH Agent Protocol

This repository is the LGH codebase for Town 2.0.

## Town Role

- `project_id = LGH`
- LGH is the local git and LAN sync infrastructure layer in Town 2.0.
- In the Town stack, LGH is part of the control-center candidate together with ActionD.

## First Read

> Town 2.0 外部文档，默认位于 `~/joe/town-v2`（相对本文件即 `../../../joe/town-v2`）。
> 路径随机器布局调整；找不到时跳过即可，不影响本仓开发。

- [`../../../joe/town-v2/README.md`](../../../joe/town-v2/README.md)
- [`../../../joe/town-v2/docs/townV2-rfc-0001.md`](../../../joe/town-v2/docs/townV2-rfc-0001.md)
- [`../../../joe/town-v2/docs/townV2-rfc-0002-multi-window-governance.md`](../../../joe/town-v2/docs/townV2-rfc-0002-multi-window-governance.md)

## Default Shared Scope

- `resident_id = codex`
- `app_id = codex`
- `user_id = joe`

## Multi-Window Rules

- Treat this window as one project cell unless explicitly acting as the main window.
- Before substantial work, recall shared memory for `LGH`, `control center`, and the current task.
- Use explicit `window_id` and `task_id` in notes or handoff when continuity matters.
- Do not assume another window knows local findings unless they are written to RMS or committed to artifacts.
- Before ending substantial work, write a concise handoff covering changes, state, pending work, and main risk.

## Repository Focus

- Preserve LGH as a stable code-state and LAN-sync layer.
- Prefer explicit remotes and auditable git flows over hidden automation.
- Avoid destructive git operations unless explicitly requested.

## 相关文档（跨仓/跨框架）

- Claude Code 入口与项目级 MCP 速查：父目录 `../CLAUDE.md`
- **最完整的操作手册**：`docs/skills/lgh-actiond/SKILL.md`（每个 MCP 工具的参数/返回、踩坑记录）
- 系统边界（LGH ↔ ActionD 职责划分）：`../docs/BOUNDARY.md`
- 愿景与路线图：`../ROADMAP.md`
- 兄弟仓：`../ActionD/`（本地 CI/CD 引擎，消费 LGH 的 git 事件）

## 开发速查

- **语言/版本**：Go 1.23（见 `go.mod`）
- **入口**：`cmd/lgh/main.go`（cobra；每个子命令一个文件：`serve.go`/`up.go`/`save.go`/`mcp.go`/...）
- **构建**：`make build`（产物 `dist/lgh`，`CGO_ENABLED=0`）
- **测试**：`make test`（`go test ./... -v -cover`）；跳过集成测试用 `make test-short`
- **Lint / 格式 / 安全**：`make lint`（golangci-lint）/ `make fmt` / `make security`（go vet）
- **包地图**：
  - `internal/server` HTTP 服务（git-http-backend CGI 包装）、`internal/git` git 操作、`internal/registry` 仓库登记
  - `internal/event` 事件总线（推给 ActionD）、`internal/mcp` MCP 服务端（10 工具）、`internal/{tunnel,mdns,ignore,config,slog}`
  - `pkg/skill` 公共 Skill SDK、`pkg/ui` 终端 UI
- **运行时前置（重要）**：
  - LGH 用 `os.Executable()` 解析自身路径——**改代码后须把新二进制覆盖到实际运行路径（如 `cp dist/lgh bin/lgh` 或安装位置），否则跑的是旧二进制**。
  - `lgh serve -d` 须先启动，`git push` 到 `http://127.0.0.1:9418/...` 才会通。
  - 与 ActionD 联动依赖兄弟目录 `../ActionD/` 存在且其守护进程在跑。
- **当前版本**：`cmd/lgh/main.go` 的 `Version`（Makefile 从此读取，单一事实源）。
