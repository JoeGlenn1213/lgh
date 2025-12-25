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

	"github.com/spf13/cobra"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/tunnel"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var (
	tunnelMethod string
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Expose LGH to the internet",
	Long: `Expose your local LGH server to the internet for remote access.

This command helps you set up a tunnel to expose your local Git server
to external collaborators or CI/CD systems.

Supported methods:
  • ssh          SSH reverse tunnel (requires remote server)
  • ngrok        ngrok tunnel (requires ngrok installed)
  • cloudflared  Cloudflare tunnel (requires cloudflared installed)
  • localtunnel  localtunnel.me (requires lt installed)

Examples:
  lgh tunnel                   # Show instructions
  lgh tunnel --method ngrok    # Start ngrok tunnel
  lgh tunnel --method cloudflared # Start Cloudflare tunnel`,
	RunE: runTunnel,
}

func init() {
	tunnelCmd.Flags().StringVarP(&tunnelMethod, "method", "m", "", "Tunnel method (ngrok, cloudflared, ssh)")
}

func runTunnel(_ *cobra.Command, _ []string) error {
	// Ensure initialized
	if err := ensureInitialized(); err != nil {
		return err
	}

	cfg := config.Get()
	t := tunnel.NewTunnel(cfg.Port)

	// Show available methods
	methods := tunnel.AvailableTunnelMethods()
	ui.Title("LGH Tunnel")
	ui.Info("Available tunnel methods: %s", tunnel.FormatMethods(methods))
	fmt.Println()

	// If no method specified, show instructions
	if tunnelMethod == "" {
		t.PrintInstructions()
		return nil
	}

	// Handle specific method
	switch tunnelMethod {
	case "ngrok":
		if !tunnel.CheckNgrok() {
			ui.Error("ngrok is not installed")
			ui.Info("Install with: %s", tunnel.GetInstallCommand("ngrok"))
			return fmt.Errorf("ngrok not found")
		}

		ui.Info("Starting ngrok tunnel on port %d...", cfg.Port)
		ui.Warning("Press Ctrl+C to stop the tunnel")
		fmt.Println()

		proc, err := tunnel.StartNgrok(cfg.Port)
		if err != nil {
			return err
		}

		// Wait for process
		return proc.Wait()

	case "cloudflared":
		if !tunnel.CheckCloudflared() {
			ui.Error("cloudflared is not installed")
			ui.Info("Install with: %s", tunnel.GetInstallCommand("cloudflared"))
			return fmt.Errorf("cloudflared not found")
		}

		ui.Info("Starting Cloudflare tunnel on port %d...", cfg.Port)
		ui.Warning("Press Ctrl+C to stop the tunnel")
		fmt.Println()

		proc, err := tunnel.StartCloudflared(cfg.Port)
		if err != nil {
			return err
		}

		return proc.Wait()

	case "ssh":
		ui.Info("SSH Reverse Tunnel")
		fmt.Println()
		ui.Info("Run this command to create a tunnel:")
		fmt.Println()
		ui.Command(t.GetSSHCommand())
		fmt.Println()
		ui.Info("Replace <remote_port>, <user>, and <remote-host> with your values.")
		return nil

	default:
		return fmt.Errorf("unknown tunnel method: %s. Available: ngrok, cloudflared, ssh", tunnelMethod)
	}
}
