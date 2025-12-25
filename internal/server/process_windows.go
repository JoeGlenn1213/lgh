//go:build windows
// +build windows

package server

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

// checkProcessRunning checks if the process with the given PID is running and is likely an LGH server
func checkProcessRunning(pid int) bool {
	// On Windows, use tasklist to verify PID and image name
	// tasklist /FI "PID eq <pid>" /NH
	cmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/NH")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}

	output := out.String()
	// If process does not exist, tasklist usually outputs "INFO: No tasks are running"
	if strings.Contains(output, "No tasks") {
		return false
	}

	// Double check if the output contains "lgh" to handle PID reuse
	// Windows filenames are case-insensitive
	return strings.Contains(strings.ToLower(output), "lgh")
}
