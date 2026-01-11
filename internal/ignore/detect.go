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

// Package ignore provides smart .gitignore generation based on project type detection
package ignore

import (
	"os"
	"path/filepath"
)

// ProjectType represents the detected project type
type ProjectType string

const (
	ProjectTypePython  ProjectType = "python"
	ProjectTypeGo      ProjectType = "go"
	ProjectTypeNode    ProjectType = "node"
	ProjectTypeJava    ProjectType = "java"
	ProjectTypeRust    ProjectType = "rust"
	ProjectTypeAI      ProjectType = "ai"
	ProjectTypeUnknown ProjectType = "unknown"
)

// DetectionResult holds the result of project type detection
type DetectionResult struct {
	Type       ProjectType
	Confidence float64 // 0.0 - 1.0
	Indicators []string
}

// DetectProjectType analyzes a directory and determines the project type
func DetectProjectType(dir string) DetectionResult {
	result := DetectionResult{
		Type:       ProjectTypeUnknown,
		Confidence: 0.0,
		Indicators: []string{},
	}

	// Detection rules in priority order
	detectors := []struct {
		projectType ProjectType
		files       []string
		weight      float64
	}{
		// AI/ML projects (highest priority - often coexist with Python)
		{ProjectTypeAI, []string{"*.ipynb", "models/", "data/", "checkpoints/"}, 0.9},
		// Python
		{ProjectTypePython, []string{"requirements.txt", "pyproject.toml", "setup.py", "Pipfile"}, 0.85},
		// Go
		{ProjectTypeGo, []string{"go.mod"}, 0.95},
		// Node/TypeScript
		{ProjectTypeNode, []string{"package.json", "tsconfig.json"}, 0.9},
		// Java
		{ProjectTypeJava, []string{"pom.xml", "build.gradle", "build.gradle.kts"}, 0.9},
		// Rust
		{ProjectTypeRust, []string{"Cargo.toml"}, 0.95},
	}

	for _, detector := range detectors {
		for _, pattern := range detector.files {
			if matchExists(dir, pattern) {
				result.Type = detector.projectType
				result.Confidence = detector.weight
				result.Indicators = append(result.Indicators, pattern)
			}
		}
		// If we found a match, break (priority order)
		if result.Type != ProjectTypeUnknown {
			break
		}
	}

	return result
}

// matchExists checks if a file/pattern exists in the directory
func matchExists(dir, pattern string) bool {
	// For directory patterns (ending with /)
	if len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
		dirPath := filepath.Join(dir, pattern[:len(pattern)-1])
		info, err := os.Stat(dirPath)
		return err == nil && info.IsDir()
	}

	// For glob patterns
	if containsGlob(pattern) {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		return err == nil && len(matches) > 0
	}

	// For exact file matches
	filePath := filepath.Join(dir, pattern)
	_, err := os.Stat(filePath)
	return err == nil
}

// containsGlob checks if a pattern contains glob characters
func containsGlob(pattern string) bool {
	for _, c := range pattern {
		if c == '*' || c == '?' || c == '[' {
			return true
		}
	}
	return false
}

// HasGitignore checks if a .gitignore file already exists
func HasGitignore(dir string) bool {
	gitignorePath := filepath.Join(dir, ".gitignore")
	_, err := os.Stat(gitignorePath)
	return err == nil
}
