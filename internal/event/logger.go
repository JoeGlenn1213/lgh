package event

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileLogger logs events to a JSONL file
type FileLogger struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

// NewFileLogger creates a logger that writes to events.jsonl in the given directory
func NewFileLogger(dir string) (*FileLogger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create event dir: %w", err)
	}

	path := filepath.Join(dir, "events.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open event log: %w", err)
	}

	return &FileLogger{
		filePath: path,
		file:     f,
	}, nil
}

// Handle processes a single event by writing it to the file
func (l *FileLogger) Handle(e Event) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return
	}

	data, err := json.Marshal(e)
	if err != nil {
		// Should not happen for our simple struct
		return
	}

	_, _ = l.file.Write(data)
	_, _ = l.file.WriteString("\n")
}

// Close closes the underlying file
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// ReadRecent reads the last N lines from the log file
// This is a naive implementation that reads the whole file.
// For v1.x with small usage, it's fine. For v2, use efficient tailing.
func (l *FileLogger) ReadRecent(limit int) ([]Event, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Not implemented for MVP
	return nil, fmt.Errorf("not implemented yet")
}
