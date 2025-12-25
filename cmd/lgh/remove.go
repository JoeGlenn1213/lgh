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
	"path/filepath"
	"strings"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/git"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	removeForce    bool
	removeKeepBare bool
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a repository from LGH",
	Long: `Remove a repository from LGH.

This command:
  1. Removes the repository from mappings.yaml
  2. Optionally deletes the bare repository
  3. Optionally removes the 'lgh' remote from source repo

Examples:
  lgh remove my-app            # Remove with confirmation
  lgh remove my-app --force    # Remove without confirmation
  lgh remove my-app --keep-bare # Keep the bare repository`,
	Aliases: []string{"rm", "delete"},
	Args:    cobra.ExactArgs(1),
	RunE:    runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Skip confirmation")
	removeCmd.Flags().BoolVar(&removeKeepBare, "keep-bare", false, "Keep the bare repository")
}

// isPathSafe checks if targetPath is safely within basePath
// SECURITY: Prevents path traversal attacks
func isPathSafe(basePath, targetPath string) bool {
	// Resolve to absolute paths
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// Resolve any symlinks
	realBase, err := filepath.EvalSymlinks(absBase)
	if err != nil {
		realBase = absBase
	}
	realTarget, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		// Target might not exist yet, use absolute path
		realTarget = absTarget
	}

	// Ensure target is within base (with trailing slash to prevent prefix attacks)
	if !strings.HasSuffix(realBase, string(os.PathSeparator)) {
		realBase += string(os.PathSeparator)
	}

	return strings.HasPrefix(realTarget, realBase) || realTarget == strings.TrimSuffix(realBase, string(os.PathSeparator))
}

func runRemove(cmd *cobra.Command, args []string) error {
	// Ensure initialized
	if err := ensureInitialized(); err != nil {
		return err
	}

	name := args[0]

	// Find repository
	reg := registry.New()
	repo, err := reg.Find(name)
	if err != nil {
		return fmt.Errorf("repository '%s' not found", name)
	}

	// SECURITY: Validate BarePath is within the configured ReposDir
	cfg := config.Get()
	if !isPathSafe(cfg.ReposDir, repo.BarePath) {
		return fmt.Errorf("security error: bare repository path '%s' is outside of repos directory '%s'",
			repo.BarePath, cfg.ReposDir)
	}

	ui.Title("Remove Repository: %s", name)
	fmt.Println()
	ui.Info("Source Path: %s", repo.SourcePath)
	ui.Info("Bare Path:   %s", repo.BarePath)
	fmt.Println()

	// Confirm if not forced
	if !removeForce {
		ui.Warning("This will remove the repository from LGH.")
		if !removeKeepBare {
			ui.Warning("The bare repository at %s will be DELETED.", repo.BarePath)
		}
		fmt.Println()

		fmt.Print("Are you sure? [y/N]: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			ui.Info("Cancelled.")
			return nil
		}
		fmt.Println()
	}

	// Remove 'lgh' remote from source repository if it exists
	if _, err := os.Stat(repo.SourcePath); err == nil {
		ui.Info("Removing 'lgh' remote from source repository...")
		if err := git.RemoveRemote(repo.SourcePath, "lgh"); err != nil {
			ui.Warning("Could not remove remote: %v", err)
		} else {
			ui.Success("Removed 'lgh' remote")
		}
	}

	// Delete bare repository
	if !removeKeepBare {
		ui.Info("Deleting bare repository...")

		// Double-check the path is a git bare repository before deleting
		if !git.IsBareRepo(repo.BarePath) {
			ui.Warning("Path is not a valid bare repository, skipping deletion for safety")
		} else {
			if err := os.RemoveAll(repo.BarePath); err != nil {
				ui.Warning("Could not delete bare repository: %v", err)
			} else {
				ui.Success("Deleted %s", repo.BarePath)
			}
		}
	}

	// Remove from registry
	ui.Info("Removing from registry...")
	if err := reg.Remove(name); err != nil {
		return fmt.Errorf("failed to remove from registry: %w", err)
	}
	ui.Success("Removed from mappings.yaml")

	fmt.Println()
	ui.Success("Repository '%s' removed successfully!", name)
	fmt.Println()

	return nil
}
