package event

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MaxLogSize is the maximum size of the log file before rotation (10MB)
const MaxLogSize = 10 * 1024 * 1024 // 10MB

// FileLogger logs events to a JSONL file asynchronously
type FileLogger struct {
	filePath string
	file     *os.File
	queue    chan Event
	wg       sync.WaitGroup
	once     sync.Once
}

// NewFileLogger creates a logger that writes to events.jsonl in the given directory
func NewFileLogger(dir string) (*FileLogger, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create event dir: %w", err)
	}

	path := filepath.Join(dir, "events.jsonl")
	// 0600 permissions for security
	// nolint:gosec // G304: path is internally constructed and trusted
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open event log: %w", err)
	}

	l := &FileLogger{
		filePath: path,
		file:     f,
		queue:    make(chan Event, 100), // Buffer up to 100 events
	}

	// Start worker
	l.wg.Add(1)
	go l.worker()

	return l, nil
}

// Handle queues an event for logging. It is safe for concurrent use.
// It will not block unless the queue is full (backpressure).
func (l *FileLogger) Handle(e Event) {
	// Try to enqueue
	select {
	case l.queue <- e:
		// success
	default:
		// Queue full. To ensure the critical path (git push) is never blocked,
		// we drop the event. In a production system we'd metric this.
		return
	}
}

func (l *FileLogger) worker() {
	defer l.wg.Done()

	for e := range l.queue {
		if err := l.rotateIfNeeded(); err != nil {
			// In case of rotation error, we try to proceed with current file
			// or just ignore. Safety first (don't crash).
			_ = err
		}

		data, err := json.Marshal(e)
		if err != nil {
			continue
		}

		// Write directly to current file handle
		if _, err := l.file.Write(data); err == nil {
			_, _ = l.file.WriteString("\n")
		}
	}
}

func (l *FileLogger) rotateIfNeeded() error {
	fi, err := l.file.Stat()
	if err != nil {
		return err
	}

	if fi.Size() < MaxLogSize {
		return nil
	}

	// Rotate
	_ = l.file.Close()

	timestamp := time.Now().Format("20060102-150405")
	backupPath := l.filePath + "." + timestamp

	if err := os.Rename(l.filePath, backupPath); err != nil {
		// Log error?
		_ = err
	}

	// nolint:gosec // G304: filePath is trusted
	f, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	l.file = f
	return nil
}

// Close closes the queue, waits for worker, and closes file
func (l *FileLogger) Close() error {
	l.once.Do(func() {
		close(l.queue)
	})
	l.wg.Wait()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}
