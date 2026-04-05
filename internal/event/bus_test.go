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

// ---- Bus ----

func TestBusSubscribe(t *testing.T) {
	bus := &Bus{}
	handlerCalled := false

	bus.handlers = append(bus.handlers, func(e Event) {
		handlerCalled = true
	})

	if len(bus.handlers) != 1 {
		t.Errorf("len(handlers) = %d, want 1", len(bus.handlers))
	}
	_ = handlerCalled // suppress unused warning
}

func TestBusPublish(t *testing.T) {
	bus := &Bus{}
	var receivedEvent Event
	var wg sync.WaitGroup
	wg.Add(1)

	bus.handlers = append(bus.handlers, func(e Event) {
		receivedEvent = e
		wg.Done()
	})

	testEvent := New(GitPush, "test-repo", map[string]interface{}{"commit": "abc123"})
	bus.publish(testEvent)

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if receivedEvent.Type != GitPush {
			t.Errorf("receivedEvent.Type = %v, want %v", receivedEvent.Type, GitPush)
		}
		if receivedEvent.RepoName != "test-repo" {
			t.Errorf("receivedEvent.RepoName = %q, want %q", receivedEvent.RepoName, "test-repo")
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for event handler")
	}
}

func TestBusPublishToMultipleHandlers(t *testing.T) {
	bus := &Bus{}
	count := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		bus.handlers = append(bus.handlers, func(e Event) {
			mu.Lock()
			count++
			mu.Unlock()
			wg.Done()
		})
	}

	testEvent := New(GitPush, "test-repo", nil)
	bus.publish(testEvent)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		mu.Lock()
		if count != 3 {
			t.Errorf("count = %d, want 3", count)
		}
		mu.Unlock()
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for event handlers")
	}
}

func TestBusHandlerPanicRecovery(t *testing.T) {
	bus := &Bus{}
	var lastEvent Event
	var wg sync.WaitGroup
	wg.Add(1)

	// First handler panics
	bus.handlers = append(bus.handlers, func(e Event) {
		panic("test panic")
	})

	// Second handler should still receive event
	bus.handlers = append(bus.handlers, func(e Event) {
		lastEvent = e
		wg.Done()
	})

	testEvent := New(GitPush, "test-repo", nil)
	bus.publish(testEvent)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if lastEvent.Type != GitPush {
			t.Errorf("lastEvent.Type = %v, want %v", lastEvent.Type, GitPush)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for event handler after panic")
	}
}

// ---- Event Type constants ----

func TestEventTypes(t *testing.T) {
	tests := []struct {
		et     Type
		expect string
	}{
		{RepoAdded, "repo.added"},
		{RepoRemoved, "repo.removed"},
		{GitPush, "git.push"},
		{GitTag, "git.tag"},
	}
	for _, tc := range tests {
		if string(tc.et) != tc.expect {
			t.Errorf("Type %v = %q, want %q", tc.et, string(tc.et), tc.expect)
		}
	}
}

// ---- New ----

func TestNew(t *testing.T) {
	evt := New(GitPush, "my-repo", map[string]interface{}{"key": "value"})

	if evt.ID == "" {
		t.Error("New() should generate a non-empty ID")
	}
	if evt.Type != GitPush {
		t.Errorf("evt.Type = %v, want %v", evt.Type, GitPush)
	}
	if evt.RepoName != "my-repo" {
		t.Errorf("evt.RepoName = %q, want %q", evt.RepoName, "my-repo")
	}
	if evt.Timestamp.IsZero() {
		t.Error("evt.Timestamp should not be zero")
	}
	if evt.Payload["key"] != "value" {
		t.Errorf("evt.Payload[key] = %v, want %v", evt.Payload["key"], "value")
	}
}

// ---- Subscribe/Publish helper functions ----

func TestSubscribePublish(t *testing.T) {
	// Reset default bus handlers for test isolation
	defaultBus.mu.Lock()
	defaultBus.handlers = nil
	defaultBus.mu.Unlock()

	var receivedType Type
	var wg sync.WaitGroup
	wg.Add(1)

	Subscribe(func(e Event) {
		receivedType = e.Type
		wg.Done()
	})

	Publish(GitPush, "test-repo", nil)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if receivedType != GitPush {
			t.Errorf("receivedType = %v, want %v", receivedType, GitPush)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for event")
	}

	// Cleanup
	defaultBus.mu.Lock()
	defaultBus.handlers = nil
	defaultBus.mu.Unlock()
}
