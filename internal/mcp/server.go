// Copyright (c) 2025 JoeGlenn1213
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package mcp provides Model Context Protocol server implementation for LGH
package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/JoeGlenn1213/lgh/internal/version"
)

// NewServer creates and configures the LGH MCP server
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"lgh",
		version.Version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	// Register tools
	registerTools(s)

	// Register resources
	registerResources(s)

	return s
}

// registerTools registers all LGH tools with the MCP server
func registerTools(s *server.MCPServer) {
	// lgh_status - Get server status (read-only)
	s.AddTool(
		mcp.NewTool("lgh_status",
			mcp.WithDescription("Get a unified health snapshot of the LGH server and the ActionD daemon. Aggregates LGH running state, PID, listen address, registered repo count, repos dir, and read-only flag, plus the ActionD daemon's running state and PID (probed via its PID file). Takes no arguments and reads only local state, so it is safe to call at any time. Use this to answer \"is LGH up?\"; use lgh_log when you need the historical runtime events that led to the current state. Returns pretty-printed JSON with top-level \"lgh\" and \"actiond\" objects."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Get LGH and ActionD health status",
				ReadOnlyHint:    mcp.ToBoolPtr(true),
				DestructiveHint: mcp.ToBoolPtr(false),
				IdempotentHint:  mcp.ToBoolPtr(true),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
		),
		handleStatus,
	)

	// lgh_list - List repositories (read-only)
	s.AddTool(
		mcp.NewTool("lgh_list",
			mcp.WithDescription("List every Git repository registered with the LGH local server, each with its name, source_path (local working dir), bare_path, clone_url (LGH server URL), and created_at. Reads only the local registry under ~/.localgithub; no network access. Use it to discover the clone_url of an existing backup or the exact registered name required by lgh_remove. Returns a JSON array, empty when nothing is registered."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "List registered repositories",
				ReadOnlyHint:    mcp.ToBoolPtr(true),
				DestructiveHint: mcp.ToBoolPtr(false),
				IdempotentHint:  mcp.ToBoolPtr(true),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithString("filter",
				mcp.Description("Optional substring filter applied to repository name or source path (case-insensitive)"),
			),
		),
		handleList,
	)

	// lgh_add - Add repository (mutating, non-destructive)
	s.AddTool(
		mcp.NewTool("lgh_add",
			mcp.WithDescription("Register a local Git repository with the LGH local server. Creates a bare repository under the LGH repos dir and adds an \"lgh\" remote pointing at it to the source repo. Initializes a git repo first if the directory is not one yet; refuses with an error if the path does not exist, the name is already registered, or a bare repo already exists at the target path. Returns the repository name and its clone URL. Use lgh_up to push subsequent commits and lgh_remove to unregister."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Register a local Git repository with LGH",
				ReadOnlyHint:    mcp.ToBoolPtr(false),
				DestructiveHint: mcp.ToBoolPtr(false),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Absolute path to the LOCAL working directory containing the Git repository"),
			),
			mcp.WithString("name",
				mcp.Description("Optional custom name for the repository on LGH"),
			),
		),
		handleAdd,
	)

	// lgh_remove - Remove repository (destructive: deletes LGH-side backup)
	s.AddTool(
		mcp.NewTool("lgh_remove",
			mcp.WithDescription("Unregister a repository from LGH. Deletes the bare backup repository from the LGH repos dir and removes the \"lgh\" remote from the local working directory; the local working tree and its git history are never touched. Fails with an error if the name is not registered. Destructive: the LGH-side backup copy and its history are permanently deleted and only come back by re-registering (lgh_add) and pushing again (lgh_up). Use lgh_list first to find the exact registered name."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Unregister a repository and delete its LGH backup",
				ReadOnlyHint:    mcp.ToBoolPtr(false),
				DestructiveHint: mcp.ToBoolPtr(true),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name of the repository to remove"),
			),
		),
		handleRemove,
	)

	// lgh_clone - Clone/restore a repository (mutating: creates a new working
	// directory; never destructive, never overwrites existing files)
	s.AddTool(
		mcp.NewTool("lgh_clone",
			mcp.WithDescription("Clone a repository from the LGH local server into a new working directory on this machine — the restore side of the backup loop and the inverse of lgh_up. Use it to recover a working copy after data loss or to materialize an existing backup on a new machine: find the exact registered name with lgh_list, make sure the server is up (lgh_status, otherwise lgh_serve_start), then clone. The name may carry a .git suffix. Refuses with a clear error when the name is not registered, the server is not running, or the destination directory already exists and is not empty — it never overwrites existing files. Returns JSON with name, clone_url, destination (absolute path), and the checked-out branch and HEAD commit."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Clone (restore) a repository from LGH",
				ReadOnlyHint:    mcp.ToBoolPtr(false),
				DestructiveHint: mcp.ToBoolPtr(false),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name of the repository registered with LGH, with or without .git suffix (find exact names with lgh_list)"),
			),
			mcp.WithString("dest",
				mcp.Description("Optional destination directory (absolute path recommended; defaults to ./<name> under the MCP server's working directory)"),
			),
		),
		handleClone,
	)

	// lgh_up - One-click commit and push (mutating, non-destructive)
	s.AddTool(
		mcp.NewTool("lgh_up",
			mcp.WithDescription("One-click backup: ensures a .gitignore exists, then git add + git commit + git push to the LGH local server (localhost, NOT GitHub/GitLab). Auto-registers the repo with LGH and initializes git if needed, so it also works on a fresh directory. Aborts with an error when trash detection finds blocking issues (large files, .env secrets) unless force is set. If ActionD is running, the result includes event_id and triggered_job_ids for the CI jobs spawned by this push; use the ActionD server's dev_cycle_run tool when you need full job tracing. Returns JSON with success, output, project_type, commit, and optional triggered_job_ids."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Commit and push to LGH (one-click backup)",
				ReadOnlyHint:    mcp.ToBoolPtr(false),
				DestructiveHint: mcp.ToBoolPtr(false),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithString("message",
				mcp.Required(),
				mcp.Description("Git commit message"),
			),
			mcp.WithString("path",
				mcp.Description("Absolute path to the LOCAL working directory (defaults to current directory)"),
			),
			mcp.WithBoolean("force",
				mcp.Description("Skip trash detection (large files, .env) and force push"),
			),
		),
		handleUp,
	)

	// lgh_up_dryrun - Preview what lgh_up would do (read-only)
	s.AddTool(
		mcp.NewTool("lgh_up_dryrun",
			mcp.WithDescription("Dry-run preview of lgh_up without touching any state: nothing is committed, pushed, registered, or written. Reports whether the directory is a git repo, the pending changed files, LGH registration status (or the name it would auto-register as), and trash-detection results including whether lgh_up would fail and why. Call this before lgh_up whenever you are unsure what a backup would include; use lgh_diff when you want the actual patch content instead of the changed-file list. Returns a JSON object with dry_run=true."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Preview lgh_up without committing or pushing",
				ReadOnlyHint:    mcp.ToBoolPtr(true),
				DestructiveHint: mcp.ToBoolPtr(false),
				IdempotentHint:  mcp.ToBoolPtr(true),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithString("path",
				mcp.Description("Absolute path to the LOCAL working directory (defaults to current directory)"),
			),
		),
		handleUpDryRun,
	)

	// lgh_save - Local save (mutating, non-destructive)
	s.AddTool(
		mcp.NewTool("lgh_save",
			mcp.WithDescription("Create a local checkpoint: ensures a .gitignore exists, then git add + git commit in the working directory WITHOUT pushing anywhere — nothing is synced to the LGH server or any other remote. Initializes git if the directory is not a repo yet, and aborts when trash detection finds blocking issues. Use this for intermediate saves; switch to lgh_up when the work should also reach the LGH backup. Returns JSON with success and a human-readable output line (\"Nothing to commit, working tree clean\" when there are no changes)."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Commit locally without pushing",
				ReadOnlyHint:    mcp.ToBoolPtr(false),
				DestructiveHint: mcp.ToBoolPtr(false),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithString("message",
				mcp.Required(),
				mcp.Description("Git commit message"),
			),
			mcp.WithString("path",
				mcp.Description("Absolute path to the LOCAL working directory (defaults to current directory)"),
			),
		),
		handleSave,
	)

	// lgh_serve_start - Start server (lifecycle)
	s.AddTool(
		mcp.NewTool("lgh_serve_start",
			mcp.WithDescription("Start the LGH HTTP server in the background by spawning \"lgh serve --daemon\" from the current binary. The server hosts the git endpoints that repos push to, so start it before lgh_up or before cloning from LGH. Returns the daemon's startup output, or an error with that output if the server cannot start (for example the port is already taken). Check lgh_status afterwards to confirm the PID and listen address."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Start the LGH HTTP server",
				ReadOnlyHint:    mcp.ToBoolPtr(false),
				DestructiveHint: mcp.ToBoolPtr(false),
				IdempotentHint:  mcp.ToBoolPtr(true),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithNumber("port",
				mcp.Description("Port to listen on (default: 9418)"),
			),
			withInteger("port"),
		),
		handleServeStart,
	)

	// lgh_serve_stop - Stop server (lifecycle)
	s.AddTool(
		mcp.NewTool("lgh_serve_stop",
			mcp.WithDescription("Stop the background LGH HTTP server by shutting down its PID. Registered repositories and local commits are unaffected; pushes simply fail until the server is restarted with lgh_serve_start. Returns a friendly notice instead of an error when the server is not running, so it is safe to call speculatively as a cleanup step. Takes no arguments; use lgh_status to check state first if you want to avoid a no-op call."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Stop the LGH HTTP server",
				ReadOnlyHint:    mcp.ToBoolPtr(false),
				DestructiveHint: mcp.ToBoolPtr(false),
				IdempotentHint:  mcp.ToBoolPtr(true),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
		),
		handleServeStop,
	)

	// lgh_log - View server logs (read-only)
	s.AddTool(
		mcp.NewTool("lgh_log",
			mcp.WithDescription("Read the LGH server's runtime log from ~/.localgithub/logs/server.jsonl and return the most recent entries as a JSON array of {ts, level, msg, component} objects (an empty array if no log file exists yet). A level filter keeps only entries at exactly that severity, so level=ERROR surfaces just the failures. Use this to investigate why a push or the server misbehaved in the past: lgh_status only shows the current health snapshot, while lgh_log shows the history that led to it."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Read LGH server logs",
				ReadOnlyHint:    mcp.ToBoolPtr(true),
				DestructiveHint: mcp.ToBoolPtr(false),
				IdempotentHint:  mcp.ToBoolPtr(true),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithNumber("limit",
				mcp.Description("Number of log entries to return (default: 20)"),
			),
			withInteger("limit"),
			mcp.WithString("level",
				mcp.Description("Filter by log level (DEBUG, INFO, WARN, ERROR)"),
				mcp.Enum("DEBUG", "INFO", "WARN", "ERROR"),
			),
		),
		handleLog,
	)

	// lgh_rollback - Rollback to previous commit (destructive: git reset --hard)
	s.AddTool(
		mcp.NewTool("lgh_rollback",
			mcp.WithDescription("Roll the repository back N commits by running git reset --hard HEAD~N. Destructive: uncommitted changes and the rolled-back commits are discarded from the local branch, so run lgh_up or lgh_save first if anything there is worth keeping. With push enabled it also force-pushes the reset state to the LGH server so the backup matches; without push only the local branch moves. Returns JSON with from_commit, to_commit, rolled_back (the discarded commit's subject), and push success. Use lgh_diff or lgh_up_dryrun to inspect what would be lost before rolling back."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Roll back the repository to a previous commit",
				ReadOnlyHint:    mcp.ToBoolPtr(false),
				DestructiveHint: mcp.ToBoolPtr(true),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithString("path",
				mcp.Description("Absolute path to the LOCAL working directory (defaults to current directory)"),
			),
			mcp.WithNumber("steps",
				mcp.Description("Number of commits to roll back (default: 1)"),
			),
			withInteger("steps"),
			mcp.WithBoolean("push",
				mcp.Description("Force push to LGH after rollback (default: false)"),
			),
		),
		handleRollback,
	)

	// lgh_diff - Show uncommitted changes (read-only)
	s.AddTool(
		mcp.NewTool("lgh_diff",
			mcp.WithDescription("Show uncommitted changes in a Git working directory: runs git diff --stat for a per-file summary plus the full unified diff, and returns JSON with has_changes, stat, diff, truncated, and staged. The full diff is truncated at 32KB to protect the caller's context. Refuses with an error when the path is not inside a git work tree. Use this to review patch content before lgh_save or lgh_up; use lgh_up_dryrun instead when you only need the changed-file list, trash detection, and LGH registration state."),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:           "Show uncommitted changes in a repository",
				ReadOnlyHint:    mcp.ToBoolPtr(true),
				DestructiveHint: mcp.ToBoolPtr(false),
				IdempotentHint:  mcp.ToBoolPtr(true),
				OpenWorldHint:   mcp.ToBoolPtr(false),
			}),
			mcp.WithString("path",
				mcp.Description("Absolute path to the LOCAL working directory (defaults to current directory)"),
			),
			mcp.WithBoolean("staged",
				mcp.Description("Show only staged (cached) changes (default: false)"),
			),
		),
		handleDiff,
	)
}

// registerResources registers all LGH resources with the MCP server
func registerResources(s *server.MCPServer) {
	// lgh://config - Current configuration
	s.AddResource(
		mcp.NewResource("lgh://config",
			"LGH Configuration",
			mcp.WithResourceDescription("Current LGH configuration settings"),
			mcp.WithMIMEType("application/json"),
		),
		handleResourceConfig,
	)

	// lgh://repos - Repository list
	s.AddResource(
		mcp.NewResource("lgh://repos",
			"Repository List",
			mcp.WithResourceDescription("List of all repositories registered with LGH"),
			mcp.WithMIMEType("application/json"),
		),
		handleResourceRepos,
	)

	// lgh://server/status - Server status
	s.AddResource(
		mcp.NewResource("lgh://server/status",
			"Server Status",
			mcp.WithResourceDescription("Current LGH server status"),
			mcp.WithMIMEType("application/json"),
		),
		handleResourceServerStatus,
	)
}
