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
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/git"
	"github.com/JoeGlenn1213/lgh/internal/registry"
	"github.com/JoeGlenn1213/lgh/internal/server"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system environment and configuration",
	Long: `Check the system environment, dependencies, and LGH configuration for potential issues.

Checks:
  • Git installation and version
  • Configuration file validity
  • Repository registry consistency
  • Server status and port availability
  • Data directory permissions`,
	RunE: runDoctor,
}

func runDoctor(_ *cobra.Command, _ []string) error {
	ui.Title("LGH Doctor")

	hasIssues := false

	// 1. Check Git
	ui.Info("Checking Environment...")
	gitVersion, err := git.GetGitVersion()
	if err != nil {
		ui.Error("✗ Git check failed: %v", err)
		hasIssues = true
	} else {
		ui.Success("Git installed: %s", gitVersion)
	}
	fmt.Println()

	// 2. Check Configuration
	ui.Info("Checking Configuration...")
	if _, loadErr := config.Load(); loadErr != nil {
		ui.Warning("! Configuration load warning: %v", loadErr)
		// Try to continue
	}
	cfg := config.Get()

	// Check Data Dir
	if info, statErr := os.Stat(cfg.DataDir); statErr != nil {
		ui.Error("✗ Data directory issue: %v", statErr)
		hasIssues = true
	} else if !info.IsDir() {
		ui.Error("✗ Data directory is not a directory: %s", cfg.DataDir)
		hasIssues = true
	} else {
		ui.Success("Data directory: %s", cfg.DataDir)
	}

	// Check Repos Dir
	if info, statErr := os.Stat(cfg.ReposDir); statErr != nil {
		ui.Error("✗ Repos directory issue: %v", statErr)
		hasIssues = true
	} else if !info.IsDir() {
		ui.Error("✗ Repos directory is not a directory: %s", cfg.ReposDir)
		hasIssues = true
	} else {
		ui.Success("Repos directory: %s", cfg.ReposDir)
	}
	fmt.Println()

	// 3. Check Registry
	ui.Info("Checking Registry...")
	reg := registry.New()
	repos, regErr := reg.List()
	if regErr != nil {
		ui.Error("✗ Failed to list repositories: %v", regErr)
		hasIssues = true
	} else {
		ui.Success("Registry loaded: %d repositories", len(repos))

		// Verify individual repos
		missingCount := 0
		for _, repo := range repos {
			if _, statErr := os.Stat(repo.BarePath); os.IsNotExist(statErr) {
				ui.Warning("  ! Missing bare repo: %s (%s)", repo.Name, repo.BarePath)
				missingCount++
			}
		}
		if missingCount > 0 {
			ui.Warning("  Found %d missing repositories", missingCount)
			hasIssues = true
		} else if len(repos) > 0 {
			ui.Success("  All repository paths valid")
		}
	}
	fmt.Println()

	// 4. Check Server Status
	ui.Info("Checking Server...")
	running, pid := server.IsRunning()
	if running {
		ui.Success("Server is running (PID: %d)", pid)

		// Check health
		conn, connErr := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port), 2*time.Second)
		if connErr != nil {
			ui.Error("✗ Server port %d is not responding: %v", cfg.Port, connErr)
			hasIssues = true
		} else {
			_ = conn.Close()
			ui.Success("Server is reachable on port %d", cfg.Port)
		}
	} else {
		ui.Info("Server is stopped")

		// Check if port is free
		address := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port)
		ln, lnErr := net.Listen("tcp", address)
		if lnErr != nil {
			ui.Warning("! Port %d seems to be in use by another process", cfg.Port)
		} else {
			_ = ln.Close()
			ui.Success("Port %d is available", cfg.Port)
		}
	}
	fmt.Println()

	if hasIssues {
		ui.Warning("Doctor found some issues. Please check the output above.")
		return fmt.Errorf("issues found")
	}

	ui.Success("All checks passed! Your LGH environment looks healthy.")
	return nil
}
