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

	"github.com/spf13/cobra"

	"github.com/JoeGlenn1213/lgh/internal/git"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remote connections",
	Long:  `Helper commands to manage git remotes effectively.`,
}

var remoteUseCmd = &cobra.Command{
	Use:   "use <remote>",
	Short: "Switch the active upstream remote for current branch",
	Long: `Switch the upstream configuration of the current branch to the specified remote.

This tells git where to push/pull by default.

Example:
  lgh remote use lgh     # Switch to LGH
  lgh remote use origin  # Switch back to Origin`,
	Args: cobra.ExactArgs(1),
	RunE: runRemoteUse,
}

func init() {
	remoteCmd.AddCommand(remoteUseCmd)
}

func runRemoteUse(_ *cobra.Command, args []string) error {
	remoteName := args[0]

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !git.IsGitRepo(wd) {
		return fmt.Errorf("not a git repository")
	}

	// 1. Get current branch
	branch, err := git.GetDefaultBranch(wd) // Get current branch
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// 2. Check if remote exists
	remotes, err := git.GetRemotes(wd)
	if err != nil {
		return err
	}
	exists := false
	for _, r := range remotes {
		if r.Name == remoteName {
			exists = true
			break
		}
	}
	if !exists {
		return fmt.Errorf("remote '%s' does not exist. Add it first with 'git remote add ...' or 'lgh add .'", remoteName)
	}

	// 3. Set Upstream
	ui.Info("Switching upstream for branch '%s' to '%s'...", branch, remoteName)

	// We assume the remote branch name is same as local
	err = git.SetUpstream(wd, branch, remoteName, branch)
	if err != nil {
		// If it fails, it's likely because the remote branch doesn't exist yet.
		// We can't easily force it without pushing.
		ui.Warning("Failed to set upstream: %v", err)
		ui.Info("This usually means the branch '%s' does not exist on remote '%s' yet.", branch, remoteName)
		ui.Info("Try pushing first:")
		ui.Command(fmt.Sprintf("git push -u %s %s", remoteName, branch))
		return nil // Not strictly an error, just a guidance state
	}

	ui.Success("🔁 Active remote switched to: %s", remoteName)
	ui.Info("🌿 Branch: %s -> %s/%s", branch, remoteName, branch)

	return nil
}
