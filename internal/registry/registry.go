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

// Package registry provide platform-specific file locking and project registration
package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/JoeGlenn1213/lgh/internal/config"
	"gopkg.in/yaml.v3"
)

// FileLock is defined in lock_unix.go and lock_windows.go

// RepoMapping represents a single repository mapping
type RepoMapping struct {
	Name       string    `yaml:"name"`
	SourcePath string    `yaml:"source_path"`
	BarePath   string    `yaml:"bare_path"`
	CreatedAt  time.Time `yaml:"created_at"`
}

// Mappings holds all repository mappings
type Mappings struct {
	Repos []RepoMapping `yaml:"repos"`
}

// Registry manages the mappings file
type Registry struct {
	path string
	mu   sync.RWMutex
}

// New creates a new Registry instance
func New() *Registry {
	return &Registry{
		path: config.GetMappingsPath(),
	}
}

// NewWithPath creates a Registry with a custom path (for testing)
func NewWithPath(path string) *Registry {
	return &Registry{
		path: path,
	}
}

// load reads the mappings file with file locking
func (r *Registry) load() (*Mappings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fileLock := NewFileLock(r.path + ".lock")
	if err := fileLock.Lock(); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer fileLock.Unlock()

	mappings := &Mappings{Repos: []RepoMapping{}}

	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return mappings, nil
		}
		return nil, fmt.Errorf("failed to read mappings file: %w", err)
	}

	if err := yaml.Unmarshal(data, mappings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mappings: %w", err)
	}

	return mappings, nil
}

// save writes the mappings file with file locking
func (r *Registry) save(mappings *Mappings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fileLock := NewFileLock(r.path + ".lock")
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer fileLock.Unlock()

	data, err := yaml.Marshal(mappings)
	if err != nil {
		return fmt.Errorf("failed to marshal mappings: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(r.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write mappings file: %w", err)
	}

	return nil
}

// Add adds a new repository mapping
func (r *Registry) Add(name, sourcePath, barePath string) error {
	mappings, err := r.load()
	if err != nil {
		return err
	}

	// Check if already exists
	for _, repo := range mappings.Repos {
		if repo.Name == name {
			return fmt.Errorf("repository '%s' already exists", name)
		}
	}

	mappings.Repos = append(mappings.Repos, RepoMapping{
		Name:       name,
		SourcePath: sourcePath,
		BarePath:   barePath,
		CreatedAt:  time.Now(),
	})

	return r.save(mappings)
}

// Remove removes a repository mapping by name
func (r *Registry) Remove(name string) error {
	mappings, err := r.load()
	if err != nil {
		return err
	}

	found := false
	newRepos := []RepoMapping{}
	for _, repo := range mappings.Repos {
		if repo.Name == name {
			found = true
			continue
		}
		newRepos = append(newRepos, repo)
	}

	if !found {
		return fmt.Errorf("repository '%s' not found", name)
	}

	mappings.Repos = newRepos
	return r.save(mappings)
}

// List returns all repository mappings
func (r *Registry) List() ([]RepoMapping, error) {
	mappings, err := r.load()
	if err != nil {
		return nil, err
	}
	return mappings.Repos, nil
}

// Find finds a repository mapping by name
func (r *Registry) Find(name string) (*RepoMapping, error) {
	mappings, err := r.load()
	if err != nil {
		return nil, err
	}

	for _, repo := range mappings.Repos {
		if repo.Name == name {
			return &repo, nil
		}
	}

	return nil, fmt.Errorf("repository '%s' not found", name)
}

// FindBySourcePath finds a repository mapping by source path
func (r *Registry) FindBySourcePath(sourcePath string) (*RepoMapping, error) {
	mappings, err := r.load()
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, err
	}

	for _, repo := range mappings.Repos {
		if repo.SourcePath == absPath {
			return &repo, nil
		}
	}

	return nil, fmt.Errorf("repository at path '%s' not found", sourcePath)
}

// Exists checks if a repository with the given name exists
func (r *Registry) Exists(name string) bool {
	_, err := r.Find(name)
	return err == nil
}

// Count returns the number of registered repositories
func (r *Registry) Count() (int, error) {
	mappings, err := r.load()
	if err != nil {
		return 0, err
	}
	return len(mappings.Repos), nil
}

// Clear removes all mappings (use with caution)
func (r *Registry) Clear() error {
	return r.save(&Mappings{Repos: []RepoMapping{}})
}
