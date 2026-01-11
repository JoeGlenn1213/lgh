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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/JoeGlenn1213/lgh/internal/ignore"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/internal/server"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up [message]",
	Short: "One-click commit and push to LGH (smart ignore enabled)",
	Long: `The 'up' command is a streamlined workflow that:
  1. Ensures .gitignore exists (auto-generates based on project type)
  2. Stages all changes (git add .)
  3. Commits with the provided message
  4. Pushes to LGH

This is the fastest way to backup your code to LGH.`,
	Example: `  # Quick backup with commit message
  lgh up "完成鉴权模块"

  # First time: add name for the repository
  lgh up -n my-awesome-project "初始化项目"

  # Force push (skip trash detection)
  lgh up "我就要推大文件" --force`,
	Args: cobra.MinimumNArgs(1),
	Run:  runUp,
}

var (
	upName     string
	upForce    bool
	upNoIgnore bool
)

func init() {
	upCmd.Flags().StringVarP(&upName, "name", "n", "", "Repository name (for first-time add)")
	upCmd.Flags().BoolVarP(&upForce, "force", "f", false, "Skip trash detection and force push")
	upCmd.Flags().BoolVar(&upNoIgnore, "no-ignore", false, "Don't auto-generate .gitignore")
	rootCmd.AddCommand(upCmd)
}

func runUp(_ *cobra.Command, args []string) {
	message := args[0]

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		ui.Error("Failed to get current directory: %v", err)
		os.Exit(1)
	}

	// Check if server is running
	running, _ := server.IsRunning()
	if !running {
		ui.Error("LGH server is not running. Start it with: lgh serve -d")
		os.Exit(1)
	}

	// Step 1: Ensure .gitignore exists (unless --no-ignore)
	if !upNoIgnore {
		var pType ignore.ProjectType
		pType, err = ignore.EnsureGitignore(cwd)
		if err != nil {
			ui.Warning("Failed to create .gitignore: %v", err)
		} else if pType != ignore.ProjectTypeUnknown {
			ui.Success("Created .gitignore for %s project", pType)
		}
	}

	// Step 2: Check if this is a git repository
	if !isGitRepo(cwd) {
		ui.Info("Initializing git repository...")
		if gitInitErr := runGitCommand(cwd, "init"); gitInitErr != nil {
			ui.Error("Failed to initialize git: %v", gitInitErr)
			os.Exit(1)
		}
	}

	// Step 3: Check if repo is registered with LGH
	reg := registry.New()
	repoMapping, err := reg.FindBySourcePath(cwd)
	if err != nil {
		// Not registered yet, need to add
		ui.Info("Repository not registered with LGH, adding now...")
		name := upName
		if name == "" {
			name = filepath.Base(cwd)
		}
		// Use existing add logic
		if err := addRepoToLGH(cwd, name, false); err != nil {
			ui.Error("Failed to add repository: %v", err)
			os.Exit(1)
		}
		ui.Success("Added repository '%s' to LGH", name)
	} else {
		ui.Info("Using existing LGH repository: %s", repoMapping.Name)
	}

	// Step 4: Trash detection (unless --force)
	if !upForce {
		report, err := ignore.DetectTrash(cwd)
		if err != nil {
			ui.Warning("Trash detection failed: %v", err)
		} else if len(report.Items) > 0 {
			printTrashReport(report)
			if report.HasBlocking {
				ui.Error("Blocking issues found. Fix them or use --force to override.")
				os.Exit(1)
			}
		}
	}

	// Step 5: Git add all
	ui.Info("Staging changes...")
	if err := runGitCommand(cwd, "add", "."); err != nil {
		ui.Error("Failed to stage changes: %v", err)
		os.Exit(1)
	}

	// Step 6: Git commit
	ui.Info("Committing: %s", message)
	if err := runGitCommand(cwd, "commit", "-m", message); err != nil {
		// Check if there's nothing to commit
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			ui.Warning("Nothing to commit, working tree clean")
		} else {
			ui.Error("Failed to commit: %v", err)
			os.Exit(1)
		}
	}

	// Step 7: Git push to LGH
	ui.Info("Pushing to LGH...")
	if err := runGitCommandWithBufferHint(cwd, "push", "-u", "lgh", "HEAD"); err != nil {
		os.Exit(1)
	}

	ui.Success("🚀 Done! Changes pushed to LGH")
}

func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printTrashReport(report *ignore.TrashReport) {
	ui.Warning("🚨 Trash Detection Report")
	fmt.Println()
	for _, item := range report.Items {
		icon := "⚠️"
		if item.Blocking {
			icon = "❌"
		}
		if item.Size > 0 {
			fmt.Printf("  %s %s (%s) - %s\n", icon, item.Path, ignore.FormatHumanSize(item.Size), item.Message)
		} else {
			fmt.Printf("  %s %s - %s\n", icon, item.Path, item.Message)
		}
	}
	fmt.Println()
	ui.Info("Total size: %s", ignore.FormatHumanSize(report.TotalSize))
}

// runGitCommandWithBufferHint runs a git command and provides helpful hints if it fails due to buffer issues
func runGitCommandWithBufferHint(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	// Print output
	if len(output) > 0 {
		fmt.Print(string(output))
	}

	if err != nil {
		outputStr := string(output)

		// Detect common buffer/size related errors
		if containsBufferError(outputStr) {
			ui.Error("Push failed - likely due to large files or buffer size limit")
			fmt.Println()
			ui.Info("💡 Try increasing git's HTTP buffer size:")
			fmt.Println("   git config --global http.postBuffer 524288000  # 500MB")
			fmt.Println()
			ui.Info("💡 Or if your repo is very large, try:")
			fmt.Println("   git config --global http.postBuffer 1048576000  # 1GB")
			fmt.Println()
			ui.Warning("Note: LGH does NOT auto-modify your git config. Please run the above command manually.")
		} else {
			ui.Error("Push failed: %v", err)
		}
		return err
	}
	return nil
}

// containsBufferError checks if the error output indicates a buffer/size issue
func containsBufferError(output string) bool {
	bufferHints := []string{
		"RPC failed",
		"curl",
		"error: RPC failed",
		"The remote end hung up unexpectedly",
		"fatal: the remote end hung up unexpectedly",
		"error: failed to push some refs",
		"send-pack",
		"HTTP 413",
		"HTTP/1.1 413",
	}
	for _, hint := range bufferHints {
		if contains(output, hint) {
			return true
		}
	}
	return false
}

// contains checks if s contains substr (case-insensitive-ish for common errors)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// addRepoToLGH adds a repository to LGH (reuses logic from add.go)
// This is a simplified version; the full version should call the add command's internal function
func addRepoToLGH(repoPath, name string, noRemote bool) error {
	// This should ideally call the shared add logic
	// For now, we'll call the add command directly
	addArgs := []string{"add", repoPath, "--name", name}
	if noRemote {
		addArgs = append(addArgs, "--no-remote")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	//nolint:gosec // G204: exe resolved internally, args constructed from input
	cmd := exec.Command(exe, addArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
