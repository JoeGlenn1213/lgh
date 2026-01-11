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

package ignore

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// MaxSingleFileSize is the maximum allowed size for a single file (50MB)
	MaxSingleFileSize int64 = 50 * 1024 * 1024
	// MaxTotalStagedSize is the maximum allowed total size for staged files (200MB)
	MaxTotalStagedSize int64 = 200 * 1024 * 1024
)

// TrashType represents the type of trash detected
type TrashType string

const (
	TrashTypeLargeFile       TrashType = "large_file"
	TrashTypeSensitiveFile   TrashType = "sensitive_file"
	TrashTypeNodeModules     TrashType = "node_modules"
	TrashTypeTotalSizeExceed TrashType = "total_size_exceed"
)

// TrashItem represents a detected trash item
type TrashItem struct {
	Type     TrashType
	Path     string
	Size     int64
	Message  string
	Blocking bool // If true, must be resolved before push
}

// TrashReport holds the results of trash detection
type TrashReport struct {
	Items       []TrashItem
	TotalSize   int64
	HasBlocking bool
}

// SensitivePatterns defines patterns for sensitive files that should never be committed
var SensitivePatterns = []string{
	".env",
	".env.local",
	".env.production",
	".env.development",
	"*.key",
	"*.pem",
	"id_rsa",
	"id_rsa.pub",
	"id_ed25519",
	"id_ed25519.pub",
	"*.p12",
	"*.pfx",
	"credentials.json",
	"service-account.json",
	"secrets.yaml",
	"secrets.yml",
}

// DangerousDirectories defines directories that should never be committed
var DangerousDirectories = []string{
	"node_modules",
	".venv",
	"venv",
	"__pycache__",
}

// DetectTrash scans a directory for potential issues using git ls-files
func DetectTrash(dir string) (*TrashReport, error) {
	report := &TrashReport{
		Items: []TrashItem{},
	}

	// Use git ls-files to get list of tracked and untracked (but not ignored) files
	// -c: cached (tracked)
	// -o: others (untracked)
	// --exclude-standard: respect .gitignore
	// -z: null-terminated output
	cmd := exec.Command("git", "ls-files", "-z", "-c", "-o", "--exclude-standard")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		// Fallback to walk if git fails (e.g. not a git repo yet, though lgh up ensures it)
		// Or return error? lgh up initializes git before this.
		// Let's fallback to walk but maybe just skip .git?
		// Actually, if git command fails, we probably can't push anyway.
		return nil, fmt.Errorf("git ls-files failed: %v", err)
	}

	files := strings.Split(string(output), "\x00")
	for _, relPath := range files {
		if relPath == "" {
			continue
		}

		fullPath := filepath.Join(dir, relPath)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue // File might have been deleted
		}

		if info.IsDir() {
			continue
		}

		// Check for dangerous directories (path components)
		parts := strings.Split(relPath, string(os.PathSeparator))
		for _, part := range parts {
			if isDangerousDir(part) {
				report.Items = append(report.Items, TrashItem{
					Type:     TrashTypeNodeModules,
					Path:     relPath,
					Message:  fmt.Sprintf("Contains dangerous directory '%s'", part),
					Blocking: true,
				})
				// Don't simplify breaking logic here to allow checking other rules?
				// Actually if it's node_modules it shouldn't be here if ignored.
				// If it IS here, it means it's NOT ignored.
				break
			}
		}

		// Check for sensitive files
		if isSensitiveFile(filepath.Base(relPath)) {
			report.Items = append(report.Items, TrashItem{
				Type:     TrashTypeSensitiveFile,
				Path:     relPath,
				Size:     info.Size(),
				Message:  "Sensitive file detected",
				Blocking: true,
			})
		}

		// Check for large files
		if info.Size() > MaxSingleFileSize {
			report.Items = append(report.Items, TrashItem{
				Type:     TrashTypeLargeFile,
				Path:     relPath,
				Size:     info.Size(),
				Message:  fmt.Sprintf("File exceeds %dMB limit", MaxSingleFileSize/(1024*1024)),
				Blocking: true,
			})
		}

		report.TotalSize += info.Size()
	}

	// Check total size
	if report.TotalSize > MaxTotalStagedSize {
		report.Items = append(report.Items, TrashItem{
			Type:     TrashTypeTotalSizeExceed,
			Size:     report.TotalSize,
			Message:  fmt.Sprintf("Total stage size (%dMB) exceeds %dMB limit", report.TotalSize/(1024*1024), MaxTotalStagedSize/(1024*1024)),
			Blocking: false, // Warning only
		})
	}

	// Update HasBlocking
	for _, item := range report.Items {
		if item.Blocking {
			report.HasBlocking = true
			break
		}
	}

	return report, nil
}

// isDangerousDir checks if a directory name is dangerous
func isDangerousDir(name string) bool {
	for _, dir := range DangerousDirectories {
		if name == dir {
			return true
		}
	}
	return false
}

// isSensitiveFile checks if a filename matches sensitive patterns
func isSensitiveFile(name string) bool {
	nameLower := strings.ToLower(name)
	for _, pattern := range SensitivePatterns {
		if strings.HasPrefix(pattern, "*.") {
			// Extension match
			ext := pattern[1:]
			if strings.HasSuffix(nameLower, ext) {
				return true
			}
		} else {
			// Exact match
			if nameLower == strings.ToLower(pattern) {
				return true
			}
		}
	}
	return false
}

// FormatHumanSize formats bytes to human readable string
func FormatHumanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
