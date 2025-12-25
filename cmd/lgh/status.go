package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/internal/server"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check LGH server status",
	Long: `Check the status of the LGH server and environment.

Shows:
  • Server running status
  • Configuration details
  • Number of registered repositories
  • Health check result`,
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Ensure initialized
	if err := ensureInitialized(); err != nil {
		return err
	}

	ui.Title("LGH Status")

	// Load configuration
	cfg := config.Get()

	// Check server status
	running, pid := server.IsRunning()

	if running {
		ui.Success("Server Status: RUNNING (PID: %d)", pid)
	} else {
		ui.Error("Server Status: STOPPED")
	}

	fmt.Println()

	// Configuration info
	ui.Info("Configuration:")
	fmt.Printf("  %-15s %s\n", "Data Directory:", cfg.DataDir)
	fmt.Printf("  %-15s %s\n", "Repos Directory:", cfg.ReposDir)
	fmt.Printf("  %-15s %s:%d\n", "Listen Address:", cfg.BindAddress, cfg.Port)
	fmt.Printf("  %-15s %v\n", "Read-Only:", cfg.ReadOnly)
	fmt.Printf("  %-15s %v\n", "mDNS Enabled:", cfg.MDNSEnabled)
	fmt.Println()

	// Repository count
	reg := registry.New()
	repos, err := reg.List()
	if err != nil {
		ui.Warning("Failed to load repositories: %v", err)
	} else {
		ui.Info("Repositories: %d registered", len(repos))
	}
	fmt.Println()

	// Health check if server is running
	if running {
		ui.Info("Health Check:")
		healthURL := fmt.Sprintf("http://%s:%d/health", cfg.BindAddress, cfg.Port)

		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port), 2*time.Second)
		if err != nil {
			ui.Error("  Connection failed: %v", err)
		} else {
			conn.Close()
			ui.Success("  %s - OK", healthURL)
		}
		fmt.Println()
	}

	// Disk usage
	reposDir := cfg.ReposDir
	if _, err := os.Stat(reposDir); err == nil {
		var totalSize int64
		entries, _ := os.ReadDir(reposDir)
		for _, entry := range entries {
			if entry.IsDir() {
				size := getDirSize(fmt.Sprintf("%s/%s", reposDir, entry.Name()))
				totalSize += size
			}
		}
		ui.Info("Disk Usage: %s", formatBytes(totalSize))
		fmt.Println()
	}

	// Show URLs if running
	if running {
		ui.Info("Server URLs:")
		fmt.Printf("  Local:     %s\n", ui.URL(server.GetServerURL()))

		// Show mDNS URL if enabled
		if cfg.MDNSEnabled {
			hostname, _ := os.Hostname()
			fmt.Printf("  mDNS:      %s\n", ui.URL(fmt.Sprintf("http://%s.local:%d", hostname, cfg.Port)))
		}
		fmt.Println()
	}

	return nil
}

// getDirSize calculates the total size of a directory
func getDirSize(path string) int64 {
	var size int64
	filepath := path
	_ = filepath // silence unused warning in simple implementation

	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			size += getDirSize(path + "/" + entry.Name())
		} else {
			size += info.Size()
		}
	}

	return size
}

// formatBytes formats bytes to human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
