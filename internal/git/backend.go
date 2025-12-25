package git

import (
	"fmt"
	"net/http"
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
	reposDir string
	readOnly bool
	gitPath  string
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
	if _, err := os.Stat(httpBackendPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("git-http-backend not found at %s", httpBackendPath)
	}

	return &Backend{
		reposDir: reposDir,
		readOnly: readOnly,
		gitPath:  gitPath,
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

	// Get git-http-backend path
	execPath, err := getGitExecPath()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	httpBackendPath := filepath.Join(execPath, "git-http-backend")

	// Setup CGI handler
	handler := &cgi.Handler{
		Path: httpBackendPath,
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
	if service == "git-receive-pack" {
		return true
	}

	return false
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
