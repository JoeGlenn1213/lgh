package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
	if err := os.MkdirAll(parentDir, 0755); err != nil {
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
