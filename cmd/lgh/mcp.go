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

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/JoeGlenn1213/lgh/internal/config"
	lghMcp "github.com/JoeGlenn1213/lgh/internal/mcp"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var (
	mcpMode string
	mcpPort int
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server for AI integration",
	Long: `Start the LGH MCP server for AI agent integration.

MCP (Model Context Protocol) allows AI tools like Cursor, Claude Desktop, 
and other agents to interact with LGH programmatically.

Modes:
  stdio - Standard I/O mode (for local AI clients)
  sse   - Server-Sent Events mode (for web-based AI) [coming soon]

Tools available via MCP:
  - lgh_status: Get server status
  - lgh_list: List repositories
  - lgh_add: Add repository
  - lgh_remove: Remove repository
  - lgh_up: One-click commit and push
  - lgh_save: Local save
  - lgh_serve_start/stop: Server control
  - lgh_log: View server logs

Resources:
  - lgh://config: Current configuration
  - lgh://repos: Repository list
  - lgh://server/status: Server status`,
	Example: `  # Start MCP server in stdio mode (for Cursor, Claude Desktop)
  lgh mcp

  # Or explicitly specify mode
  lgh mcp --mode stdio

  # Test with JSON-RPC
  echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | lgh mcp`,
	RunE: runMcp,
}

func init() {
	mcpCmd.Flags().StringVar(&mcpMode, "mode", "stdio", "Transport mode: stdio (default) or sse")
	mcpCmd.Flags().IntVar(&mcpPort, "port", 9419, "Port for SSE mode (default: 9419)")
	rootCmd.AddCommand(mcpCmd)
}

func runMcp(_ *cobra.Command, _ []string) error {
	// Load config
	if _, err := config.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch mcpMode {
	case "stdio":
		return runStdioMode()
	case "sse":
		return fmt.Errorf("SSE mode is not yet implemented")
	default:
		return fmt.Errorf("unknown mode: %s (use 'stdio' or 'sse')", mcpMode)
	}
}

func runStdioMode() error {
	mcpServer := lghMcp.NewServer()

	// Write startup message to stderr (not stdout, to avoid interfering with JSON-RPC)
	fmt.Fprintln(os.Stderr, ui.Green("LGH MCP Server started (stdio mode)"))
	fmt.Fprintln(os.Stderr, "Ready to accept JSON-RPC requests...")

	// Start the server with stdio transport
	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}
