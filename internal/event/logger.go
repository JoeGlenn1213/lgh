package event

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

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
		// Queue full. We must decide: block or drop?
		// For an "observer", dropping is better than hanging the user's git push.
		// However, 100 events buffer is large enough for normal use.
		// Let's try to block for a short time, then drop?
		// For MVP simplicity: just block. If disk is that slow, system is broken.
		l.queue <- e
	}
}

func (l *FileLogger) worker() {
	defer l.wg.Done()

	// Create encoder once
	encoder := json.NewEncoder(l.file)

	for e := range l.queue {
		// Use JSON encoder which append newline automatically?
		// No, standard encoder appends newline.
		// But let's stick to explicit Marshal for control if needed, or Encoder is fine.
		if err := encoder.Encode(e); err != nil {
			// Log error to stderr?
			// fmt.Fprintf(os.Stderr, "Failed to log event: %v\n", err)
		}
		// Ensure it's flushed? Encoder usually buffers?
		// For events, we might want immediate flush. O_APPEND file might not need explicit Sync for every line.
	}
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
