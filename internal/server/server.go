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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/event"
	"github.com/JoeGlenn1213/lgh/internal/git"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

// Server represents the LGH HTTP server
type Server struct {
	cfg        *config.Config
	httpServer *http.Server
}

// New creates a new LGH server instance
// ReadOnly mode is now taken from cfg.ReadOnly for consistency
func New(cfg *config.Config) *Server {
	return &Server{
		cfg: cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create Git backend handler using cfg.ReadOnly
	gitHandler, err := git.CreateHandler(s.cfg.ReposDir, s.cfg.ReadOnly)
	if err != nil {
		return fmt.Errorf("failed to create git handler: %w", err)
	}

	// Build handler chain
	var handler = gitHandler

	// Add virtual owner middleware (to support /owner/repo.git paths)
	handler = s.virtualOwnerMiddleware(handler)

	// Add logging middleware
	handler = s.loggingMiddleware(handler)

	// Add authentication middleware if enabled
	if s.cfg.AuthEnabled && s.cfg.AuthUser != "" && s.cfg.AuthPasswordHash != "" {
		authMiddleware := NewAuthMiddleware(s.cfg.AuthUser, s.cfg.AuthPasswordHash)
		handler = authMiddleware.Wrap(handler)
		ui.Success("Authentication enabled (user: %s)", s.cfg.AuthUser)
	}

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoint (always public, bypasses auth)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Debug Event Injection (v1.1.0)
	// Allows replaying/injecting events via HTTP (Localhost only recommended)
	// Takes a JSON event body and broadcasts it via the Broker.
	mux.HandleFunc("/debug/events", s.handleDebugEvents)

	// Git backend for all .git paths
	mux.Handle("/", handler)

	// Create server with security hardening
	addr := fmt.Sprintf("%s:%d", s.cfg.BindAddress, s.cfg.Port)
	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       30 * time.Minute, // Long timeout for large pushes
		WriteTimeout:      30 * time.Minute,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second, // SECURITY: Prevent slowloris attacks
		MaxHeaderBytes:    1 << 20,          // SECURITY: 1MB max header size
	}

	// Check if port is available
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use: %w", s.cfg.Port, err)
	}
	_ = ln.Close()

	// Save PID file
	if err := s.savePID(); err != nil {
		ui.Warning("Failed to save PID file: %v", err)
	}

	// Setup graceful shutdown
	go s.handleShutdown()

	// Display server info
	s.displayStartupInfo()

	// Initialize Event Broker
	event.StartBroker()

	// Start IPC Listener (v1.1.0)
	s.startIPC()

	// Start server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Remove PID file
	_ = os.Remove(config.GetPIDPath())

	return s.httpServer.Shutdown(ctx)
}

// savePID saves the current process ID to a file
func (s *Server) savePID() error {
	pidPath := config.GetPIDPath()
	return os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0600)
}

// handleShutdown handles graceful shutdown signals
func (s *Server) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	ui.Info("\nShutting down server...")

	if err := s.Stop(); err != nil {
		ui.Error("Error during shutdown: %v", err)
	} else {
		ui.Success("Server stopped gracefully")
	}

	// Flush events
	event.Shutdown()

	os.Exit(0)
}

// displayStartupInfo displays server startup information
func (s *Server) displayStartupInfo() {
	ui.Success("LGH Server started successfully!")
	fmt.Println()
	ui.Info("  Address:   http://%s:%d", s.cfg.BindAddress, s.cfg.Port)
	ui.Info("  Repos Dir: %s", s.cfg.ReposDir)

	if s.cfg.ReadOnly {
		ui.Warning("  Mode:      READ-ONLY (push disabled)")
	} else {
		ui.Info("  Mode:      Read/Write")
	}

	fmt.Println()
	ui.Info("Press Ctrl+C to stop the server")
	fmt.Println()
}

// loggingMiddleware logs incoming requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Log request (only for git operations, not health checks)
		if r.URL.Path != "/health" {
			statusColor := ui.Green
			if rw.statusCode >= 400 {
				statusColor = ui.Red
			} else if rw.statusCode >= 300 {
				statusColor = ui.Yellow
			}

			fmt.Printf("%s %s %s %s %s\n",
				ui.Gray(r.Method),
				r.URL.Path,
				statusColor(fmt.Sprintf("%d", rw.statusCode)),
				ui.Gray(duration.String()),
				ui.Gray(r.RemoteAddr),
			)
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// IsRunning checks if the server is running by checking PID file
// Fixed: Uses platform-specific checkProcessRunning to handle PID reuse and existence check
func IsRunning() (bool, int) {
	pidPath := config.GetPIDPath()
	// nolint:gosec // G304: Potential file inclusion via variable. pidPath is a trusted path from config.
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false, 0
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		_ = os.Remove(pidPath)
		return false, 0
	}

	if !checkProcessRunning(pid) {
		_ = os.Remove(pidPath)
		return false, 0
	}

	return true, pid
}

// GetServerURL returns the server URL
func GetServerURL() string {
	cfg := config.Get()
	return fmt.Sprintf("http://%s:%d", cfg.BindAddress, cfg.Port)
}

// virtualOwnerMiddleware strips the first path segment if it looks like an owner
// e.g. /owner/repo.git -> /repo.git
// This allows tools that strictly enforce owner/repo structure to work with flat LGH repos.
func (s *Server) virtualOwnerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

		// Find segment ending in .git
		gitIdx := -1
		for i, p := range parts {
			if strings.HasSuffix(p, ".git") {
				gitIdx = i
				break
			}
		}

		// If .git is the second segment (index 1), and index 0 is "lgh"
		// We strict this to "lgh" to avoid security confusion or arbitrary path support.
		if gitIdx == 1 && parts[0] == "lgh" {
			// Reconstruct path starting from repo name
			// /lgh/repo.git/info/refs -> /repo.git/info/refs
			newPath := "/" + strings.Join(parts[1:], "/")

			// We modify the request path context effectively
			r.URL.Path = newPath
		}

		next.ServeHTTP(w, r)
	})
}

// handleDebugEvents handles event injection via HTTP
func (s *Server) handleDebugEvents(w http.ResponseWriter, r *http.Request) {
	// Security: Only allow POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Security: Strict Localhost Enforcement
	// We only allow local processes to inject debug events.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if host != "127.0.0.1" && host != "::1" {
		ui.Warning("Blocked external attempt to access /debug/events from %s", host)
		http.Error(w, "Forbidden: Localhost only", http.StatusForbidden)
		return
	}

	var evt event.Event
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Tag the event as replayed
	if evt.Payload == nil {
		evt.Payload = make(map[string]interface{})
	}
	evt.Payload["_replayed"] = true

	// Broadcast
	event.Broadcast(evt)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
