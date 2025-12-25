//go:build darwin || linux
// +build darwin linux

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
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}

	procName := strings.TrimSpace(out.String())
	return strings.Contains(procName, "lgh")
}
