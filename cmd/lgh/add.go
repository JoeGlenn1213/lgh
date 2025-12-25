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

// Package main provides the LGH command line interface
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/git"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/internal/server"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var (
	repoName string
	noRemote bool
)

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Add a local repository to LGH",
	Long: `Add a local Git repository to LGH for HTTP hosting.

This command:
  1. Creates a bare repository in ~/.localgithub/repos/
  2. Adds a remote named 'lgh' to your local repository
  3. Registers the mapping in mappings.yaml

If no path is specified, the current directory is used.

Examples:
  lgh add                      # Add current directory
  lgh add ./my-project         # Add specific path
  lgh add . --name my-app      # Add with custom name
  lgh add . --no-remote        # Don't add remote to source repo`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&repoName, "name", "n", "", "Custom name for the repository")
	addCmd.Flags().BoolVar(&noRemote, "no-remote", false, "Don't add 'lgh' remote to the source repository")
}

func runAdd(_ *cobra.Command, args []string) error {
	// Ensure initialized
	if err := ensureInitialized(); err != nil {
		return err
	}

	// Check git is installed
	if _, err := git.CheckGitInstalled(); err != nil {
		return err
	}

	// Get path
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Expand and resolve path
	absPath, err := expandPath(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// Check if it's a git repository
	if !git.IsGitRepo(absPath) {
		return fmt.Errorf("not a git repository: %s", absPath)
	}

	// Determine repository name
	name := repoName
	if name == "" {
		name = filepath.Base(absPath)
	}

	// Ensure name ends with .git for bare repo
	bareRepoName := name
	if filepath.Ext(bareRepoName) != ".git" {
		bareRepoName = name + ".git"
	}

	ui.Title("Adding Repository: %s", name)

	// Check if already registered
	reg := registry.New()
	if reg.Exists(name) {
		return fmt.Errorf("repository '%s' is already registered. Use 'lgh remove %s' first", name, name)
	}

	// Create bare repository path
	cfg := config.Get()
	barePath := filepath.Join(cfg.ReposDir, bareRepoName)

	// Check if bare repo already exists
	if _, err := os.Stat(barePath); err == nil {
		return fmt.Errorf("bare repository already exists at %s", barePath)
	}

	// Create bare repository
	ui.Info("Creating bare repository...")
	if err := git.InitBareRepo(barePath); err != nil {
		return err
	}
	ui.Success("Created %s", barePath)

	// Build the remote URL
	remoteURL := fmt.Sprintf("http://%s:%d/%s", cfg.BindAddress, cfg.Port, bareRepoName)

	// Add remote to source repository
	if !noRemote {
		ui.Info("Adding 'lgh' remote to source repository...")
		if err := git.AddRemote(absPath, "lgh", remoteURL); err != nil {
			ui.Warning("Failed to add remote: %v", err)
			ui.Info("You can add it manually: git remote add lgh %s", remoteURL)
		} else {
			ui.Success("Added remote 'lgh' -> %s", remoteURL)
		}
	}

	// Register mapping
	ui.Info("Registering repository...")
	if err := reg.Add(name, absPath, barePath); err != nil {
		// Cleanup: remove bare repo if registration fails
		_ = os.RemoveAll(barePath)
		return fmt.Errorf("failed to register repository: %w", err)
	}
	ui.Success("Registered in mappings.yaml")

	// Print success and next steps
	fmt.Println()
	ui.Success("Repository '%s' added successfully!", name)
	fmt.Println()

	// Check if server is running
	if running, _ := server.IsRunning(); running {
		ui.Info("Clone URL: %s", ui.URL(remoteURL))
		fmt.Println()
		ui.Info("Push your code:")
		ui.Command("git push lgh main")
	} else {
		ui.Warning("Server is not running!")
		ui.Info("Start the server first:")
		ui.Command("lgh serve")
		fmt.Println()
		ui.Info("Then push your code:")
		ui.Command("git push lgh main")
	}
	fmt.Println()

	return nil
}
