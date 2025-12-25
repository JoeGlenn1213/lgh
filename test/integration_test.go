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

// Package test provides integration tests for LGH
package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

var lghBinary string

func init() {
	// Get the path to the lgh binary
	wd, _ := os.Getwd()
	binaryName := "lgh"
	if runtime.GOOS == "windows" {
		binaryName = "lgh.exe"
	}
	lghBinary = filepath.Join(filepath.Dir(wd), binaryName)
}

func TestLGHInit(t *testing.T) {
	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "lgh-test-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// Run lgh init
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	cmd := exec.Command(lghBinary, "init")
	cmd.Env = append(os.Environ(), "HOME="+tmpHome)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lgh init failed: %v\nOutput: %s", err, output)
	}

	// Verify directory structure
	lghDir := filepath.Join(tmpHome, ".localgithub")
	if _, err := os.Stat(lghDir); os.IsNotExist(err) {
		t.Error("~/.localgithub directory not created")
	}

	reposDir := filepath.Join(lghDir, "repos")
	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		t.Error("~/.localgithub/repos directory not created")
	}

	configFile := filepath.Join(lghDir, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("config.yaml not created")
	}

	mappingsFile := filepath.Join(lghDir, "mappings.yaml")
	if _, err := os.Stat(mappingsFile); os.IsNotExist(err) {
		t.Error("mappings.yaml not created")
	}

	t.Log("lgh init: PASSED")
}

func TestLGHStatus(t *testing.T) {
	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "lgh-test-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// First init
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	initCmd := exec.Command(lghBinary, "init")
	initCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	if initOutput, initErr := initCmd.CombinedOutput(); initErr != nil {
		t.Fatalf("lgh init failed: %v\nOutput: %s", initErr, initOutput)
	}

	// Run lgh status
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	statusCmd := exec.Command(lghBinary, "status")
	statusCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	statusOutput, statusErr := statusCmd.CombinedOutput()
	if statusErr != nil {
		t.Fatalf("lgh status failed: %v\nOutput: %s", statusErr, statusOutput)
	}

	// Check output contains expected information
	outputStr := string(statusOutput)
	if !strings.Contains(outputStr, "Status") {
		t.Error("Status output missing expected information")
	}

	t.Log("lgh status: PASSED")
}

func TestLGHList(t *testing.T) {
	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "lgh-test-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// First init
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	initCmd := exec.Command(lghBinary, "init")
	initCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("lgh init failed: %v\nOutput: %s", err, output)
	}

	// Run lgh list (should show empty)
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	listCmd := exec.Command(lghBinary, "list")
	listCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	listOutput, listErr := listCmd.CombinedOutput()
	if listErr != nil {
		t.Fatalf("lgh list failed: %v\nOutput: %s", listErr, listOutput)
	}

	// Should contain "No repositories" message
	if !strings.Contains(string(listOutput), "No repositories") {
		t.Error("Expected 'No repositories' message")
	}

	t.Log("lgh list (empty): PASSED")
}

func TestLGHAddAndList(t *testing.T) {
	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "lgh-test-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// Create a test git repository
	testRepo, err := os.MkdirTemp("", "lgh-test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create test repo: %v", err)
	}
	defer os.RemoveAll(testRepo)

	// Initialize git repo
	gitInit := exec.Command("git", "init")
	gitInit.Dir = testRepo
	if gInitOut, gInitErr := gitInit.CombinedOutput(); gInitErr != nil {
		t.Fatalf("git init failed: %v\nOutput: %s", gInitErr, gInitOut)
	}

	// Create a file and commit
	readmeFile := filepath.Join(testRepo, "README.md")
	if wErr := os.WriteFile(readmeFile, []byte("# Test Repo\n"), 0600); wErr != nil {
		t.Fatalf("Failed to create README: %v", wErr)
	}

	gitAdd := exec.Command("git", "add", ".")
	gitAdd.Dir = testRepo
	_ = gitAdd.Run()

	gitCommit := exec.Command("git", "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "-m", "Initial commit")
	gitCommit.Dir = testRepo
	_ = gitCommit.Run()

	// First init LGH
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	initCmd := exec.Command(lghBinary, "init")
	initCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	if initOut, initErr := initCmd.CombinedOutput(); initErr != nil {
		t.Fatalf("lgh init failed: %v\nOutput: %s", initErr, initOut)
	}

	// Add the test repo
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	addCmd := exec.Command(lghBinary, "add", testRepo, "--name", "test-repo", "--no-remote")
	addCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	addOutput, addErr := addCmd.CombinedOutput()
	if addErr != nil {
		t.Fatalf("lgh add failed: %v\nOutput: %s", addErr, addOutput)
	}

	// Verify bare repo was created
	bareRepoPath := filepath.Join(tmpHome, ".localgithub", "repos", "test-repo.git")
	if _, statErr := os.Stat(bareRepoPath); os.IsNotExist(statErr) {
		t.Error("Bare repository not created")
	}

	// Run lgh list
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	listCmd := exec.Command(lghBinary, "list")
	listCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	output, err := listCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lgh list failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "test-repo") {
		t.Error("Repository not shown in list")
	}

	t.Log("lgh add and list: PASSED")
}

func TestLGHRemove(t *testing.T) {
	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "lgh-test-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// Create a test git repository
	testRepo, err := os.MkdirTemp("", "lgh-test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create test repo: %v", err)
	}
	defer os.RemoveAll(testRepo)

	// Initialize git repo
	gitInit := exec.Command("git", "init")
	gitInit.Dir = testRepo
	_ = gitInit.Run()

	// First init LGH
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	initCmdNoShadow := exec.Command(lghBinary, "init")
	initCmdNoShadow.Env = append(os.Environ(), "HOME="+tmpHome)
	_ = initCmdNoShadow.Run()

	// Add the test repo
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	addCmdNoShadow := exec.Command(lghBinary, "add", testRepo, "--name", "test-repo", "--no-remote")
	addCmdNoShadow.Env = append(os.Environ(), "HOME="+tmpHome)
	_ = addCmdNoShadow.Run()

	// Remove the repo
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	removeCmd := exec.Command(lghBinary, "remove", "test-repo", "--force")
	removeCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	output, err := removeCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lgh remove failed: %v\nOutput: %s", err, output)
	}

	// Verify bare repo was deleted
	bareRepoPath := filepath.Join(tmpHome, ".localgithub", "repos", "test-repo.git")
	// nolint:gosec // G304: Stat checks a derived path
	if _, statErr := os.Stat(bareRepoPath); !os.IsNotExist(statErr) {
		t.Error("Bare repository was not deleted")
	}

	// Verify not in list
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	listCmdNoShadow := exec.Command(lghBinary, "list")
	listCmdNoShadow.Env = append(os.Environ(), "HOME="+tmpHome)
	output, _ = listCmdNoShadow.CombinedOutput()
	if strings.Contains(string(output), "test-repo") {
		t.Error("Repository still shown in list after removal")
	}

	t.Log("lgh remove: PASSED")
}

func TestLGHServeAndClone(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "lgh-test-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// Create a test git repository
	testRepo, err := os.MkdirTemp("", "lgh-test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create test repo: %v", err)
	}
	defer os.RemoveAll(testRepo)

	// Initialize git repo with content
	gitInit := exec.Command("git", "init")
	gitInit.Dir = testRepo
	if gInitOut, gInitErr := gitInit.CombinedOutput(); gInitErr != nil {
		t.Fatalf("git init failed: %v\nOutput: %s", gInitErr, gInitOut)
	}

	readmeFile := filepath.Join(testRepo, "README.md")
	// nolint:gosec // G306: Permissions are set to 0600 for a test file.
	gitAdd.Dir = testRepo
	_ = gitAdd.Run()
	gitCommit := exec.Command("git", "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "-m", "Initial commit")
	gitCommit.Dir = testRepo
	_ = gitCommit.Run()

	// Init LGH
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	initCmd := exec.Command(lghBinary, "init")
	initCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	_ = initCmd.Run()

	// Add the test repo
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	addCmd := exec.Command(lghBinary, "add", testRepo, "--name", "test-repo", "--no-remote")
	addCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	_ = addCmd.Run()

	// Push changes
	bareRepoPath := filepath.Join(tmpHome, ".localgithub", "repos", "test-repo.git")
	// nolint:gosec // G204: Subprocess launched with variable. bareRepoPath is a trusted path.
	pushCmd := exec.Command("git", "push", bareRepoPath, "HEAD:main")
	pushCmd.Dir = testRepo
	if output, pErr := pushCmd.CombinedOutput(); pErr != nil {
		t.Fatalf("git push failed: %v\nOutput: %s", pErr, output)
	}

	// Start the server in background
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	serveCmd := exec.Command(lghBinary, "serve", "--port", "19418")
	serveCmd.Env = append(os.Environ(), "HOME="+tmpHome)
	var serveOutput bytes.Buffer
	serveCmd.Stdout = &serveOutput
	serveCmd.Stderr = &serveOutput

	if startErr := serveCmd.Start(); startErr != nil {
		t.Fatalf("Failed to start server: %v", startErr)
	}
	defer serveCmd.Process.Kill()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Try to clone
	cloneDir, err := os.MkdirTemp("", "lgh-test-clone-*")
	if err != nil {
		t.Fatalf("Failed to create clone dir: %v", err)
	}
	defer os.RemoveAll(cloneDir)

	// nolint:gosec // G204: Subprocess launched with variable. cloneDir is a trusted path.
	cloneCmd := exec.Command("git", "clone", "http://127.0.0.1:19418/test-repo.git", cloneDir+"/cloned")
	cOutput, cErr := cloneCmd.CombinedOutput()
	if cErr != nil {
		t.Logf("Clone output: %s", cOutput)
		t.Logf("Server output: %s", serveOutput.String())
		t.Fatalf("git clone failed: %v", cErr)
	}

	// Verify cloned content
	clonedReadme := filepath.Join(cloneDir, "cloned", "README.md")
	if _, err := os.Stat(clonedReadme); os.IsNotExist(err) {
		t.Error("Cloned README.md not found")
	}

	t.Log("lgh serve and clone: PASSED")
}

func TestLGHVersion(t *testing.T) {
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	cmd := exec.Command(lghBinary, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lgh --version failed: %v", err)
	}

	if !strings.Contains(string(output), "LGH") {
		t.Error("Version output missing LGH identifier")
	}

	t.Log("lgh --version: PASSED")
}

func TestLGHHelp(t *testing.T) {
	// nolint:gosec // G204: Subprocess launched with variable. lghBinary is a trusted path.
	cmd := exec.Command(lghBinary, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lgh --help failed: %v", err)
	}

	outputStr := string(output)
	expectedCommands := []string{"init", "serve", "add", "list", "status", "tunnel"}
	for _, expected := range expectedCommands {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Help output missing command: %s", expected)
		}
	}

	t.Log("lgh --help: PASSED")
}
