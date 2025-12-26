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
)

var (
	// Version information
	Version = "1.0.5"
	// BuildDate is the timestamp when the binary was built
	BuildDate = "unknown"
	// GitCommit is the commit hash of the build
	GitCommit = "unknown"
)

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "lgh",
	Short: "LGH - LocalGitHub: A lightweight local Git hosting service",
	Long: `LGH (LocalGitHub) is a lightweight local Git hosting service.

It provides a simple way to host Git repositories locally via HTTP,
similar to GitHub but running entirely on your machine.

FEATURES:
  • HTTP Git hosting via git-http-backend
  • Daemon mode for background operation
  • Built-in authentication for secure sharing
  • mDNS discovery for LAN access
  • Easy tunnel integration for remote access

QUICK START:
  $ lgh init              # Initialize LGH environment
  $ lgh serve -d          # Start server in background
  $ lgh add .             # Add current directory as a repo
  $ git push lgh main     # Push to your local GitHub!

SERVER MANAGEMENT:
  $ lgh serve             # Start in foreground
  $ lgh serve -d          # Start in background (daemon)
  $ lgh status            # Check server status
  $ lgh stop              # Stop the server

COMMON OPTIONS:
  $ lgh serve --port 8080       # Use custom port
  $ lgh serve --bind 0.0.0.0    # Allow LAN access
  $ lgh serve --read-only       # Disable push
  $ lgh auth setup              # Enable authentication

For more information, visit: https://github.com/JoeGlenn1213/lgh`,
	Version: Version,
}

func init() {
	// Add version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(`LGH (LocalGitHub) v%s
Build Date: %s
Git Commit: %s
`, Version, BuildDate, GitCommit))

	// Add -v shorthand for version (Cobra uses --version by default)
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(tunnelCmd)
	rootCmd.AddCommand(authCmd)

	// New in v1.0.4
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(remoteCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(doctorCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
