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

// Package git provides Git backend and repository management for LGH
package git

import (
	"fmt"
	"net/http"

	// nolint:gosec // G504: net/http/cgi is required for Git backend and safe in modern Go.
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/JoeGlenn1213/lgh/internal/config"
)

// Backend handles Git HTTP backend requests
type Backend struct {
	reposDir        string
	readOnly        bool
	gitPath         string
	httpBackendPath string
}

// NewBackend creates a new Git HTTP backend handler
func NewBackend(reposDir string, readOnly bool) (*Backend, error) {
	gitPath, err := CheckGitInstalled()
	if err != nil {
		return nil, err
	}

	// Find git-http-backend
	gitExecPath, err := getGitExecPath()
	if err != nil {
		return nil, err
	}

	httpBackendPath := filepath.Join(gitExecPath, "git-http-backend")
	// On Windows, try with .exe extension if the base name doesn't exist
	if _, err := os.Stat(httpBackendPath); os.IsNotExist(err) {
		httpBackendPathExe := httpBackendPath + ".exe"
		if _, errExe := os.Stat(httpBackendPathExe); errExe == nil {
			httpBackendPath = httpBackendPathExe
		} else {
			return nil, fmt.Errorf("git-http-backend not found at %s or %s", httpBackendPath, httpBackendPathExe)
		}
	}

	return &Backend{
		reposDir:        reposDir,
		readOnly:        readOnly,
		gitPath:         gitPath,
		httpBackendPath: httpBackendPath,
	}, nil
}

// getGitExecPath returns the git exec path
func getGitExecPath() (string, error) {
	cmd := exec.Command("git", "--exec-path")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git exec path: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ServeHTTP implements http.Handler for Git HTTP backend
func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract repository name from path
	// URL format: /{repo}.git/info/refs or /{repo}.git/git-receive-pack etc.
	repoPath, gitPath := b.parseRequest(r)
	if repoPath == "" {
		http.Error(w, "Invalid repository path", http.StatusBadRequest)
		return
	}

	// Check if read-only mode and this is a push request
	if b.readOnly && b.isPushRequest(r, gitPath) {
		http.Error(w, "Repository is read-only. Push operations are not allowed.", http.StatusForbidden)
		return
	}

	// Build the full path to the bare repository
	fullRepoPath := filepath.Join(b.reposDir, repoPath)

	// Verify repository exists
	if !IsBareRepo(fullRepoPath) {
		http.Error(w, fmt.Sprintf("Repository not found: %s", repoPath), http.StatusNotFound)
		return
	}

	// Setup CGI handler using pre-validated httpBackendPath
	handler := &cgi.Handler{
		Path: b.httpBackendPath,
		Env: []string{
			"GIT_PROJECT_ROOT=" + b.reposDir,
			"GIT_HTTP_EXPORT_ALL=1",
			"REMOTE_USER=lgh-user",
		},
	}

	// The PATH_INFO needs to be set correctly
	// We need to update the request path for CGI
	originalPath := r.URL.Path
	r.URL.Path = "/" + repoPath + gitPath

	handler.ServeHTTP(w, r)

	// Restore original path
	r.URL.Path = originalPath
}

// parseRequest extracts the repository name and git path from the request
func (b *Backend) parseRequest(r *http.Request) (string, string) {
	path := r.URL.Path

	// Match patterns like /repo.git/info/refs, /repo.git/git-upload-pack, etc.
	re := regexp.MustCompile(`^/([^/]+\.git)(.*)$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) != 3 {
		return "", ""
	}

	return matches[1], matches[2]
}

// isPushRequest checks if the request is a push operation
func (b *Backend) isPushRequest(r *http.Request, gitPath string) bool {
	// Push requests are POST to git-receive-pack
	if r.Method != http.MethodPost {
		return false
	}

	if strings.Contains(gitPath, "git-receive-pack") {
		return true
	}

	// Also check for service=git-receive-pack in query string
	service := r.URL.Query().Get("service")
	return service == "git-receive-pack"
}

// Handler returns an http.Handler for the Git backend
func Handler(readOnly bool) (http.Handler, error) {
	cfg := config.Get()
	return NewBackend(cfg.ReposDir, readOnly)
}

// CreateHandler creates a new handler with explicit configuration
func CreateHandler(reposDir string, readOnly bool) (http.Handler, error) {
	return NewBackend(reposDir, readOnly)
}
