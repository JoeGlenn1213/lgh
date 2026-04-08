// Copyright (c) 2025 JoeGlenn1213
// Commit Status Storage for CI/CD integration

package git

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CommitStatus represents the CI status of a commit
type CommitStatus struct {
	CommitSHA string    `json:"commit_sha"`
	Plugin    string    `json:"plugin"`
	Status    string    `json:"status"` // "pending", "success", "failure", "error", "cancelled"
	Summary   string    `json:"summary,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// CommitStatusReport represents the aggregate status of a commit
type CommitStatusReport struct {
	CommitSHA string          `json:"commit_sha"`
	Overall   string          `json:"overall"` // "pending", "success", "failure", "error"
	Statuses  []CommitStatus  `json:"statuses"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// StatusStore manages commit status storage
type StatusStore struct {
	mu      sync.RWMutex
	dataDir string
}

// NewStatusStore creates a new status store
func NewStatusStore(dataDir string) *StatusStore {
	return &StatusStore{
		dataDir: filepath.Join(dataDir, "statuses"),
	}
}

// Update adds or updates a commit status
func (s *StatusStore) Update(repo, commitSHA string, status CommitStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure directory exists
	repoDir := filepath.Join(s.dataDir, sanitizeRepoName(repo))
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("failed to create status directory: %w", err)
	}

	// Read existing statuses
	statusFile := filepath.Join(repoDir, commitSHA+".json")
	report, err := s.readReport(statusFile)
	if err != nil {
		report = &CommitStatusReport{
			CommitSHA: commitSHA,
			Statuses:  []CommitStatus{},
		}
	}

	// Update or append the status
	status.Timestamp = time.Now()
	found := false
	for i, cs := range report.Statuses {
		if cs.Plugin == status.Plugin {
			report.Statuses[i] = status
			found = true
			break
		}
	}
	if !found {
		report.Statuses = append(report.Statuses, status)
	}

	// Calculate overall status
	report.Overall = calculateOverallStatus(report.Statuses)
	report.UpdatedAt = time.Now()

	// Write back
	return s.writeReport(statusFile, report)
}

// Get retrieves the status report for a commit
func (s *StatusStore) Get(repo, commitSHA string) (*CommitStatusReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repoDir := filepath.Join(s.dataDir, sanitizeRepoName(repo))
	statusFile := filepath.Join(repoDir, commitSHA+".json")
	return s.readReport(statusFile)
}

func (s *StatusStore) readReport(path string) (*CommitStatusReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var report CommitStatusReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to parse status file: %w", err)
	}
	return &report, nil
}

func (s *StatusStore) writeReport(path string, report *CommitStatusReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func sanitizeRepoName(repo string) string {
	// Remove .git suffix if present and sanitize
	name := repo
	if len(name) > 4 && name[len(name)-4:] == ".git" {
		name = name[:len(name)-4]
	}
	// Replace any problematic characters
	result := make([]byte, 0, len(name))
	for _, c := range []byte(name) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else {
			result = append(result, '-')
		}
	}
	return string(result)
}

func calculateOverallStatus(statuses []CommitStatus) string {
	if len(statuses) == 0 {
		return "pending"
	}
	
	hasFailure := false
	hasError := false
	hasPending := false
	
	for _, cs := range statuses {
		switch cs.Status {
		case "failure":
			hasFailure = true
		case "error", "cancelled":
			hasError = true
		case "pending":
			hasPending = true
		}
	}
	
	if hasFailure {
		return "failure"
	}
	if hasError {
		return "error"
	}
	if hasPending {
		return "pending"
	}
	return "success"
}