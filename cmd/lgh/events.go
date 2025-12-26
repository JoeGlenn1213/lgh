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
	eventsLimit int
	eventsWatch bool
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "View system events log",
	Long:  `View the recent activity log (push events, repo management, etc.).`,
	RunE:  runEvents,
}

func init() {
	eventsCmd.Flags().IntVarP(&eventsLimit, "limit", "n", 20, "Number of events to show")
	eventsCmd.Flags().BoolVarP(&eventsWatch, "watch", "w", false, "Watch for new events (tail -f)")
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
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	for i := start; i < len(lines); i++ {
		printEventLine(lines[i])
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
