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
	"encoding/json"
	"net"
	"os"
	"path/filepath"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/event"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

// startIPC starts the Unix Domain Socket listener for event subscription
func (s *Server) startIPC() {
	cfg := config.Get()
	sockPath := filepath.Join(cfg.DataDir, "lgh.sock")

	// Cleanup old socket
	if _, err := os.Stat(sockPath); err == nil {
		if err := os.Remove(sockPath); err != nil {
			ui.Error("Failed to remove old socket: %v", err)
			return
		}
	}

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		ui.Error("Failed to start IPC listener: %v", err)
		return
	}

	// Restrict permissions so only the current user can access
	if err := os.Chmod(sockPath, 0600); err != nil {
		ui.Warning("Failed to set socket permissions: %v", err)
	}

	ui.Info("IPC listener started at %s", sockPath)

	go func() {
		defer listener.Close()
		defer os.Remove(sockPath)

		for {
			conn, err := listener.Accept()
			if err != nil {
				// Prevent spamming logs on shutdown
				return
			}
			go handleIPCConnection(conn)
		}
	}()
}

func handleIPCConnection(conn net.Conn) {
	defer conn.Close()

	// 1. Subscribe to broker (Server -> Client)
	ch := event.SubscribeClient()
	defer event.UnsubscribeClient(ch)

	// 2. Start Writer (Server -> Client)
	// We run this in the main goroutine to keep the handler alive until disconnect
	encoder := json.NewEncoder(conn)
	for evt := range ch {
		if err := encoder.Encode(evt); err != nil {
			return // Client disconnected or error
		}
	}
}
