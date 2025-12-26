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

	"github.com/spf13/cobra"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/git"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage and inspect repository state",
	Long:  `Manage local repository state, inspect bare repositories, and check synchronization status.`,
}

// lgh repo status
var repoStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current repository status and remotes",
	Long: `In a git project, show clearly which remote you are connected to, 
branch status, and LGH integration info.`,
	RunE: runRepoStatus,
}

// lgh repo inspect <name>
var repoInspectCmd = &cobra.Command{
	Use:   "inspect <name>",
	Short: "Inspect a bare repository in LGH registry",
	Long:  `Show details about a bare repository stored in LGH, including HEAD, branches, and last commit.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoInspect,
}

// lgh repo set-default <name> <branch>
var repoSetDefaultCmd = &cobra.Command{
	Use:   "set-default <name> <branch>",
	Short: "Set default branch for a bare repository",
	Long:  `Change the HEAD symbolic ref of a bare repository in LGH.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runRepoSetDefault,
}

func init() {
	repoCmd.AddCommand(repoStatusCmd)
	repoCmd.AddCommand(repoInspectCmd)
	repoCmd.AddCommand(repoSetDefaultCmd)
}

func runRepoStatus(_ *cobra.Command, _ []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !git.IsGitRepo(wd) {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	// 1. Basic Info
	repoName := filepath.Base(wd)
	ui.Info("📦 Repository: %s", repoName)
	ui.Info("📍 Local Path: %s", wd)
	fmt.Println()

	// Pre-fetch Branch & Upstream Info to determine active remote
	head, err := git.GetDefaultBranch(wd) // This gets the current checked out branch/HEAD
	if err != nil {
		head = "unknown"
	}
	upstream, err := git.GetUpstream(wd, head)
	var activeRemoteName string
	if err == nil && upstream != "" {
		// upstream is usually "remote/branch", e.g. "origin/main"
		parts := strings.SplitN(upstream, "/", 2)
		if len(parts) > 0 {
			activeRemoteName = parts[0]
		}
	}

	// 2. Remotes
	ui.Info("🔗 Remotes:")
	remotes, err := git.GetRemotes(wd)
	if err != nil {
		ui.Warning("  Failed to get remotes: %v", err)
	} else {
		cfg := config.Get()
		lghPortStr := fmt.Sprintf(":%d", cfg.Port)

		for _, remote := range remotes {
			suffix := ""

			// Check if this is the active remote (based on upstream)
			if activeRemoteName != "" && remote.Name == activeRemoteName {
				suffix += ui.Green("   ✅ active")
			}

			// Determine if it's an LGH remote (for info purposes)
			isLGH := remote.Name == "lgh" || strings.Contains(remote.URL, "127.0.0.1"+lghPortStr) || strings.Contains(remote.URL, "localhost"+lghPortStr)
			if isLGH {
				// Only add LGH tag if valid, and maybe distinct icon
				// If it's already active, we don't need too much noise, but "lgh" name is self-explanatory.
				// Let's add a small house icon if it's LGH but not "lgh" by name, or just to be cool.
				// Actually, user was confused by "(lgh)" suffix attached to checkmark.
				// Let's leave it clean. If remote.Name is "lgh", that's enough.
				// If remote.Name is NOT "lgh" but points to localhost, maybe show it.
				if remote.Name != "lgh" {
					suffix += ui.Cyan("  (LGH)")
				}
			}

			fmt.Printf("  - %-8s → %s%s\n", remote.Name, remote.URL, suffix)
		}
	}
	fmt.Println()

	// 3. Branch Info
	ui.Info("🌿 Branch:")
	fmt.Printf("  - HEAD        : %s\n", head)

	if upstream == "" {
		fmt.Printf("  - Upstream    : %s\n", ui.Gray("(none)"))
	} else {
		fmt.Printf("  - Upstream    : %s\n", upstream)
	}

	// 4. LGH Default Branch (if registered)
	reg := registry.New()
	var lghRepo *registry.RepoMapping
	if existing, err := reg.FindBySourcePath(wd); err == nil {
		lghRepo = existing
	} else if existing, err := reg.Find(repoName); err == nil {
		lghRepo = existing
	}

	if lghRepo != nil {
		defaultBranch, err := git.GetDefaultBranch(lghRepo.BarePath)
		if err != nil {
			defaultBranch = "unknown"
		}
		fmt.Printf("  - LGH default : %s\n", defaultBranch)
	}
	fmt.Println()

	// 5. Push Target Summary
	if upstream != "" {
		// Highlight if pushing to LGH
		if activeRemoteName == "lgh" {
			ui.Success("🟢 Push target: %s", upstream)
		} else {
			ui.Info("⚪ Push target: %s", upstream)
		}
	} else {
		ui.Warning("🔴 No upstream configured. Use 'lgh remote use lgh' to set.")
	}
	fmt.Println()

	return nil
}

func runRepoInspect(_ *cobra.Command, args []string) error {
	name := args[0]

	reg := registry.New()
	repo, err := reg.Find(name)
	if err != nil {
		return fmt.Errorf("repository '%s' not found in registry", name)
	}

	ui.Title("Inspecting %s", name)

	ui.Info("📦 Repo: %s", filepath.Base(repo.BarePath))
	ui.Info("📂 Source: %s", repo.SourcePath)
	ui.Info("🏰 Bare:   %s", repo.BarePath)
	fmt.Println()

	ui.Info("🧠 Bare Repo Info:")

	// HEAD
	head, err := git.GetDefaultBranch(repo.BarePath)
	if err != nil {
		ui.Warning("  - HEAD : unknown (%v)", err)
	} else {
		ui.Info("  - Default branch : %s", head)
	}

	// Branches
	branches, err := git.GetBranches(repo.BarePath)
	if err != nil {
		ui.Warning("  - Branches : failed to list (%v)", err)
	} else {
		ui.Info("  - Branches : %s", strings.Join(branches, ", "))
	}
	fmt.Println()

	// Last Push Info
	if head != "" {
		ui.Info("📌 Last Push (on %s):", head)
		commit, err := git.GetLastCommit(repo.BarePath, head)
		if err != nil {
			ui.Gray("  No commits yet or failed to read log.")
		} else {
			fmt.Printf("  - Commit : %s\n", commit.Hash)
			fmt.Printf("  - Author : %s\n", commit.Author)
			fmt.Printf("  - Time   : %s\n", commit.Date)
			fmt.Printf("  - Msg    : %s\n", commit.Msg)
		}
	}
	fmt.Println()

	return nil
}

func runRepoSetDefault(_ *cobra.Command, args []string) error {
	name := args[0]
	branch := args[1]

	reg := registry.New()
	repo, err := reg.Find(name)
	if err != nil {
		return fmt.Errorf("repository '%s' not found in registry", name)
	}

	ui.Info("Setting default branch of '%s' to '%s'...", name, branch)

	if err := git.SetHead(repo.BarePath, branch); err != nil {
		return fmt.Errorf("failed to set HEAD: %w", err)
	}

	ui.Success("Default branch updated.")
	return nil
}
