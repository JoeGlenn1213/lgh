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

// Package event implements the internal event bus and logging infrastructure.
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
					_ = r // suppress unused var lint
				}
			}()
			handler(evt)
		}(h)
	}
}

// Closer is the interface for resources that need to be closed on shutdown
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
