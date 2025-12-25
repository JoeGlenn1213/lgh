package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information
	Version   = "1.0.0"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "lgh",
	Short: "LGH - LocalGitHub: A lightweight local Git hosting service",
	Long: `LGH (LocalGitHub) is a lightweight local Git hosting service.

It provides a simple way to host Git repositories locally via HTTP,
similar to GitHub but running entirely on your machine.

Features:
  • HTTP Git hosting via git-http-backend
  • Simple CLI for managing repositories
  • mDNS discovery for LAN access
  • Easy tunnel integration for remote access

Quick Start:
  1. lgh init          Initialize LGH environment
  2. lgh serve         Start the HTTP server
  3. lgh add .         Add current directory as a repository
  4. git push lgh main Push to your local GitHub!

For more information, visit: https://github.com/JoeGlenn1213/lgh`,
	Version: Version,
}

func init() {
	// Add version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(`LGH (LocalGitHub) v%s
Build Date: %s
Git Commit: %s
`, Version, BuildDate, GitCommit))

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(tunnelCmd)
	rootCmd.AddCommand(authCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
