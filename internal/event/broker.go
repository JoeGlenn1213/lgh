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

package event

import (
	"sync"
)

// Broker manages real-time event subscriptions
type Broker struct {
	clients map[chan Event]bool
	mu      sync.RWMutex
}

var defaultBroker = &Broker{
	clients: make(map[chan Event]bool),
}

// StartBroker starts listening to internal events and broadcasting them.
// Should be called once at server startup.
func StartBroker() {
	Subscribe(func(e Event) {
		defaultBroker.Broadcast(e)
	})
}

// SubscribeClient registers a new client channel
func SubscribeClient() chan Event {
	ch := make(chan Event, 100) // Buffered channel
	defaultBroker.mu.Lock()
	defaultBroker.clients[ch] = true
	defaultBroker.mu.Unlock()
	return ch
}

// UnsubscribeClient removes a client channel
func UnsubscribeClient(ch chan Event) {
	defaultBroker.mu.Lock()
	if _, ok := defaultBroker.clients[ch]; ok {
		delete(defaultBroker.clients, ch)
		close(ch)
	}
	defaultBroker.mu.Unlock()
}

// Broadcast sends an event to all connected clients.
// This bypasses the main event bus (logging), useful for replaying events.
func Broadcast(e Event) {
	defaultBroker.Broadcast(e)
}

// Broadcast sends an event to all connected clients
func (b *Broker) Broadcast(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.clients {
		select {
		case ch <- e:
		default:
			// Client saturated, skip drop
			// In production we might want to close slow clients
		}
	}
}
