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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/JoeGlenn1213/lgh/internal/config"
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

// getFloat gets a float argument
func getFloat(args map[string]interface{}, key string) float64 {
	if v, ok := args[key].(float64); ok {
		return v
	}
	return 0
}

// Tool Handlers

func handleStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	running, pid := server.IsRunning()
	cfg := config.Get()
	reg := registry.New()
	repos, _ := reg.List()

	result := map[string]interface{}{
		"server_running": running,
		"pid":            pid,
		"address":        fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port),
		"repos_count":    len(repos),
		"repos_dir":      cfg.ReposDir,
		"read_only":      cfg.ReadOnly,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	reg := registry.New()
	repos, err := reg.List()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list repositories: %v", err)), nil
	}

	cfg := config.Get()
	var repoList []map[string]interface{}
	for _, repo := range repos {
		repoList = append(repoList, map[string]interface{}{
			"name":        repo.Name,
			"source_path": repo.SourcePath,
			"bare_path":   repo.BarePath,
			"clone_url":   fmt.Sprintf("http://%s:%d/lgh/%s.git", cfg.BindAddress, cfg.Port, repo.Name),
			"created_at":  repo.CreatedAt.Format("2006-01-02 15:04:05"),
		})
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
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Path does not exist: %s", absPath)), nil
	}

	// Use lgh add command
	cmdArgs := []string{"add", absPath}
	if name != "" {
		cmdArgs = append(cmdArgs, "--name", name)
	}

	cmd, err := getLGHCmd(ctx, cmdArgs...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve lgh binary: %v", err)), nil
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add repository: %s", string(output))), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}

func handleRemove(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	name := getString(params, "name")

	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	// Use lgh remove command with -y flag
	cmd, err := getLGHCmd(ctx, "remove", name, "-y")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve lgh binary: %v", err)), nil
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove repository: %s", string(output))), nil
	}

	return mcp.NewToolResultText(string(output)), nil
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

	// Build and run the command
	cmdArgs := []string{"up", message}
	if force {
		cmdArgs = append(cmdArgs, "--force")
	}

	cmd, err := getLGHCmd(ctx, cmdArgs...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve lgh binary: %v", err)), nil
	}
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"success":      err == nil,
		"output":       string(output),
		"project_type": string(projectType),
	}
	if err != nil {
		result["error"] = err.Error()
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

	cmd, err := getLGHCmd(ctx, "save", message)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve lgh binary: %v", err)), nil
	}
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"success": err == nil,
		"output":  string(output),
	}
	if err != nil {
		result["error"] = err.Error()
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleServeStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	port := getFloat(params, "port")

	args := []string{"serve", "--daemon"}
	if port > 0 {
		args = append(args, "--port", fmt.Sprintf("%d", int(port)))
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

func handleServeStop(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmd, err := getLGHCmd(ctx, "stop")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve lgh binary: %v", err)), nil
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to stop server: %s", string(output))), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}

func handleLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := getArgsMap(request)
	limit := getFloat(params, "limit")
	level := getString(params, "level")

	if limit <= 0 {
		limit = 20
	}

	cfg := config.Get()
	logPath := filepath.Join(cfg.DataDir, "logs", "server.jsonl")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return mcp.NewToolResultText("[]"), nil
	}

	lines, err := slog.ReadLastLines(logPath, int(limit), level)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read logs: %v", err)), nil
	}

	// Format as JSON array
	output := "[" + strings.Join(lines, ",") + "]"
	return mcp.NewToolResultText(output), nil
}

// getLGHCmd returns an exec.Cmd for the current LGH binary
func getLGHCmd(ctx context.Context, args ...string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return exec.CommandContext(ctx, exe, args...), nil
}

// Resource Handlers

func handleResourceConfig(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
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

func handleResourceRepos(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
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

func handleResourceServerStatus(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
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
