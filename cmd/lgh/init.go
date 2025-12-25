package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize LGH environment",
	Long: `Initialize the LGH environment by creating the necessary directories
and configuration files.

This command creates:
  • ~/.localgithub/           Main data directory
  • ~/.localgithub/repos/     Repository storage
  • ~/.localgithub/config.yaml Default configuration
  • ~/.localgithub/mappings.yaml Repository mappings`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	ui.Title("Initializing LGH Environment")

	lghDir := config.GetLGHDir()
	reposDir := config.GetReposDir()

	// Check if already initialized
	if _, err := os.Stat(lghDir); err == nil {
		configPath := config.GetConfigPath()
		if _, err := os.Stat(configPath); err == nil {
			ui.Warning("LGH is already initialized at %s", lghDir)
			ui.Info("Use 'lgh serve' to start the server")
			return nil
		}
	}

	// Create directories
	ui.Info("Creating directory structure...")

	dirs := []string{lghDir, reposDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		ui.Success("Created %s", dir)
	}

	// Create default configuration
	ui.Info("Creating default configuration...")
	if err := config.CreateDefaultConfig(); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	ui.Success("Created %s", config.GetConfigPath())

	// Create empty mappings file
	mappingsPath := config.GetMappingsPath()
	if err := os.WriteFile(mappingsPath, []byte("repos: []\n"), 0644); err != nil {
		return fmt.Errorf("failed to create mappings file: %w", err)
	}
	ui.Success("Created %s", mappingsPath)

	// Print success message
	fmt.Println()
	ui.Success("LGH initialized successfully!")
	fmt.Println()
	ui.Info("Next steps:")
	ui.Command("lgh serve              # Start the HTTP server")
	ui.Command("lgh add <path>         # Add a repository")
	ui.Command("lgh list               # List all repositories")
	fmt.Println()

	return nil
}

// ensureInitialized checks if LGH is initialized
func ensureInitialized() error {
	lghDir := config.GetLGHDir()
	configPath := config.GetConfigPath()

	if _, err := os.Stat(lghDir); os.IsNotExist(err) {
		return fmt.Errorf("LGH is not initialized. Run 'lgh init' first")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("LGH configuration not found. Run 'lgh init' first")
	}

	return nil
}

// expandPath expands ~ and resolves symlinks
func expandPath(path string) (string, error) {
	if path == "" {
		return os.Getwd()
	}

	// Expand ~
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If path doesn't exist yet, return absolute path
		if os.IsNotExist(err) {
			return absPath, nil
		}
		return "", err
	}

	return realPath, nil
}
