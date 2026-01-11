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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"github.com/JoeGlenn1213/lgh/internal/slog"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var (
	logLimit int
	logWatch bool
	logLevel string
	logJSON  bool
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "View LGH server logs (errors, warnings, info)",
	Long: `View the LGH service runtime logs.

Unlike 'lgh events' which shows system events (git.push, repo.added),
'lgh log' shows server runtime logs including errors and warnings.

This is useful for:
  - Debugging server issues
  - Monitoring service health  
  - AI agents to detect anomalies`,
	Example: `  # Show recent logs
  lgh log

  # Show only errors
  lgh log --level ERROR

  # Watch logs in real-time
  lgh log --watch

  # Output as JSON (for AI/MCP)
  lgh log --json`,
	RunE: runLog,
}

func init() {
	logCmd.Flags().IntVarP(&logLimit, "limit", "n", 50, "Number of log entries to show")
	logCmd.Flags().BoolVarP(&logWatch, "watch", "w", false, "Watch for new logs (tail -f)")
	logCmd.Flags().StringVar(&logLevel, "level", "", "Filter by level (DEBUG, INFO, WARN, ERROR)")
	logCmd.Flags().BoolVar(&logJSON, "json", false, "Output as JSON (for AI/MCP integration)")
	rootCmd.AddCommand(logCmd)
}

func runLog(_ *cobra.Command, _ []string) error {
	// Ensure config loaded
	if _, err := config.Load(); err != nil {
		return err
	}
	cfg := config.Get()
	logPath := filepath.Join(cfg.DataDir, "logs", "server.jsonl")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		ui.Warning("No server logs found at %s", logPath)
		ui.Info("Server logs are created when 'lgh serve' runs")
		return nil
	}

	if logWatch {
		return watchLogs(logPath)
	}

	return showRecentLogs(logPath, logLimit)
}

func showRecentLogs(path string, n int) error {
	lines, err := slog.ReadLastLines(path, n, logLevel)
	if err != nil {
		return err
	}

	// Print output
	if logJSON {
		// Output as JSON array for AI consumption
		fmt.Println("[")
		for i, line := range lines {
			fmt.Print("  ", line)
			if i < len(lines)-1 {
				fmt.Println(",")
			} else {
				fmt.Println()
			}
		}
		fmt.Println("]")
	} else {
		for _, line := range lines {
			printLogLine(line)
		}
	}

	return nil
}

func printLogLine(line string) {
	if line == "" {
		return
	}

	var entry slog.Entry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		fmt.Println(line)
		return
	}

	ts := entry.Timestamp.Format("01-02 15:04:05")

	var levelColor func(a ...interface{}) string
	switch entry.Level {
	case slog.LevelError, slog.LevelFatal:
		levelColor = ui.Red
	case slog.LevelWarn:
		levelColor = ui.Yellow
	case slog.LevelInfo:
		levelColor = ui.Cyan
	case slog.LevelDebug:
		levelColor = ui.Gray
	default:
		levelColor = ui.Gray
	}

	component := ""
	if entry.Component != "" {
		component = fmt.Sprintf("[%s] ", entry.Component)
	}

	fmt.Printf("%s  %-5s  %s%s\n",
		ui.Gray(ts),
		levelColor(entry.Level),
		component,
		entry.Message,
	)

	// Print fields if any
	if len(entry.Fields) > 0 {
		for k, v := range entry.Fields {
			fmt.Printf("         %s=%v\n", ui.Gray(k), v)
		}
	}
}

func watchLogs(path string) error {
	// nolint:gosec // G304: path is trusted
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	ui.Info("Watching server logs... (Ctrl+C to stop)")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return err
		}

		line = strings.TrimSpace(line)
		if logLevel != "" {
			var entry slog.Entry
			if err := json.Unmarshal([]byte(line), &entry); err == nil {
				if string(entry.Level) != strings.ToUpper(logLevel) {
					continue
				}
			}
		}

		if logJSON {
			fmt.Println(line)
		} else {
			printLogLine(line)
		}
	}
}
