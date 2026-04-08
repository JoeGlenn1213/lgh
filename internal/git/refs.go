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

package git

import (
	"os/exec"
	"strings"
)

// GetRefs returns all references (branches and tags) and their hashes
func GetRefs(repoPath string) (map[string]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "show-ref")
	output, err := cmd.CombinedOutput()
	// show-ref returns exit code 1 if no refs exist (new repo)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return make(map[string]string), nil
		}
		return nil, err
	}

	refs := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			ref := parts[1]
			refs[ref] = hash
		}
	}
	return refs, nil
}

// GetChangedFiles returns the list of files changed between two commits.
// For created refs, oldHash should be the empty tree hash or parent of first commit.
// For deleted refs, returns empty slice.
func GetChangedFiles(repoPath, oldHash, newHash string) ([]string, error) {
	// Handle deletion - no files to report
	if newHash == "" || newHash == "0000000000000000000000000000000000000000" {
		return []string{}, nil
	}

	// Handle creation - diff against empty tree or use --root
	var cmd *exec.Cmd
	if oldHash == "" || oldHash == "0000000000000000000000000000000000000000" {
		// New branch/tag - diff against empty tree or show all files in first commit
		cmd = exec.Command("git", "-C", repoPath, "diff-tree", "--no-commit-id", "--name-only", "-r", newHash)
	} else {
		// Updated branch - diff between old and new
		cmd = exec.Command("git", "-C", repoPath, "diff", "--name-only", oldHash, newHash)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	result := make([]string, 0, len(files))
	for _, f := range files {
		if f != "" {
			result = append(result, f)
		}
	}
	return result, nil
}
