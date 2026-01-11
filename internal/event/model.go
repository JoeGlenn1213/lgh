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
	"time"

	"github.com/google/uuid"
)

// Type represents the type of an event
type Type string

const (
	// RepoAdded indicates a new repository was registered
	RepoAdded Type = "repo.added"
	// RepoRemoved indicates a repository was unregistered
	RepoRemoved Type = "repo.removed"

	// GitPush indicates a git push operation (receive-pack) occurred
	GitPush Type = "git.push"
	// GitTag indicates a tag was created/pushed
	GitTag Type = "git.tag"
)

// Event represents a system event in LGH
type Event struct {
	ID        string                 `json:"id"`
	Type      Type                   `json:"type"`
	RepoName  string                 `json:"repo"` // The name of the repository involved
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// New creates a new event with a UUID and current timestamp
func New(eventType Type, repoName string, payload map[string]interface{}) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		RepoName:  repoName,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}
