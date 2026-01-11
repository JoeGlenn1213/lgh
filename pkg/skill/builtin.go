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

package skill

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/JoeGlenn1213/lgh/internal/ignore"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/internal/server"
)

func init() {
	// Register built-in skills
	Register(&BackupSkill{})
	Register(&StatusSkill{})
	Register(&ListReposSkill{})
}

// BackupSkill - One-click backup (lgh up)
type BackupSkill struct{}

// Meta returns the metadata for the BackupSkill
func (s *BackupSkill) Meta() Metadata {
	return Metadata{
		ID:          "lgh.backup",
		Name:        "Backup Code",
		Description: "One-click backup: auto gitignore + add + commit + push to LGH",
		Category:    "git",
		InputSchema: []Param{
			{Name: "path", Type: "path", Required: true, Description: "Repository path"},
			{Name: "message", Type: "string", Required: true, Description: "Commit message"},
			{Name: "force", Type: "bool", Required: false, Default: false, Description: "Skip trash detection"},
		},
		Tags: []string{"git", "backup", "push", "commit"},
	}
}

// Execute performs the backup operation
func (s *BackupSkill) Execute(ctx context.Context, input Input) (*Result, error) {
	start := time.Now()

	path, _ := input["path"].(string)
	message, _ := input["message"].(string)
	force, _ := input["force"].(bool)

	if path == "" || message == "" {
		return &Result{
			Success:   false,
			Error:     "path and message are required",
			Duration:  time.Since(start),
			Timestamp: start,
		}, nil
	}

	// Ensure gitignore
	projectType, _ := ignore.EnsureGitignore(path)

	// Build command
	args := []string{"up", message}
	if force {
		args = append(args, "--force")
	}

	// Resolve LGH binary
	// Note: We cannot simply use os.Executable() because this package might be imported
	// by an external application (e.g. Dandelion Town). In that case, os.Executable()
	// would point to the external app, not lgh.
	exe, err := resolveLGHBinary()
	if err != nil {
		return &Result{
			Success:   false,
			Error:     err.Error(),
			Duration:  time.Since(start),
			Timestamp: start,
		}, nil
	}

	//nolint:gosec // G204: exe resolved internally, args constructed from input
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()

	result := &Result{
		Success:   err == nil,
		Duration:  time.Since(start),
		Timestamp: start,
		Output: Output{
			"raw_output":   string(output),
			"project_type": string(projectType),
		},
	}
	if err != nil {
		result.Error = err.Error()
	}

	return result, nil
}

// StatusSkill - Get LGH status
type StatusSkill struct{}

// Meta returns the metadata for the StatusSkill
func (s *StatusSkill) Meta() Metadata {
	return Metadata{
		ID:          "lgh.status",
		Name:        "Server Status",
		Description: "Check if LGH server is running and get details",
		Category:    "server",
		InputSchema: []Param{},
		Tags:        []string{"status", "health", "server"},
	}
}

// Execute returns the status of LGH
func (s *StatusSkill) Execute(_ context.Context, _ Input) (*Result, error) {
	start := time.Now()

	running, pid := server.IsRunning()
	reg := registry.New()
	repos, _ := reg.List()

	return &Result{
		Success:   true,
		Duration:  time.Since(start),
		Timestamp: start,
		Output: Output{
			"server_running": running,
			"pid":            pid,
			"repos_count":    len(repos),
		},
	}, nil
}

// ListReposSkill - List repositories
type ListReposSkill struct{}

// Meta returns the metadata for the ListReposSkill
func (s *ListReposSkill) Meta() Metadata {
	return Metadata{
		ID:          "lgh.list",
		Name:        "List Repositories",
		Description: "List all repositories registered with LGH",
		Category:    "repo",
		InputSchema: []Param{},
		Tags:        []string{"list", "repos", "repositories"},
	}
}

// Execute lists all repositories in LGH
func (s *ListReposSkill) Execute(_ context.Context, _ Input) (*Result, error) {
	start := time.Now()

	reg := registry.New()
	repos, err := reg.List()
	if err != nil {
		return &Result{
			Success:   false,
			Error:     err.Error(),
			Duration:  time.Since(start),
			Timestamp: start,
		}, nil
	}

	var repoList []map[string]interface{}
	for _, repo := range repos {
		repoList = append(repoList, map[string]interface{}{
			"name":        repo.Name,
			"source_path": repo.SourcePath,
			"bare_path":   repo.BarePath,
			"created_at":  repo.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &Result{
		Success:   true,
		Duration:  time.Since(start),
		Timestamp: start,
		Output: Output{
			"count": len(repoList),
			"repos": repoList,
		},
	}, nil
}

// RunSkill executes a skill by ID. It is a convenience function.
// RunSkill ...
func RunSkill(ctx context.Context, id string, input Input) (*Result, error) {
	s := Get(id)
	if s == nil {
		return nil, fmt.Errorf("skill not found: %s", id)
	}
	return s.Execute(ctx, input)
}

// GetString safely extracts a string from Input
func (i Input) GetString(key string) string {
	if v, ok := i[key].(string); ok {
		return v
	}
	return ""
}

// GetBool safely extracts a bool from Input
func (i Input) GetBool(key string) bool {
	if v, ok := i[key].(bool); ok {
		return v
	}
	return false
}

// GetPath safely extracts and resolves a path from Input
func (i Input) GetPath(key string) string {
	path := i.GetString(key)
	if path == "" {
		return ""
	}
	// Expand ~ and resolve absolute
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

// resolveLGHBinary attempts to find the correct lgh binary path
// Strategy:
// 1. Check LGH_BINARY env var
// 2. Check if current executable is "lgh" (self-call optimization)
// 3. Fallback to PATH lookup
func resolveLGHBinary() (string, error) {
	// 1. Explicit override
	if envPath := os.Getenv("LGH_BINARY"); envPath != "" {
		return envPath, nil
	}

	// 2. Self-check
	selfPath, err := os.Executable()
	if err == nil {
		base := filepath.Base(selfPath)
		// Handle .exe for Windows if needed, though usually just suffix check
		if base == "lgh" || base == "lgh.exe" {
			return selfPath, nil
		}
	}

	// 3. PATH lookup
	path, err := exec.LookPath("lgh")
	if err != nil {
		return "", fmt.Errorf("lgh binary not found in PATH (and not running as lgh): %v", err)
	}
	return path, nil
}
