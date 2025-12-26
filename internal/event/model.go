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
