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
	"strings"

	"github.com/spf13/cobra"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/git"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <repository> [directory]",
	Short: "Clone a repository from LGH",
	Long: `Clone a repository from the local LGH server.

This is a wrapper around 'git clone' that simplifies cloning from localhost.
You can provide just the repository name instead of the full URL.

Examples:
  lgh clone my-app              # Clones http://127.0.0.1:9418/my-app.git
  lgh clone my-app ./dev/app    # Clone to specific directory`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runClone,
}

func runClone(_ *cobra.Command, args []string) error {
	repoName := args[0]
	destPath := ""
	if len(args) > 1 {
		destPath = args[1]
	}

	cloneURL := repoName

	// If it doesn't look like a URL and doesn't have .git extension, assume it's an LGH repo name
	if !strings.Contains(repoName, "://") && !strings.Contains(repoName, "@") {
		// Ensure initialized
		if err := ensureInitialized(); err != nil {
			return err
		}

		// Check if repo exists in registry
		reg := registry.New()
		if !reg.Exists(repoName) {
			return fmt.Errorf("repository '%s' not found in LGH registry", repoName)
		}

		// Build URL
		cfg := config.Get()

		// Handle auth if enabled (simple hint)
		authPrefix := ""
		if cfg.AuthEnabled {
			// We don't store plain password, so we can't fully auto-fill,
			// but we can look for credentials or just warn the user.
			// For now, let git handle the auth prompt or credential helper.
			ui.Info("Authentication is enabled. You may need to enter credentials.")
		}

		cloneURL = fmt.Sprintf("http://%s%s:%d/%s.git", authPrefix, cfg.BindAddress, cfg.Port, repoName)
		ui.Info("Cloning %s...", cloneURL)
	}

	// Execute git clone
	if err := git.CloneRepo(cloneURL, destPath); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	ui.Success("Successfully cloned %s", repoName)
	return nil
}
