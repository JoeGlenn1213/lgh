package event

import (
	"sync"
)

// Handler is a function that processes an event
type Handler func(Event)

// Bus serves as the central event dispatcher
type Bus struct {
	handlers []Handler
	mu       sync.RWMutex
}

var defaultBus = &Bus{}

// Subscribe adds a subscriber to the default event bus
func Subscribe(h Handler) {
	defaultBus.mu.Lock()
	defer defaultBus.mu.Unlock()
	defaultBus.handlers = append(defaultBus.handlers, h)
}

// Publish creates and broadcasts an event to all subscribers
// It executes handlers synchronously to adhere to "simple and reliable" principle.
// Handlers are responsible for their own concurrency if needed.
func Publish(eventType Type, repoName string, payload map[string]interface{}) {
	evt := New(eventType, repoName, payload)
	defaultBus.publish(evt)
}

func (b *Bus) publish(evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, h := range b.handlers {
		// Recover from panic in handlers to prevent crashing the main app
		func(handler Handler) {
			defer func() {
				if r := recover(); r != nil {
					// In a real app we might log this panic
					// fmt.Fprintf(os.Stderr, "Panic in event handler: %v\n", r)
				}
			}()
			handler(evt)
		}(h)
	}
}

// Lifecycle management
type Closer interface {
	Close() error
}

var (
	closers   []Closer
	closersMu sync.Mutex
)

// RegisterCloser registers a resource that needs to be closed on shutdown
func RegisterCloser(c Closer) {
	closersMu.Lock()
	defer closersMu.Unlock()
	closers = append(closers, c)
}

// Shutdown closes all registered resources
func Shutdown() {
	closersMu.Lock()
	defer closersMu.Unlock()

	// Close in reverse order of registration
	for i := len(closers) - 1; i >= 0; i-- {
		_ = closers[i].Close()
	}
}
