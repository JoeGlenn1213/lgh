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
	"testing"
	"time"
)

// ---- Broker ----

func TestBrokerSubscribeClient(t *testing.T) {
	// Reset default broker for test isolation
	defaultBroker.mu.Lock()
	defaultBroker.clients = make(map[chan Event]bool)
	defaultBroker.mu.Unlock()

	ch := SubscribeClient()
	if ch == nil {
		t.Fatal("SubscribeClient() returned nil channel")
	}

	defaultBroker.mu.RLock()
	if len(defaultBroker.clients) != 1 {
		t.Errorf("len(clients) = %d, want 1", len(defaultBroker.clients))
	}
	if _, ok := defaultBroker.clients[ch]; !ok {
		t.Error("client channel not found in clients map")
	}
	defaultBroker.mu.RUnlock()

	// Cleanup
	UnsubscribeClient(ch)
}

func TestBrokerUnsubscribeClient(t *testing.T) {
	// Reset default broker for test isolation
	defaultBroker.mu.Lock()
	defaultBroker.clients = make(map[chan Event]bool)
	defaultBroker.mu.Unlock()

	ch := SubscribeClient()
	UnsubscribeClient(ch)

	defaultBroker.mu.RLock()
	if len(defaultBroker.clients) != 0 {
		t.Errorf("len(clients) = %d, want 0 after unsubscribe", len(defaultBroker.clients))
	}
	defaultBroker.mu.RUnlock()
}

func TestBrokerBroadcast(t *testing.T) {
	// Reset default broker for test isolation
	defaultBroker.mu.Lock()
	defaultBroker.clients = make(map[chan Event]bool)
	defaultBroker.mu.Unlock()

	ch1 := make(chan Event, 10)
	ch2 := make(chan Event, 10)

	defaultBroker.mu.Lock()
	defaultBroker.clients[ch1] = true
	defaultBroker.clients[ch2] = true
	defaultBroker.mu.Unlock()

	testEvent := New(GitPush, "test-repo", nil)
	Broadcast(testEvent)

	var wg sync.WaitGroup
	wg.Add(2)
	timeout := time.After(500 * time.Millisecond)

	go func() {
		select {
		case e := <-ch1:
			if e.Type != GitPush {
				t.Errorf("ch1 received Type = %v, want %v", e.Type, GitPush)
			}
		case <-timeout:
			t.Error("Timeout waiting for event on ch1")
		}
		wg.Done()
	}()

	go func() {
		select {
		case e := <-ch2:
			if e.Type != GitPush {
				t.Errorf("ch2 received Type = %v, want %v", e.Type, GitPush)
			}
		case <-timeout:
			t.Error("Timeout waiting for event on ch2")
		}
		wg.Done()
	}()

	wg.Wait()
}

func TestBrokerBroadcastSlowClient(t *testing.T) {
	// Reset default broker for test isolation
	defaultBroker.mu.Lock()
	defaultBroker.clients = make(map[chan Event]bool)
	defaultBroker.mu.Unlock()

	ch := make(chan Event, 1) // Buffer of 1

	defaultBroker.mu.Lock()
	defaultBroker.clients[ch] = true
	defaultBroker.mu.Unlock()

	// Send first event - should succeed
	testEvent1 := New(GitPush, "repo1", nil)
	Broadcast(testEvent1)

	// Send second event - should not block (default case in select)
	testEvent2 := New(GitTag, "repo2", nil)
	Broadcast(testEvent2)

	// First event should be received
	select {
	case e := <-ch:
		if e.Type != GitPush {
			t.Errorf("Received Type = %v, want %v", e.Type, GitPush)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for first event")
	}
}

func TestBrokerBroadcastNoClients(t *testing.T) {
	// Reset default broker for test isolation
	defaultBroker.mu.Lock()
	defaultBroker.clients = make(map[chan Event]bool)
	defaultBroker.mu.Unlock()

	// Should not panic
	testEvent := New(GitPush, "test-repo", nil)
	Broadcast(testEvent)
}

// ---- Global broker functions ----

func TestGlobalSubscribeClient(t *testing.T) {
	// Reset default broker for test isolation
	defaultBroker.mu.Lock()
	defaultBroker.clients = make(map[chan Event]bool)
	defaultBroker.mu.Unlock()

	ch := SubscribeClient()
	if ch == nil {
		t.Fatal("SubscribeClient() returned nil")
	}

	defaultBroker.mu.RLock()
	if len(defaultBroker.clients) != 1 {
		t.Errorf("len(clients) = %d, want 1", len(defaultBroker.clients))
	}
	defaultBroker.mu.RUnlock()

	// Cleanup
	UnsubscribeClient(ch)
}

func TestGlobalUnsubscribeClient(t *testing.T) {
	// Reset default broker for test isolation
	defaultBroker.mu.Lock()
	defaultBroker.clients = make(map[chan Event]bool)
	defaultBroker.mu.Unlock()

	ch := SubscribeClient()
	UnsubscribeClient(ch)

	defaultBroker.mu.RLock()
	if len(defaultBroker.clients) != 0 {
		t.Errorf("len(clients) = %d, want 0 after unsubscribe", len(defaultBroker.clients))
	}
	defaultBroker.mu.RUnlock()
}

func TestGlobalBroadcast(t *testing.T) {
	// Reset default broker for test isolation
	defaultBroker.mu.Lock()
	defaultBroker.clients = make(map[chan Event]bool)
	defaultBroker.mu.Unlock()

	ch := SubscribeClient()
	defer UnsubscribeClient(ch)

	testEvent := New(GitPush, "test-repo", nil)
	Broadcast(testEvent)

	select {
	case e := <-ch:
		if e.Type != GitPush {
			t.Errorf("Received Type = %v, want %v", e.Type, GitPush)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for broadcasted event")
	}
}
