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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/server"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

const (
	// MinPasswordLength is the minimum required password length
	MinPasswordLength = 8
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication settings",
	Long: `Manage LGH authentication settings.

Authentication is required when exposing LGH to the network.
Use Basic Auth with a username and password hash stored in config.

Subcommands:
  lgh auth setup     Interactive setup for authentication
  lgh auth hash      Generate password hash
  lgh auth disable   Disable authentication
  lgh auth status    Show current auth status`,
}

var authSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive authentication setup",
	Long: `Set up authentication interactively.

This will prompt for username and password (password is hidden), then update config.yaml.
Password must be at least 8 characters.`,
	RunE: runAuthSetup,
}

var authHashCmd = &cobra.Command{
	Use:   "hash",
	Short: "Generate password hash",
	Long: `Generate a secure password hash for configuration.

Prompts for password with hidden input.
Password must be at least 8 characters.

Example:
  lgh auth hash    # Prompts for password (hidden)`,
	RunE: runAuthHash,
}

var authDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable authentication",
	RunE:  runAuthDisable,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE:  runAuthStatus,
}

func init() {
	authCmd.AddCommand(authSetupCmd)
	authCmd.AddCommand(authHashCmd)
	authCmd.AddCommand(authDisableCmd)
	authCmd.AddCommand(authStatusCmd)
}

// readPassword reads a password from terminal with echo disabled
// Uses golang.org/x/term for cross-platform support (macOS, Linux, Windows)
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Use x/term for cross-platform hidden input
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		// Fallback to regular input if terminal control fails
		fmt.Println()
		ui.Warning("Cannot hide password input - password will be visible!")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(prompt)
		password, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(password), nil
	}

	fmt.Println() // Print newline after password
	return string(passwordBytes), nil
}

// validatePassword checks password meets requirements
func validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters (got %d)", MinPasswordLength, len(password))
	}
	return nil
}

func runAuthSetup(cmd *cobra.Command, args []string) error {
	if err := ensureInitialized(); err != nil {
		return err
	}

	ui.Title("Authentication Setup")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Get username
	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Get password (hidden input)
	password, err := readPassword("Enter password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Validate password
	if err := validatePassword(password); err != nil {
		return err
	}

	// Confirm password
	confirm, err := readPassword("Confirm password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}

	// Generate hash
	hash, err := server.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Load and update config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cfg.AuthEnabled = true
	cfg.AuthUser = username
	cfg.AuthPasswordHash = hash

	// Save config
	if err := updateAuthConfig(cfg); err != nil {
		return err
	}

	fmt.Println()
	ui.Success("Authentication configured successfully!")
	fmt.Println()
	ui.Info("Git clients can authenticate using:")
	ui.Command(fmt.Sprintf("git clone http://%s:<password>@<host>:<port>/repo.git", username))
	fmt.Println()
	ui.Info("Or configure Git credential helper:")
	ui.Command("git config credential.helper store")
	fmt.Println()

	return nil
}

func runAuthHash(cmd *cobra.Command, args []string) error {
	// Get password (hidden input)
	password, err := readPassword("Enter password to hash: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Validate password
	if err := validatePassword(password); err != nil {
		return err
	}

	hash, err := server.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	fmt.Println()
	ui.Success("Password hash generated:")
	fmt.Println()
	fmt.Println(hash)
	fmt.Println()
	ui.Info("Add this to your config.yaml:")
	fmt.Println("  auth_enabled: true")
	fmt.Println("  auth_user: <your-username>")
	fmt.Printf("  auth_password_hash: %s\n", hash)
	fmt.Println()

	return nil
}

func runAuthDisable(cmd *cobra.Command, args []string) error {
	if err := ensureInitialized(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cfg.AuthEnabled = false
	cfg.AuthUser = ""
	cfg.AuthPasswordHash = ""

	if err := updateAuthConfig(cfg); err != nil {
		return err
	}

	ui.Success("Authentication disabled.")
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	if err := ensureInitialized(); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ui.Title("Authentication Status")
	fmt.Println()

	if cfg.AuthEnabled && cfg.AuthUser != "" && cfg.AuthPasswordHash != "" {
		ui.Success("Authentication: ENABLED")
		ui.Info("  Username: %s", cfg.AuthUser)
		ui.Info("  Password: ********")
	} else {
		ui.Warning("Authentication: DISABLED")
		fmt.Println()
		ui.Info("To enable authentication:")
		ui.Command("lgh auth setup")
	}
	fmt.Println()

	return nil
}

// updateAuthConfig updates authentication settings in config file
func updateAuthConfig(cfg *config.Config) error {
	configPath := config.GetConfigPath()

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	content := string(data)

	// Update or add auth settings
	lines := strings.Split(content, "\n")
	var newLines []string
	hasAuthEnabled := false
	hasAuthUser := false
	hasAuthHash := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "auth_enabled:") {
			newLines = append(newLines, fmt.Sprintf("auth_enabled: %t", cfg.AuthEnabled))
			hasAuthEnabled = true
		} else if strings.HasPrefix(trimmed, "auth_user:") {
			newLines = append(newLines, fmt.Sprintf("auth_user: %s", strconv.Quote(cfg.AuthUser)))
			hasAuthUser = true
		} else if strings.HasPrefix(trimmed, "auth_password_hash:") {
			newLines = append(newLines, fmt.Sprintf("auth_password_hash: %s", strconv.Quote(cfg.AuthPasswordHash)))
			hasAuthHash = true
		} else {
			newLines = append(newLines, line)
		}
	}

	// Add missing auth settings
	if !hasAuthEnabled {
		newLines = append(newLines, fmt.Sprintf("auth_enabled: %t", cfg.AuthEnabled))
	}
	if !hasAuthUser {
		newLines = append(newLines, fmt.Sprintf("auth_user: %s", strconv.Quote(cfg.AuthUser)))
	}
	if !hasAuthHash {
		newLines = append(newLines, fmt.Sprintf("auth_password_hash: %s", strconv.Quote(cfg.AuthPasswordHash)))
	}

	// Write back with secure permissions (0600 = owner read/write only)
	return os.WriteFile(configPath, []byte(strings.Join(newLines, "\n")), 0600)
}
