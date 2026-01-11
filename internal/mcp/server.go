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

// Package mcp provides Model Context Protocol server implementation for LGH
package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Version is the MCP server version
const Version = "1.2.0"

// NewServer creates and configures the LGH MCP server
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"lgh",
		Version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	// Register tools
	registerTools(s)

	// Register resources
	registerResources(s)

	return s
}

// registerTools registers all LGH tools with the MCP server
func registerTools(s *server.MCPServer) {
	// lgh_status - Get server status
	s.AddTool(
		mcp.NewTool("lgh_status",
			mcp.WithDescription("Get LGH server status including running state and repository list"),
		),
		handleStatus,
	)

	// lgh_list - List repositories
	s.AddTool(
		mcp.NewTool("lgh_list",
			mcp.WithDescription("List all repositories registered with LGH local server. Returns source_path (local working dir) and clone_url (LGH server URL)."),
		),
		handleList,
	)

	// lgh_add - Add repository
	s.AddTool(
		mcp.NewTool("lgh_add",
			mcp.WithDescription("Register a local Git repository with LGH local server. This creates a bare repo on LGH and adds 'lgh' remote to the source repo."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Absolute path to the LOCAL working directory containing the Git repository"),
			),
			mcp.WithString("name",
				mcp.Description("Optional custom name for the repository on LGH"),
			),
		),
		handleAdd,
	)

	// lgh_remove - Remove repository
	s.AddTool(
		mcp.NewTool("lgh_remove",
			mcp.WithDescription("Remove a repository from LGH"),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name of the repository to remove"),
			),
		),
		handleRemove,
	)

	// lgh_up - One-click commit and push
	s.AddTool(
		mcp.NewTool("lgh_up",
			mcp.WithDescription("One-click backup: auto .gitignore + git add + git commit + git push to LGH local server. NOT GitHub/GitLab - this pushes to localhost LGH."),
			mcp.WithString("message",
				mcp.Required(),
				mcp.Description("Git commit message"),
			),
			mcp.WithString("path",
				mcp.Description("Absolute path to the LOCAL working directory (defaults to current directory)"),
			),
			mcp.WithBoolean("force",
				mcp.Description("Skip trash detection (large files, .env) and force push"),
			),
		),
		handleUp,
	)

	// lgh_save - Local save
	s.AddTool(
		mcp.NewTool("lgh_save",
			mcp.WithDescription("Local save only: git add + git commit WITHOUT push. Changes stay in local working directory, not synced to LGH server."),
			mcp.WithString("message",
				mcp.Required(),
				mcp.Description("Git commit message"),
			),
			mcp.WithString("path",
				mcp.Description("Absolute path to the LOCAL working directory (defaults to current directory)"),
			),
		),
		handleSave,
	)

	// lgh_serve_start - Start server
	s.AddTool(
		mcp.NewTool("lgh_serve_start",
			mcp.WithDescription("Start the LGH HTTP server in background"),
			mcp.WithNumber("port",
				mcp.Description("Port to listen on (default: 9418)"),
			),
		),
		handleServeStart,
	)

	// lgh_serve_stop - Stop server
	s.AddTool(
		mcp.NewTool("lgh_serve_stop",
			mcp.WithDescription("Stop the LGH HTTP server"),
		),
		handleServeStop,
	)

	// lgh_log - View server logs
	s.AddTool(
		mcp.NewTool("lgh_log",
			mcp.WithDescription("View LGH server runtime logs (errors, warnings, info)"),
			mcp.WithNumber("limit",
				mcp.Description("Number of log entries to return (default: 20)"),
			),
			mcp.WithString("level",
				mcp.Description("Filter by log level (DEBUG, INFO, WARN, ERROR)"),
			),
		),
		handleLog,
	)
}

// registerResources registers all LGH resources with the MCP server
func registerResources(s *server.MCPServer) {
	// lgh://config - Current configuration
	s.AddResource(
		mcp.NewResource("lgh://config",
			"LGH Configuration",
			mcp.WithResourceDescription("Current LGH configuration settings"),
			mcp.WithMIMEType("application/json"),
		),
		handleResourceConfig,
	)

	// lgh://repos - Repository list
	s.AddResource(
		mcp.NewResource("lgh://repos",
			"Repository List",
			mcp.WithResourceDescription("List of all repositories registered with LGH"),
			mcp.WithMIMEType("application/json"),
		),
		handleResourceRepos,
	)

	// lgh://server/status - Server status
	s.AddResource(
		mcp.NewResource("lgh://server/status",
			"Server Status",
			mcp.WithResourceDescription("Current LGH server status"),
			mcp.WithMIMEType("application/json"),
		),
		handleResourceServerStatus,
	)
}
