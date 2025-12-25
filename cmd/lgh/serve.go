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
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the LGH HTTP server",
	Long: `Start the LGH HTTP server to serve Git repositories over HTTP.

The server listens on localhost by default (127.0.0.1:9418).
Use --read-only to prevent push operations.
Use --mdns to enable mDNS for local network discovery.

Examples:
  lgh serve                    # Start with defaults
  lgh serve --read-only        # Start in read-only mode
  lgh serve --port 8080        # Use custom port
  lgh serve --mdns             # Enable mDNS discovery`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().BoolVarP(&readOnlyFlag, "read-only", "r", false, "Enable read-only mode (disable push)")
	serveCmd.Flags().BoolVarP(&enableMDNS, "mdns", "m", false, "Enable mDNS for LAN discovery")
	serveCmd.Flags().IntVarP(&serverPort, "port", "p", 0, "Port to listen on (default: 9418)")
	serveCmd.Flags().StringVarP(&bindAddress, "bind", "b", "", "Address to bind to (default: 127.0.0.1)")
}

func runServe(_ *cobra.Command, args []string) error {
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

	// Security warning for non-localhost bindings
	if cfg.BindAddress != "127.0.0.1" && cfg.BindAddress != "localhost" {
		ui.Warning("⚠️  WARNING: Binding to %s exposes LGH to the network!", cfg.BindAddress)
		if !cfg.ReadOnly {
			ui.Warning("⚠️  Consider using --read-only to prevent unauthorized push access")
		}
	}

	// Check if already running
	if running, pid := server.IsRunning(); running {
		ui.Warning("LGH server is already running (PID: %d)", pid)
		ui.Info("Use 'lgh status' to check the server status")
		return nil
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
