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
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/mdns"
	"github.com/JoeGlenn1213/lgh/internal/server"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var (
	readOnlyFlag bool
	enableMDNS   bool
	serverPort   int
	bindAddress  string
	daemonFlag   bool
	allowUnsafe  bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the LGH HTTP server",
	Long: `Start the LGH HTTP server to serve Git repositories over HTTP.

The server listens on localhost by default (127.0.0.1:9418).
Use --read-only to prevent push operations.
Use --mdns to enable mDNS for local network discovery.
Use --daemon to run in background mode.

Examples:
  lgh serve                    # Start with defaults
  lgh serve --daemon           # Start in background
  lgh serve --read-only        # Start in read-only mode
  lgh serve --port 8080        # Use custom port
  lgh serve --mdns             # Enable mDNS discovery
  lgh serve --bind 0.0.0.0 --allow-unsafe # Allow public access (DANGEROUS)`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().BoolVarP(&readOnlyFlag, "read-only", "r", false, "Enable read-only mode (disable push)")
	serveCmd.Flags().BoolVarP(&enableMDNS, "mdns", "m", false, "Enable mDNS for LAN discovery")
	serveCmd.Flags().IntVarP(&serverPort, "port", "p", 0, "Port to listen on (default: 9418)")
	serveCmd.Flags().StringVarP(&bindAddress, "bind", "b", "", "Address to bind to (default: 127.0.0.1)")
	serveCmd.Flags().BoolVarP(&daemonFlag, "daemon", "d", false, "Run server in background (daemon mode)")
	serveCmd.Flags().BoolVar(&allowUnsafe, "allow-unsafe", false, "Allow binding to non-localhost without auth/read-only")
}

func runServe(_ *cobra.Command, _ []string) error {
	// Ensure initialized
	if err := ensureInitialized(); err != nil {
		return err
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Override with command-line flags (only if explicitly set)
	if serverPort > 0 {
		cfg.Port = serverPort
	}
	if bindAddress != "" {
		cfg.BindAddress = bindAddress
	}
	// CLI --read-only flag can only ENABLE read-only, never disable it
	// This ensures config.yaml read_only=true is always respected
	if readOnlyFlag {
		cfg.ReadOnly = true
	}
	if enableMDNS {
		cfg.MDNSEnabled = true
	}

	// Security Validation for non-localhost bindings
	isLocalhost := cfg.BindAddress == "127.0.0.1" || cfg.BindAddress == "localhost"
	isSafeMode := cfg.AuthEnabled || cfg.ReadOnly

	if !isLocalhost {
		if !isSafeMode && !allowUnsafe {
			return fmt.Errorf("SECURITY ERROR: Binding to %s exposes LGH to the network without protection.\n"+
				"To proceed, you must enable Authentication, Read-Only mode, or use --allow-unsafe.", cfg.BindAddress)
		}

		if allowUnsafe && !isSafeMode {
			ui.Warning("⚠️  RUNNING IN UNSAFE MODE: External access enabled without Auth or Read-Only!")
		} else {
			ui.Info("ℹ️  Running in networked mode (Protected by Auth/ReadOnly)")
		}
	}

	// Check if already running
	if running, pid := server.IsRunning(); running {
		ui.Warning("LGH server is already running (PID: %d)", pid)
		ui.Info("Use 'lgh status' to check the server status")
		return nil
	}

	// Daemon mode: start server in background
	if daemonFlag {
		return startDaemon(cfg)
	}

	// Start mDNS if enabled
	if cfg.MDNSEnabled {
		mdnsService, err := mdns.NewService(cfg.Port)
		if err != nil {
			ui.Warning("Failed to initialize mDNS: %v", err)
		} else {
			if err := mdnsService.Start(); err != nil {
				ui.Warning("Failed to start mDNS: %v", err)
			} else {
				ui.Success("mDNS enabled: %s", mdnsService.GetServiceURL())
			}
		}
	}

	// Create and start server using cfg.ReadOnly (respects config.yaml)
	srv := server.New(cfg)
	return srv.Start()
}

func startDaemon(cfg *config.Config) error {
	// Build command arguments
	args := []string{"serve"}
	if cfg.Port != 9418 {
		args = append(args, "--port", fmt.Sprintf("%d", cfg.Port))
	}
	if cfg.BindAddress != "127.0.0.1" && cfg.BindAddress != "" {
		args = append(args, "--bind", cfg.BindAddress)
	}
	if cfg.ReadOnly {
		args = append(args, "--read-only")
	}
	if cfg.MDNSEnabled {
		args = append(args, "--mdns")
	}

	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Start process in background
	// #nosec G204 -- executable is our own binary path, args are trusted
	cmd := exec.Command(executable, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Detach from parent process
	cmd.SysProcAttr = server.GetDaemonSysProcAttr()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	ui.Success("LGH server started in background (PID: %d)", cmd.Process.Pid)
	ui.Info("Address: http://%s:%d", cfg.BindAddress, cfg.Port)
	ui.Info("Use 'lgh stop' to stop the server")
	ui.Info("Use 'lgh status' to check server status")

	return nil
}
