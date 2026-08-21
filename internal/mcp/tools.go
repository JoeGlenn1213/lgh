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

package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/event"
	"github.com/JoeGlenn1213/lgh/internal/git"
	"github.com/JoeGlenn1213/lgh/internal/ignore"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/internal/server"
	"github.com/JoeGlenn1213/lgh/internal/slog"
)

// getArgsMap extracts arguments as a map from request
func getArgsMap(request mcp.CallToolRequest) map[string]interface{} {
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		return args
	}
	return make(map[string]interface{})
}

// getString gets a string argument
func getString(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

// getBool gets a boolean argument
func getBool(args map[string]interface{}, key string) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return false
}

// getInt gets an int argument
func getInt(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return defaultVal
}

// withInteger narrows a WithNumber-declared property to JSON Schema
// "integer" so models emit 3 instead of 3.0. mcp-go v0.43.2 has no
// WithInteger; replace this helper when the library gains one.
func withInteger(name string) mcp.ToolOption {
	return func(t *mcp.Tool) {
		if prop, ok := t.InputSchema.Properties[name].(map[string]any); ok {
			prop["type"] = "integer"
		}
	}
}

// Tool Handlers

func handleStatus(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	running, pid := server.IsRunning()
	cfg := config.Get()
	reg := registry.New()
	repos, _ := reg.List()

	// Check ActionD status via its PID file
	actiondRunning := false
	actiondPID := 0
	home, homeErr := os.UserHomeDir()
	if homeErr == nil {
		pidFile := filepath.Join(home, ".localgithub", "actions", "actiond.pid")
		if data, err := os.ReadFile(pidFile); err == nil {
			var apid int
			if _, parseErr := fmt.Sscanf(string(data), "%d", &apid); parseErr == nil && apid > 0 {
				// Use os.FindProcess + Signal(0) — pure Go, no shell-out
				if p, pErr := os.FindProcess(apid); pErr == nil {
					if p.Signal(syscall.Signal(0)) == nil {
						actiondRunning = true
						actiondPID = apid
					}
				}
			}
		}
	}

	result := map[string]interface{}{
		"lgh": map[string]interface{}{
			"running":     running,
			"pid":         pid,
			"address":     fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port),
			"repos_count": len(repos),
			"repos_dir":   cfg.ReposDir,
			"read_only":   cfg.ReadOnly,
		},
		"actiond": map[string]interface{}{
			"running": actiondRunning,
			"pid":     actiondPID,
		},
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleList(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	filter := strings.ToLower(getString(params, "filter"))

	reg := registry.New()
	repos, err := reg.List()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list repositories: %v", err)), nil
	}

	cfg := config.Get()
	var repoList []map[string]interface{}
	for _, repo := range repos {
		// Apply optional filter (matches against name or source_path)
		if filter != "" {
			nameMatch := strings.Contains(strings.ToLower(repo.Name), filter)
			pathMatch := strings.Contains(strings.ToLower(repo.SourcePath), filter)
			if !nameMatch && !pathMatch {
				continue
			}
		}
		repoList = append(repoList, map[string]interface{}{
			"name":        repo.Name,
			"source_path": repo.SourcePath,
			"bare_path":   repo.BarePath,
			"clone_url":   fmt.Sprintf("http://%s:%d/lgh/%s.git", cfg.BindAddress, cfg.Port, repo.Name),
			"created_at":  repo.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	if repoList == nil {
		repoList = []map[string]interface{}{}
	}

	data, _ := json.MarshalIndent(repoList, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleAdd(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	path := getString(params, "path")
	name := getString(params, "name")

	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	// Expand path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
	}

	// Check if path exists
	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return mcp.NewToolResultError(fmt.Sprintf("Path does not exist: %s", absPath)), nil
	}

	if name == "" {
		name = filepath.Base(absPath)
	}

	bareRepoName := name
	if filepath.Ext(bareRepoName) != ".git" {
		bareRepoName = name + ".git"
	}

	reg := registry.New()
	if reg.Exists(name) {
		return mcp.NewToolResultError(fmt.Sprintf("repository '%s' is already registered", name)), nil
	}

	if !git.IsGitRepo(absPath) {
		if err := git.InitRepo(absPath); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to initialize git repository: %v", err)), nil
		}
	}

	cfg := config.Get()
	barePath := filepath.Join(cfg.ReposDir, bareRepoName)

	if _, err := os.Stat(barePath); err == nil {
		return mcp.NewToolResultError(fmt.Sprintf("bare repository already exists at %s", barePath)), nil
	}

	if err := git.InitBareRepo(barePath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create bare repository: %v", err)), nil
	}

	remoteURL := fmt.Sprintf("http://%s:%d/lgh/%s", cfg.BindAddress, cfg.Port, bareRepoName)
	_ = git.AddRemote(absPath, "lgh", remoteURL)

	if err := reg.Add(name, absPath, barePath); err != nil {
		_ = os.RemoveAll(barePath)
		return mcp.NewToolResultError(fmt.Sprintf("failed to register repository: %v", err)), nil
	}

	event.Publish(event.RepoAdded, name, map[string]interface{}{
		"source": absPath,
		"bare":   barePath,
		"url":    remoteURL,
	})

	return mcp.NewToolResultText(fmt.Sprintf("Repository '%s' added successfully. Clone URL: %s", name, remoteURL)), nil
}

func handleRemove(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	name := getString(params, "name")

	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	reg := registry.New()
	repo, err := reg.Find(name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("repository '%s' not found", name)), nil
	}

	if _, err := os.Stat(repo.SourcePath); err == nil {
		_ = git.RemoveRemote(repo.SourcePath, "lgh")
	}

	if git.IsBareRepo(repo.BarePath) {
		_ = os.RemoveAll(repo.BarePath)
	}

	if err := reg.Remove(name); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to remove from registry: %v", err)), nil
	}

	event.Publish(event.RepoRemoved, name, map[string]interface{}{
		"source": repo.SourcePath,
		"bare":   repo.BarePath,
	})

	return mcp.NewToolResultText(fmt.Sprintf("Repository '%s' removed successfully", name)), nil
}

func handleUp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	message := getString(params, "message")
	path := getString(params, "path")
	force := getBool(params, "force")

	if message == "" {
		return mcp.NewToolResultError("message is required"), nil
	}

	// Default to current directory
	workDir := path
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get current directory: %v", err)), nil
		}
	}

	// Ensure .gitignore exists
	projectType, _ := ignore.EnsureGitignore(workDir)

	if !git.IsGitRepo(workDir) {
		if err := git.InitRepo(workDir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to initialize git repository: %v", err)), nil
		}
	}

	reg := registry.New()
	if _, err := reg.FindBySourcePath(workDir); err != nil {
		name := filepath.Base(workDir)
		bareRepoName := name + ".git"
		cfg := config.Get()
		barePath := filepath.Join(cfg.ReposDir, bareRepoName)
		_ = git.InitBareRepo(barePath)
		remoteURL := fmt.Sprintf("http://%s:%d/lgh/%s", cfg.BindAddress, cfg.Port, bareRepoName)
		_ = git.AddRemote(workDir, "lgh", remoteURL)
		_ = reg.Add(name, workDir, barePath)
		event.Publish(event.RepoAdded, name, map[string]interface{}{
			"source": workDir,
			"bare":   barePath,
			"url":    remoteURL,
		})
	}

	if !force {
		report, err := ignore.DetectTrash(workDir)
		if err == nil && report != nil && report.HasBlocking {
			return mcp.NewToolResultError("blocking issues found by trash detection. Use force: true to override."), nil
		}
	}

	output := ""
	err := git.CommitChanges(workDir, message)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			output += "Nothing to commit. "
			err = nil
		}
	} else {
		output += "Committed. "
	}

	if err == nil {
		if pushErr := git.PushToRemoteUpstream(workDir, "lgh", "HEAD"); pushErr != nil {
			err = pushErr
			output += "Push failed: " + pushErr.Error()
		} else {
			output += "Pushed to LGH."
		}
	}

	result := map[string]interface{}{
		"success":      err == nil,
		"output":       output,
		"project_type": string(projectType),
	}

	// Extract commit hash and job IDs if possible (ActionD integration)
	var commitHash string
	if err == nil {
		cmdHash := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
		cmdHash.Dir = workDir
		if hashOut, hashErr := cmdHash.Output(); hashErr == nil {
			commitHash = strings.TrimSpace(string(hashOut))
			result["commit"] = commitHash
		}

		// Query ActionD via event_id for precise job matching (no more sleep+guess)
		// LGH events carry a UUID that ActionD stores as event_id on each job.
		// We extract the event_id from the LGH event log for this commit.
		if commitHash != "" {
			eventID := findEventIDForCommit(commitHash, workDir)
			if eventID != "" {
				result["event_id"] = eventID
				// Poll ActionD by event_id — much more reliable than sleep+substring
				triggeredJobIDs := pollActionDByEventID(eventID, 10*time.Second)
				if len(triggeredJobIDs) > 0 {
					result["triggered_job_ids"] = triggeredJobIDs
				}
			}
		}

		// Keep the hint for backward compatibility
		result["triggered_jobs_hint"] = "Jobs may have been triggered in ActionD. Use dev_cycle_run instead for full tracing."
	}

	if err != nil {
		result["error"] = err.Error()
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleUpDryRun previews what lgh_up would do without actually committing or pushing.
// It shows: pending changed files, trash detection results, and registration status.
func handleUpDryRun(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	path := getString(params, "path")

	workDir := path
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get current directory: %v", err)), nil
		}
	}

	result := map[string]interface{}{
		"path":     workDir,
		"dry_run":  true,
	}

	// Check git repo state
	isGit := git.IsGitRepo(workDir)
	result["is_git_repo"] = isGit

	// Check registry
	reg := registry.New()
	repoMapping, regErr := reg.FindBySourcePath(workDir)
	result["is_registered"] = regErr == nil
	if regErr == nil {
		result["registered_name"] = repoMapping.Name
		result["clone_url"] = fmt.Sprintf("http://%s:%d/lgh/%s.git",
			config.Get().BindAddress, config.Get().Port, repoMapping.Name)
	} else {
		result["would_auto_register"] = true
		result["would_register_as"] = filepath.Base(workDir)
	}

	// Collect staged/unstaged changes via git status --short
	if isGit {
		cmd := exec.Command("git", "-C", workDir, "status", "--short")
		statusOut, err := cmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(statusOut)), "\n")
			var changedFiles []string
			for _, l := range lines {
				if l != "" {
					changedFiles = append(changedFiles, l)
				}
			}
			result["pending_changes"] = changedFiles
			result["pending_count"] = len(changedFiles)
		}
	}

	// Trash detection
	trashReport, trashErr := ignore.DetectTrash(workDir)
	if trashErr == nil && trashReport != nil {
		trashItems := []map[string]interface{}{}
		for _, item := range trashReport.Items {
			trashItems = append(trashItems, map[string]interface{}{
				"path":     item.Path,
				"size":     item.Size,
				"blocking": item.Blocking,
				"message":  item.Message,
			})
		}
		result["trash"] = map[string]interface{}{
			"has_blocking": trashReport.HasBlocking,
			"total_size":   trashReport.TotalSize,
			"items":        trashItems,
		}
		if trashReport.HasBlocking {
			result["would_fail"] = true
			result["fail_reason"] = "blocking trash detected; use force:true to override"
		}
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleSave(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	params := getArgsMap(request)
	message := getString(params, "message")
	path := getString(params, "path")

	if message == "" {
		return mcp.NewToolResultError("message is required"), nil
	}

	workDir := path
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get current directory: %v", err)), nil
		}
	}

	ignore.EnsureGitignore(workDir)
	if !git.IsGitRepo(workDir) {
		if err := git.InitRepo(workDir); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to initialize git repository: %v", err)), nil
		}
	}

	report, err := ignore.DetectTrash(workDir)
	if err == nil && report != nil && report.HasBlocking {
		return mcp.NewToolResultError("blocking issues found by trash detection"), nil
	}

	output := ""
	err = git.CommitChanges(workDir, message)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			output = "Nothing to commit, working tree clean"
			err = nil
		}
	} else {
		output = "Saved locally"
	}

	result := map[string]interface{}{
		"success": err == nil,
		"output":  output,
	}
	if err != nil {
		result["error"] = err.Error()
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleServeStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	port := getInt(params, "port", 0)

	args := []string{"serve", "--daemon"}
	if port > 0 {
		args = append(args, "--port", fmt.Sprintf("%d", port))
	}

	cmd, err := getLGHCmd(ctx, args...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve lgh binary: %v", err)), nil
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to start server: %s", string(output))), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}

func handleServeStop(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	running, pid := server.IsRunning()
	if !running {
		return mcp.NewToolResultText("LGH server is not running"), nil
	}

	if err := server.StopServer(pid); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to stop server: %v", err)), nil
	}

	return mcp.NewToolResultText("LGH server stopped successfully"), nil
}

func handleRollback(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	path := getString(params, "path")
	steps := getInt(params, "steps", 1)
	push := getBool(params, "push")

	if steps <= 0 {
		steps = 1
	}

	// Default to current directory
	workDir := path
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get current directory: %v", err)), nil
		}
	}

	// Get current commit before rollback
	cmd := exec.CommandContext(ctx, "git", "-C", workDir, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get current commit: %v", err)), nil
	}
	beforeCommit := strings.TrimSpace(string(output))

	// Get target commit (N steps back)
	cmd = exec.CommandContext(ctx, "git", "-C", workDir, "rev-parse", fmt.Sprintf("HEAD~%d", steps))
	output, err = cmd.Output()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to find commit %d steps back: %v", steps, err)), nil
	}
	targetCommit := strings.TrimSpace(string(output))

	// Get commit message for info
	cmd = exec.CommandContext(ctx, "git", "-C", workDir, "log", "-1", "--format=%s", beforeCommit)
	msgOutput, _ := cmd.Output()
	rollbackMsg := strings.TrimSpace(string(msgOutput))

	// Perform git reset --hard
	cmd = exec.CommandContext(ctx, "git", "-C", workDir, "reset", "--hard", targetCommit)
	resetOutput, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to reset: %s", string(resetOutput))), nil
	}

	result := map[string]interface{}{
		"success":       true,
		"from_commit":   beforeCommit,
		"to_commit":     targetCommit,
		"steps":         steps,
		"rolled_back":   rollbackMsg,
		"local_changed": true,
	}

	// Optionally push to LGH (force push required)
	if push {
		cmd = exec.CommandContext(ctx, "git", "-C", workDir, "push", "lgh", "--force")
		pushOutput, pushErr := cmd.CombinedOutput()
		if pushErr != nil {
			result["push_success"] = false
			result["push_error"] = string(pushOutput)
		} else {
			result["push_success"] = true
			result["remote_changed"] = true
		}
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleLog(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	limit := getInt(params, "limit", 20)
	level := getString(params, "level")

	if limit <= 0 {
		limit = 20
	}

	cfg := config.Get()
	logPath := filepath.Join(cfg.DataDir, "logs", "server.jsonl")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return mcp.NewToolResultText("[]"), nil
	}

	lines, err := slog.ReadLastLines(logPath, limit, level)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read logs: %v", err)), nil
	}

	// Format as JSON array
	output := "[" + strings.Join(lines, ",") + "]"
	return mcp.NewToolResultText(output), nil
}

func handleDiff(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	path := getString(params, "path")
	staged := getBool(params, "staged")

	workDir := path
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get current directory: %v", err)), nil
		}
	}

	// Check if it's a git repo
	cmd := exec.CommandContext(ctx, "git", "-C", workDir, "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Not a git repository: %s", workDir)), nil
	}

	// Build diff command
	args := []string{"-C", workDir, "diff"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--stat") // Summary first

	cmd = exec.CommandContext(ctx, "git", args...)
	statOutput, _ := cmd.Output()

	// Also get the full diff (truncated for large diffs)
	fullArgs := []string{"-C", workDir, "diff"}
	if staged {
		fullArgs = append(fullArgs, "--cached")
	}
	cmd = exec.CommandContext(ctx, "git", fullArgs...)
	fullOutput, _ := cmd.Output()

	// Truncate if too large (>32KB) to avoid flooding AI context
	diff := string(fullOutput)
	truncated := false
	if len(diff) > 32*1024 {
		diff = diff[:32*1024]
		truncated = true
	}

	result := map[string]interface{}{
		"has_changes": len(statOutput) > 0,
		"stat":        string(statOutput),
		"diff":        diff,
		"truncated":   truncated,
		"staged":      staged,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// getLGHCmd returns an exec.Cmd for the current LGH binary
func getLGHCmd(ctx context.Context, args ...string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	//nolint:gosec // G204: exe is trusted (os.Executable), args are commands
	return exec.CommandContext(ctx, exe, args...), nil
}

// findEventIDForCommit reads the LGH event log to find the event_id for a given commit hash.
// Events are stored as JSONL in ~/.localgithub/events/events.jsonl
func findEventIDForCommit(commitHash, workDir string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	eventLogPath := filepath.Join(home, ".localgithub", "events", "events.jsonl")
	f, err := os.Open(eventLogPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	// Get repo name from workDir, preferring registered name if found
	repoName := filepath.Base(workDir)
	reg := registry.New()
	if mapping, regErr := reg.FindBySourcePath(workDir); regErr == nil && mapping != nil && mapping.Name != "" {
		repoName = mapping.Name
	}
	normRepo := normalizeRepoName(repoName)

	// Read from the end (most recent events first) — scan last 50 lines
	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 256*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Search from end for a matching push event
	start := len(lines) - 50
	if start < 0 {
		start = 0
	}
	for i := len(lines) - 1; i >= start; i-- {
		var evt event.Event
		if json.Unmarshal([]byte(lines[i]), &evt) != nil {
			continue
		}
		if evt.Type != event.GitPush {
			continue
		}
		// Match by repo name (case-insensitive, with/without .git, and normalized)
		cleanEvtRepo := strings.TrimSuffix(strings.ToLower(evt.RepoName), ".git")
		cleanRepoName := strings.TrimSuffix(strings.ToLower(repoName), ".git")
		if cleanEvtRepo != cleanRepoName && normalizeRepoName(evt.RepoName) != normRepo {
			continue
		}
		// Check if this event's changes contain our commit hash
		if payload, ok := evt.Payload["changes"].(map[string]interface{}); ok {
			for _, change := range payload {
				if changeMap, ok := change.(map[string]interface{}); ok {
					newHash, _ := changeMap["new"].(string)
					if strings.HasPrefix(newHash, commitHash) || strings.HasPrefix(commitHash, newHash) {
						return evt.ID
					}
				}
			}
		}
	}
	return ""
}

func normalizeRepoName(name string) string {
	name = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".git")
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// pollActionDByEventID polls ActionD's /api/actions/by-event/{event_id} endpoint
// until all jobs reach terminal state or timeout. Returns job IDs.
func pollActionDByEventID(eventID string, timeout time.Duration) []string {
	client := http.Client{Timeout: 3 * time.Second}
	deadline := time.Now().Add(timeout)
	var jobIDs []string

	for time.Now().Before(deadline) {
		resp, err := client.Get(fmt.Sprintf("http://localhost:3000/api/actions/by-event/%s", eventID))
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var jobs []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&jobs)
		resp.Body.Close()

		if len(jobs) == 0 {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Collect job IDs
		jobIDs = make([]string, 0, len(jobs))
		allTerminal := true
		for _, j := range jobs {
			if id, ok := j["id"].(string); ok {
				jobIDs = append(jobIDs, id)
			}
			status, _ := j["status"].(string)
			if status != "done" && status != "failed" && status != "cancelled" {
				allTerminal = false
			}
		}

		if allTerminal {
			return jobIDs
		}

		time.Sleep(500 * time.Millisecond)
	}

	return jobIDs
}

// Resource Handlers

func handleResourceConfig(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	cfg := config.Get()
	data, _ := json.MarshalIndent(map[string]interface{}{
		"data_dir":     cfg.DataDir,
		"repos_dir":    cfg.ReposDir,
		"bind_address": cfg.BindAddress,
		"port":         cfg.Port,
		"read_only":    cfg.ReadOnly,
		"mdns_enabled": cfg.MDNSEnabled,
		"auth_enabled": cfg.AuthEnabled,
	}, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: request.Params.URI, MIMEType: "application/json", Text: string(data)},
	}, nil
}

func handleResourceRepos(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	reg := registry.New()
	repos, _ := reg.List()

	var repoList []map[string]interface{}
	for _, repo := range repos {
		repoList = append(repoList, map[string]interface{}{
			"name":        repo.Name,
			"source_path": repo.SourcePath,
			"created_at":  repo.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	data, _ := json.MarshalIndent(repoList, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: request.Params.URI, MIMEType: "application/json", Text: string(data)},
	}, nil
}

func handleResourceServerStatus(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	running, pid := server.IsRunning()
	cfg := config.Get()

	data, _ := json.MarshalIndent(map[string]interface{}{
		"running": running,
		"pid":     pid,
		"address": fmt.Sprintf("http://%s:%d", cfg.BindAddress, cfg.Port),
	}, "", "  ")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: request.Params.URI, MIMEType: "application/json", Text: string(data)},
	}, nil
}
