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

package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Remote represents a git remote
type Remote struct {
	Name string
	URL  string
}

// CommitInfo represents basic commit information
type CommitInfo struct {
	Hash   string
	Author string
	Date   string
	Msg    string
}

// CheckGitInstalled checks if git is installed and returns its path
func CheckGitInstalled() (string, error) {
	path, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("git is not installed or not in PATH: %w", err)
	}
	return path, nil
}

// GetGitVersion returns the installed git version
func GetGitVersion() (string, error) {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// InitBareRepo creates a bare git repository at the specified path
func InitBareRepo(barePath string) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(barePath)
	if err := os.MkdirAll(parentDir, 0700); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Check if already exists
	if _, err := os.Stat(barePath); err == nil {
		return fmt.Errorf("bare repository already exists at %s", barePath)
	}

	// Initialize bare repository
	cmd := exec.Command("git", "init", "--bare", barePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to init bare repo: %s, %w", string(output), err)
	}

	return nil
}

// InitRepo initializes a git repository at the specified path
func InitRepo(repoPath string) error {
	// Check if already a git repo
	if IsGitRepo(repoPath) {
		return nil // Already initialized
	}

	// Initialize repository
	cmd := exec.Command("git", "init", repoPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to init repo: %s, %w", string(output), err)
	}

	return nil
}

// AddRemote adds a remote to a git repository
func AddRemote(repoPath, remoteName, remoteURL string) error {
	// Verify it's a git repository
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Maybe it's already in .git directory
		if _, err := os.Stat(filepath.Join(repoPath, "HEAD")); os.IsNotExist(err) {
			return fmt.Errorf("not a git repository: %s", repoPath)
		}
	}

	// Check if remote already exists
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", remoteName)
	if err := cmd.Run(); err == nil {
		// Remote exists, update it
		cmd = exec.Command("git", "-C", repoPath, "remote", "set-url", remoteName, remoteURL)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to update remote: %s, %w", string(output), err)
		}
		return nil
	}

	// Add new remote
	cmd = exec.Command("git", "-C", repoPath, "remote", "add", remoteName, remoteURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add remote: %s, %w", string(output), err)
	}

	return nil
}

// RemoveRemote removes a remote from a git repository
func RemoveRemote(repoPath, remoteName string) error {
	cmd := exec.Command("git", "-C", repoPath, "remote", "remove", remoteName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove remote: %s, %w", string(output), err)
	}
	return nil
}

// GetRemoteURL gets the URL of a remote
func GetRemoteURL(repoPath, remoteName string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", remoteName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("remote '%s' not found", remoteName)
	}
	return strings.TrimSpace(string(output)), nil
}

// IsGitRepo checks if the given path is a git repository
func IsGitRepo(path string) bool {
	// Check for .git directory (working tree)
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}

	// Check for bare repository (has HEAD file directly)
	headFile := filepath.Join(path, "HEAD")
	if _, err := os.Stat(headFile); err == nil {
		return true
	}

	return false
}

// IsBareRepo checks if the given path is a bare git repository
func IsBareRepo(path string) bool {
	// Bare repos have HEAD directly in the directory
	headFile := filepath.Join(path, "HEAD")
	configFile := filepath.Join(path, "config")

	if _, err := os.Stat(headFile); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return false
	}

	// Check if it's actually bare by looking at config
	cmd := exec.Command("git", "-C", path, "config", "--get", "core.bare")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "true"
}

// GetDefaultBranch returns the default branch name of a repository
func GetDefaultBranch(repoPath string) (string, error) {
	// Try to get HEAD reference
	cmd := exec.Command("git", "-C", repoPath, "symbolic-ref", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to checking common branch names
		for _, branch := range []string{"main", "master"} {
			// nolint:gosec // G204: Subprocess launched with variable. repoPath and branch are trusted components.
			checkCmd := exec.Command("git", "-C", repoPath, "rev-parse", "--verify", branch)
			if checkCmd.Run() == nil {
				return branch, nil
			}
		}
		return "", fmt.Errorf("could not determine default branch")
	}
	return strings.TrimSpace(string(output)), nil
}

// CloneRepo clones a repository
func CloneRepo(sourceURL, destPath string) error {
	cmd := exec.Command("git", "clone", sourceURL, destPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone: %s, %w", string(output), err)
	}
	return nil
}

// PushToRemote pushes to a remote repository
func PushToRemote(repoPath, remoteName, branch string) error {
	cmd := exec.Command("git", "-C", repoPath, "push", remoteName, branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push: %s, %w", string(output), err)
	}
	return nil
}

// PushToRemoteUpstream pushes to a remote repository and sets upstream
func PushToRemoteUpstream(repoPath, remoteName, branch string) error {
	cmd := exec.Command("git", "-C", repoPath, "push", "-u", remoteName, branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push: %s, %w", string(output), err)
	}
	return nil
}

// GetRemotes returns a list of all remotes
func GetRemotes(repoPath string) ([]Remote, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	remoteMap := make(map[string]string)

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// parts[0] is name, parts[1] is url
			remoteMap[parts[0]] = parts[1]
		}
	}

	var remotes []Remote
	for name, url := range remoteMap {
		remotes = append(remotes, Remote{Name: name, URL: url})
	}
	return remotes, nil
}

// GetBranches returns a list of branches
func GetBranches(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// GetLastCommit returns info about the last commit on a branch
func GetLastCommit(repoPath, branch string) (*CommitInfo, error) {
	// Format: hash|author|date|msg
	format := "%h|%an|%ar|%s"
	// nolint:gosec // Trusted input
	cmd := exec.Command("git", "-C", repoPath, "log", "-1", fmt.Sprintf("--format=%s", format), branch)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid log format")
	}

	return &CommitInfo{
		Hash:   parts[0],
		Author: parts[1],
		Date:   parts[2],
		Msg:    parts[3],
	}, nil
}

// SetHead sets the HEAD symbolic ref (for default branch)
func SetHead(repoPath, branch string) error {
	// nolint:gosec // Trusted input
	cmd := exec.Command("git", "-C", repoPath, "symbolic-ref", "HEAD", "refs/heads/"+branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set HEAD: %s, %w", string(output), err)
	}
	return nil
}

// SetUpstream sets the upstream for a branch
func SetUpstream(repoPath, branch, remote, remoteBranch string) error {
	// nolint:gosec // Trusted input
	cmd := exec.Command("git", "-C", repoPath, "branch", "--set-upstream-to="+remote+"/"+remoteBranch, branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set upstream: %s, %w", string(output), err)
	}
	return nil
}

// GetUpstream gets the upstream for a branch (returns "remote/branch")
func GetUpstream(repoPath, branch string) (string, error) {
	// nolint:gosec // Trusted input
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no upstream configured")
	}
	return strings.TrimSpace(string(output)), nil
}
