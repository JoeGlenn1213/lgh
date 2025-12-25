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

//go:build darwin || linux
// +build darwin linux

// Package server implements the LGH server and process management
package server

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// checkProcessRunning checks if the process with the given PID is running and is likely an LGH server
func checkProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, sending signal 0 checks if process exists
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	// SECURITY: Verify the process is actually an LGH server (handles PID reuse)
	return isLGHProcess(pid)
}

// isLGHProcess checks if the given PID is an LGH process using ps
func isLGHProcess(pid int) bool {
	// nolint:gosec // G204: Subprocess launched with a potential tainted input. pid is a trusted integer.
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}

	procName := strings.TrimSpace(out.String())
	return strings.Contains(procName, "lgh")
}
