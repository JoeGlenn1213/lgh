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

// Package tunnel provides remote access via ngrok and cloudflared
package tunnel

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Tunnel provides functionality to expose local LGH server to the internet
type Tunnel struct {
	localPort  int
	remoteHost string
	remotePort int
	sshUser    string
}

// NewTunnel creates a new tunnel configuration
func NewTunnel(localPort int) *Tunnel {
	return &Tunnel{
		localPort:  localPort,
		remotePort: 0, // Will be auto-assigned or specified
	}
}

// SetRemote sets the remote SSH server for tunneling
func (t *Tunnel) SetRemote(user, host string, port int) {
	t.sshUser = user
	t.remoteHost = host
	t.remotePort = port
}

// GetSSHCommand returns the SSH command for manual tunneling
func (t *Tunnel) GetSSHCommand() string {
	// Generate a reverse tunnel command
	// ssh -R <remote_port>:localhost:<local_port> user@remote-host
	if t.remoteHost == "" {
		return fmt.Sprintf("ssh -R <remote_port>:localhost:%d <user>@<remote-host>", t.localPort)
	}

	remotePort := t.remotePort
	if remotePort == 0 {
		remotePort = t.localPort
	}

	return fmt.Sprintf("ssh -R %d:localhost:%d %s@%s",
		remotePort, t.localPort, t.sshUser, t.remoteHost)
}

// PrintInstructions prints instructions for setting up a tunnel
func (t *Tunnel) PrintInstructions() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    LGH Tunnel Instructions                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("To expose your local LGH server to the internet, you have several options:")
	fmt.Println()
	fmt.Println("┌──────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Option 1: SSH Reverse Tunnel (requires remote server)           │")
	fmt.Println("└──────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Printf("  Run this command on your local machine:\n")
	fmt.Println()
	fmt.Printf("    %s\n", t.GetSSHCommand())
	fmt.Println()
	fmt.Println("  Replace <remote_port>, <user>, and <remote-host> with your values.")
	fmt.Println()
	fmt.Println("┌──────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Option 2: Use ngrok (recommended for quick sharing)             │")
	fmt.Println("└──────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  1. Install ngrok: https://ngrok.com/download")
	fmt.Printf("  2. Run: ngrok http %d\n", t.localPort)
	fmt.Println("  3. Share the generated URL with your collaborators")
	fmt.Println()
	fmt.Println("┌──────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Option 3: Use Cloudflare Tunnel (free, stable)                  │")
	fmt.Println("└──────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  1. Install cloudflared: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation")
	fmt.Printf("  2. Run: cloudflared tunnel --url http://localhost:%d\n", t.localPort)
	fmt.Println()
	fmt.Println("┌──────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Option 4: Use localtunnel (npm package)                         │")
	fmt.Println("└──────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  1. Install: npm install -g localtunnel")
	fmt.Printf("  2. Run: lt --port %d\n", t.localPort)
	fmt.Println()
}

// CheckNgrok checks if ngrok is installed
func CheckNgrok() bool {
	_, err := exec.LookPath("ngrok")
	return err == nil
}

// CheckCloudflared checks if cloudflared is installed
func CheckCloudflared() bool {
	_, err := exec.LookPath("cloudflared")
	return err == nil
}

// StartNgrok starts an ngrok tunnel and returns the process
func StartNgrok(port int) (*exec.Cmd, error) {
	if !CheckNgrok() {
		return nil, fmt.Errorf("ngrok is not installed")
	}

	// nolint:gosec // G204: Subprocess launched with variable. port is a trusted integer.
	cmd := exec.Command("ngrok", "http", fmt.Sprintf("%d", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ngrok: %w", err)
	}

	return cmd, nil
}

// StartCloudflared starts a cloudflared tunnel
func StartCloudflared(port int) (*exec.Cmd, error) {
	if !CheckCloudflared() {
		return nil, fmt.Errorf("cloudflared is not installed")
	}

	// nolint:gosec // G204: Subprocess launched with variable. port is a trusted integer.
	cmd := exec.Command("cloudflared", "tunnel", "--url", fmt.Sprintf("http://localhost:%d", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cloudflared: %w", err)
	}

	return cmd, nil
}

// GetInstallCommand returns the install command for the current OS
func GetInstallCommand(tool string) string {
	os := runtime.GOOS

	switch tool {
	case "ngrok":
		switch os {
		case "darwin":
			return "brew install ngrok"
		case "linux":
			return "snap install ngrok"
		default:
			return "Download from https://ngrok.com/download"
		}
	case "cloudflared":
		switch os {
		case "darwin":
			return "brew install cloudflared"
		case "linux":
			return "See https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation"
		default:
			return "Download from Cloudflare website"
		}
	}

	return ""
}

// AvailableTunnelMethods returns available tunnel methods on this system
func AvailableTunnelMethods() []string {
	methods := []string{"ssh"}

	if CheckNgrok() {
		methods = append(methods, "ngrok")
	}
	if CheckCloudflared() {
		methods = append(methods, "cloudflared")
	}

	// Check for localtunnel
	if _, err := exec.LookPath("lt"); err == nil {
		methods = append(methods, "localtunnel")
	}

	return methods
}

// FormatMethods formats available methods as a string
func FormatMethods(methods []string) string {
	return strings.Join(methods, ", ")
}
