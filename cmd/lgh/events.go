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
	"github.com/JoeGlenn1213/lgh/internal/event"
	"github.com/JoeGlenn1213/lgh/pkg/ui"
)

var (
	eventsLimit  int
	eventsWatch  bool
	eventsFilter string
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "View system events log",
	Long: `View the recent activity log.
	
Supports filtering by event type (--type) and real-time watching (--watch).
Optimization: Reads from the end of file for fast access to large logs.`,
	RunE: runEvents,
}

func init() {
	eventsCmd.Flags().IntVarP(&eventsLimit, "limit", "n", 20, "Number of events to show")
	eventsCmd.Flags().BoolVarP(&eventsWatch, "watch", "w", false, "Watch for new events (tail -f)")
	eventsCmd.Flags().StringVar(&eventsFilter, "type", "", "Filter events by type (e.g. git.push, repo.added)")
}

func runEvents(_ *cobra.Command, _ []string) error {
	// Ensure config loaded
	if _, err := config.Load(); err != nil {
		return err
	}
	cfg := config.Get()
	logPath := filepath.Join(cfg.DataDir, "events", "events.jsonl")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		ui.Warning("No events log found at %s", logPath)
		return nil
	}

	if eventsWatch {
		return watchEvents(logPath)
	}

	return showRecentEvents(logPath, eventsLimit)
}

func showRecentEvents(path string, n int) error {
	// nolint:gosec // G304: path is internally constructed from config
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Optimization: If file is larger than 2MB, read only last 2MB for recent events
	// This helps with performance on large logs.
	const maxReadSize = 2 * 1024 * 1024
	var offset int64

	fi, err := file.Stat()
	if err == nil && fi.Size() > maxReadSize {
		offset = fi.Size() - maxReadSize
		if _, err := file.Seek(offset, 0); err != nil {
			return err
		}
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	// Increase buffer size
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	// If we seeked, the first line might be partial, discard it
	if offset > 0 {
		scanner.Scan()
	}

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Filter and select last N in reverse
	var output []string
	count := 0

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if eventsFilter != "" {
			// Fast check string contains to avoid json parse?
			// No, safer to parse or just check matches since type is usually near start.
			// Let's do a simple string matching first for performance logic?
			// "type":"git.push"
			if !strings.Contains(line, fmt.Sprintf(`"type":%q`, eventsFilter)) &&
				!strings.Contains(line, fmt.Sprintf(`"type":"%s"`, eventsFilter)) {
				// Handle json spacing variation? Basic check is "type":"VALUE"
				// Safest is JSON decode.
				var evt event.Event
				if err := json.Unmarshal([]byte(line), &evt); err == nil {
					if string(evt.Type) != eventsFilter {
						continue
					}
				} else {
					continue // skip broken lines
				}
			}
		}

		output = append([]string{line}, output...)
		count++
		if count >= n {
			break
		}
	}

	for _, line := range output {
		printEventLine(line)
	}
	return nil
}

func printEventLine(line string) {
	if line == "" {
		return
	}
	var evt event.Event
	if err := json.Unmarshal([]byte(line), &evt); err != nil {
		// print raw if parse fail
		fmt.Println(line)
		return
	}

	// Pretty print
	// Timestamp | TYPE | Repo | Detail
	ts := evt.Timestamp.Format("15:04:05")

	var typeColor func(a ...interface{}) string
	switch evt.Type {
	case event.GitPush:
		typeColor = ui.Green
	case event.RepoAdded:
		typeColor = ui.Cyan
	case event.RepoRemoved:
		typeColor = ui.Red
	default:
		typeColor = ui.Gray
	}

	payloadStr := ""
	if evt.Type == event.GitPush {
		if changes, ok := evt.Payload["changes"].(map[string]interface{}); ok {
			var refs []string
			for ref, val := range changes {
				changeMap, _ := val.(map[string]interface{})
				action, _ := changeMap["action"].(string)
				newHash, _ := changeMap["new"].(string)

				shortRef := strings.TrimPrefix(ref, "refs/heads/")
				if shortRef == "" {
					shortRef = ref // fallback
				}

				// Show symbols
				symbol := "~"
				if action == "created" {
					symbol = "+"
				}
				if action == "deleted" {
					symbol = "-"
				}

				label := fmt.Sprintf("%s%s", symbol, shortRef)

				// Append hash if not deleted
				if action != "deleted" && len(newHash) >= 7 {
					shortHash := newHash[:7]
					// Also check if not zero hash (though action check covers it usually)
					if shortHash != "0000000" {
						label += fmt.Sprintf(":%s", shortHash)
					}
				}

				refs = append(refs, label)
			}
			payloadStr = strings.Join(refs, ", ")
		}
	} else if evt.Type == event.RepoAdded {
		if bare, ok := evt.Payload["bare"].(string); ok {
			payloadStr = filepath.Base(bare)
		}
	}

	fmt.Printf("%s  %-12s  %-15s  %s\n",
		ui.Gray(ts),
		typeColor(evt.Type),
		evt.RepoName,
		payloadStr,
	)
}

func watchEvents(path string) error {
	// nolint:gosec // G304: path is trusted
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Seek to end
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	ui.Info("Watching for events... (Ctrl+C to stop)")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return err
		}
		printEventLine(strings.TrimSpace(line))
	}
}
