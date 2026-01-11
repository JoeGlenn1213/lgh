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
	"github.com/JoeGlenn1213/lgh/pkg/ui"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save [message]",
	Short: "Local save: stage and commit without pushing",
	Long: `The 'save' command is a local-only workflow that:
  1. Ensures .gitignore exists (auto-generates based on project type)
  2. Stages all changes (git add .)
  3. Commits with the provided message

Unlike 'lgh up', this does NOT push to LGH. Use this for work-in-progress saves.`,
	Example: `  # Save work in progress
  lgh save "WIP: 还没写完别推"

  # Force save (skip trash detection)
  lgh save "临时存档" --force`,
	Args: cobra.MinimumNArgs(1),
	Run:  runSave,
}

var (
	saveForce    bool
	saveNoIgnore bool
)

func init() {
	saveCmd.Flags().BoolVarP(&saveForce, "force", "f", false, "Skip trash detection")
	saveCmd.Flags().BoolVar(&saveNoIgnore, "no-ignore", false, "Don't auto-generate .gitignore")
	rootCmd.AddCommand(saveCmd)
}

func runSave(cmd *cobra.Command, args []string) {
	message := args[0]

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		ui.Error("Failed to get current directory: %v", err)
		os.Exit(1)
	}

	// Step 1: Ensure .gitignore exists (unless --no-ignore)
	if !saveNoIgnore {
		projectType, err := ignore.EnsureGitignore(cwd)
		if err != nil {
			ui.Warning("Failed to create .gitignore: %v", err)
		} else if projectType != ignore.ProjectTypeUnknown {
			ui.Success("Created .gitignore for %s project", projectType)
		}
	}

	// Step 2: Check if this is a git repository
	if !isGitRepoSave(cwd) {
		ui.Info("Initializing git repository...")
		if err := runGitCommandSave(cwd, "init"); err != nil {
			ui.Error("Failed to initialize git: %v", err)
			os.Exit(1)
		}
	}

	// Step 3: Trash detection (unless --force)
	if !saveForce {
		report, err := ignore.DetectTrash(cwd)
		if err != nil {
			ui.Warning("Trash detection failed: %v", err)
		} else if len(report.Items) > 0 {
			printTrashReportSave(report)
			if report.HasBlocking {
				ui.Error("Blocking issues found. Fix them or use --force to override.")
				os.Exit(1)
			}
		}
	}

	// Step 4: Git add all
	ui.Info("Staging changes...")
	if err := runGitCommandSave(cwd, "add", "."); err != nil {
		ui.Error("Failed to stage changes: %v", err)
		os.Exit(1)
	}

	// Step 5: Git commit
	ui.Info("Committing: %s", message)
	if err := runGitCommandSave(cwd, "commit", "-m", message); err != nil {
		// Check if there's nothing to commit
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			ui.Warning("Nothing to commit, working tree clean")
		} else {
			ui.Error("Failed to commit: %v", err)
			os.Exit(1)
		}
	}

	ui.Success("📦 Saved locally! Use 'lgh up' when ready to push.")
}

func isGitRepoSave(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

func runGitCommandSave(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printTrashReportSave(report *ignore.TrashReport) {
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
