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

	"github.com/JoeGlenn1213/lgh/internal/server"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running LGH server",
	Long: `Stop the LGH server that is running in the background.

This command will gracefully terminate the LGH server process.

Examples:
  lgh stop   # Stop the running server`,
	RunE: runStop,
}

func runStop(_ *cobra.Command, _ []string) error {
	running, pid := server.IsRunning()
	if !running {
		ui.Info("LGH server is not running")
		return nil
	}

	ui.Info("Stopping LGH server (PID: %d)...", pid)

	if err := server.StopServer(pid); err != nil {
		return err
	}

	ui.Success("LGH server stopped successfully")
	return nil
}
