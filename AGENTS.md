# LGH — Agent Guide

Guidance for AI coding agents (and humans in a hurry) working in this repo.

## Project in one line

LGH (Local Git Hub) turns a local directory into a Git server: `lgh serve` exposes an HTTP remote (`http://127.0.0.1:9418/...`) with repo registry, events, LAN sync, and an MCP server — the git/event backbone that [ActionD](https://github.com/JoeGlenn1213/ActionD) listens to.

## Dev Quickstart

- **Language/version**: Go 1.23 (see `go.mod`).
- **Entry**: `cmd/lgh/main.go` (cobra; one file per subcommand: `serve.go`/`up.go`/`save.go`/`mcp.go`/...).
- **Build**: `make build` (artifact `dist/lgh`, `CGO_ENABLED=0`).
- **Test**: `make test` (`go test ./... -v -cover`); `make test-short` skips integration tests.
- **Lint / fmt / security**: `make lint` (golangci-lint) / `make fmt` / `make security` (go vet).
- **Package map**:
  - `internal/server` HTTP service (git-http-backend CGI wrapper), `internal/git` git operations, `internal/registry` repo registry.
  - `internal/event` event bus (feeds ActionD), `internal/mcp` MCP server (10 tools), `internal/{tunnel,mdns,ignore,config,slog}`.
  - `pkg/skill` shared Skill SDK, `pkg/ui` terminal UI.
- **Runtime prerequisites (important)**:
  - LGH resolves its own path via `os.Executable()` — after code changes, copy the new binary over the running location (e.g. `cp dist/lgh bin/lgh` or the install path), otherwise you are testing the old binary.
  - **macOS gotcha (field-tested)**: overwriting a signed/running binary in place invalidates the kernel-cached signature and the next execution gets **SIGKILL (exit 137)**. Re-sign with `codesign -f -s - <binary>` after overwriting; if the target is in `/usr/local/bin` (root-owned, inode can't change without sudo) and re-signing fails, place the binary in a user dir such as `~/.local/bin/` (new inode) instead.
  - Start `lgh serve -d` first; `git push` to `http://127.0.0.1:9418/...` only works while it runs.
  - ActionD integration expects the sibling [ActionD](https://github.com/JoeGlenn1213/ActionD) checkout and its daemon running.
- **Version**: canonical version lives in `internal/version/version.go` (single source of truth, read by the Makefile).

## Conventions

- Preserve LGH as a stable code-state and LAN-sync layer.
- Prefer explicit remotes and auditable git flows over hidden automation.
- Avoid destructive git operations unless explicitly requested.
