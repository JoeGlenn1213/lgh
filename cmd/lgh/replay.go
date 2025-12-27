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
	"net/http"
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
	replayLast int
	replayType string
)

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay historical events to listeners",
	Long: `Read past events from the log and re-broadcast them to connected listeners (Agents).
	
Events are injected via the local server's debug API and will be tagged with {"_replayed": true}.
Useful for testing integrations and debugging event-driven workflows without performing real Git actions.`,
	RunE: runReplay,
}

func init() {
	eventsCmd.AddCommand(replayCmd)
	replayCmd.Flags().IntVarP(&replayLast, "last", "n", 10, "Number of recent events to replay")
	replayCmd.Flags().StringVar(&replayType, "type", "", "Filter events by type (e.g. git.push)")
}

func runReplay(_ *cobra.Command, _ []string) error {
	// 1. Load Config (to get server URL)
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 2. Read Events
	logPath := filepath.Join(cfg.DataDir, "events", "events.jsonl")
	events, err := readLastEvents(logPath, replayLast, replayType)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		ui.Warning("No matching events found to replay.")
		return nil
	}

	// 3. Send to Server via HTTP
	ui.Info("Replaying %d events to %s...", len(events), "localhost")

	serverURL := fmt.Sprintf("http://127.0.0.1:%d/debug/events", cfg.Port)
	client := &http.Client{Timeout: 2 * time.Second}
	successCount := 0

	for _, evt := range events {
		fmt.Printf("  -> %s %s\n", evt.Type, evt.RepoName)

		// Encode
		data, err := json.Marshal(evt)
		if err != nil {
			return err
		}

		// POST
		resp, err := client.Post(serverURL, "application/json", strings.NewReader(string(data)))
		if err != nil {
			return fmt.Errorf("failed to send event: %w (is server running?)", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ui.Warning("  ! Server returned %s", resp.Status)
		} else {
			successCount++
		}

		time.Sleep(10 * time.Millisecond)
	}

	ui.Success("Replayed %d events successfully.", successCount)
	return nil
}

// readLastEvents matches showRecentEvents logic but returns struct slice
func readLastEvents(path string, limit int, filterType string) ([]event.Event, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	// Simple read all for now (simpler than reverse seeker for Replay which usually needs old->new order?)
	// Wait, "Replay last 10" usually means "The 10 most recent events, in chronological order".
	// My `events.go` read loop was creating a list.

	scanner := bufio.NewScanner(file)
	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	// Filter

	// We scan from start just to be simple, but we only keep ones that match filter?
	// Logic:
	// 1. Collect all candidates (apply type filter).
	// 2. Take last N.

	var candidates []event.Event
	for _, line := range allLines {
		var evt event.Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			continue
		}

		if filterType != "" && string(evt.Type) != filterType {
			continue
		}
		candidates = append(candidates, evt)
	}

	count := len(candidates)
	start := 0
	if count > limit {
		start = count - limit
	}

	return candidates[start:], nil
}
