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

// Package skill provides a simple Skill interface for LGH capabilities.
// This package is designed to be imported by external projects.
//
// Usage:
//
//	import "github.com/JoeGlenn1213/lgh/pkg/skill"
//
//	// Get a skill
//	backup := skill.Get("lgh.backup")
//
//	// Execute it
//	result, err := backup.Execute(ctx, skill.Input{
//	    "path": "/my/project",
//	    "message": "Daily backup",
//	})
package skill

import (
	"context"
	"time"
)

// Input represents skill input parameters
type Input map[string]interface{}

// Output represents skill execution result
type Output map[string]interface{}

// Metadata describes a skill
type Metadata struct {
	ID          string   `json:"id"`          // Unique identifier, e.g., "lgh.backup"
	Name        string   `json:"name"`        // Human-readable name
	Description string   `json:"description"` // What this skill does
	Category    string   `json:"category"`    // Category: "repo", "git", "server"
	InputSchema []Param  `json:"inputs"`      // Expected inputs
	Tags        []string `json:"tags"`        // Search tags
}

// Param describes an input parameter
type Param struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "bool", "number", "path"
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

// Result represents skill execution result
type Result struct {
	Success   bool          `json:"success"`
	Output    Output        `json:"output,omitempty"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// Skill is the core interface that all LGH skills implement
type Skill interface {
	// Meta returns skill metadata
	Meta() Metadata

	// Execute runs the skill with given input
	Execute(ctx context.Context, input Input) (*Result, error)
}

// skillRegistry holds all registered skills
var skillRegistry = make(map[string]Skill)

// Register adds a skill to the registry
func Register(s Skill) {
	skillRegistry[s.Meta().ID] = s
}

// Get retrieves a skill by ID
func Get(id string) Skill {
	return skillRegistry[id]
}

// List returns all registered skills
func List() []Metadata {
	var metas []Metadata
	for _, s := range skillRegistry {
		metas = append(metas, s.Meta())
	}
	return metas
}

// ListByCategory returns skills in a category
func ListByCategory(category string) []Metadata {
	var metas []Metadata
	for _, s := range skillRegistry {
		meta := s.Meta()
		if meta.Category == category {
			metas = append(metas, meta)
		}
	}
	return metas
}
