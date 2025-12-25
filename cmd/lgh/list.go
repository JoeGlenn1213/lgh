package main

import (
	"fmt"
	"os"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/internal/server"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered repositories",
	Long: `List all repositories that have been added to LGH.

Shows repository name, source path, and clone URL.

Examples:
  lgh list           # List all repositories
  lgh ls             # Short alias`,
	Aliases: []string{"ls"},
	RunE:    runList,
}

func runList(cmd *cobra.Command, args []string) error {
	// Ensure initialized
	if err := ensureInitialized(); err != nil {
		return err
	}

	// Load registry
	reg := registry.New()
	repos, err := reg.List()
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	// Check if empty
	if len(repos) == 0 {
		ui.Info("No repositories registered yet.")
		fmt.Println()
		ui.Info("Add a repository:")
		ui.Command("lgh add <path>")
		return nil
	}

	// Get server info
	cfg := config.Get()
	baseURL := fmt.Sprintf("http://%s:%d", cfg.BindAddress, cfg.Port)
	serverRunning, _ := server.IsRunning()

	// Print header
	ui.Title("Registered Repositories (%d)", len(repos))

	if !serverRunning {
		ui.Warning("Server is not running. Start with 'lgh serve'")
		fmt.Println()
	}

	// Create table
	table := ui.NewTable([]string{"Name", "Source Path", "Clone URL", "Created"})

	for _, repo := range repos {
		// Check if source path exists
		sourceExists := "✓"
		if _, err := os.Stat(repo.SourcePath); os.IsNotExist(err) {
			sourceExists = "✗ (missing)"
		}

		// Build clone URL
		cloneURL := fmt.Sprintf("%s/%s.git", baseURL, repo.Name)
		if !serverRunning {
			cloneURL = ui.Gray("(server offline)")
		}

		// Format created time
		created := repo.CreatedAt.Format("2006-01-02 15:04")

		table.AddRow([]string{
			ui.Bold(repo.Name),
			fmt.Sprintf("%s %s", repo.SourcePath, ui.Gray(sourceExists)),
			cloneURL,
			ui.Gray(created),
		})
	}

	table.Render()
	fmt.Println()

	return nil
}
