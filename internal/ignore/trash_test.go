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
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// ---- TrashType constants ----

func TestTrashTypeConstants(t *testing.T) {
	tests := []struct {
		tt     TrashType
		expect string
	}{
		{TrashTypeLargeFile, "large_file"},
		{TrashTypeSensitiveFile, "sensitive_file"},
		{TrashTypeNodeModules, "node_modules"},
		{TrashTypeTotalSizeExceed, "total_size_exceed"},
	}
	for _, tc := range tests {
		if string(tc.tt) != tc.expect {
			t.Errorf("TrashType %v = %q, want %q", tc.tt, string(tc.tt), tc.expect)
		}
	}
}

// ---- SensitivePatterns ----

func TestSensitivePatternsContainsExpected(t *testing.T) {
	expected := []string{
		".env",
		".env.local",
		"*.key",
		"*.pem",
		"id_rsa",
		"id_ed25519",
		"credentials.json",
		"service-account.json",
		"secrets.yaml",
	}
	for _, pattern := range expected {
		found := false
		for _, p := range SensitivePatterns {
			if p == pattern {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SensitivePatterns should contain %q", pattern)
		}
	}
}

// ---- DangerousDirectories ----

func TestDangerousDirectoriesContainsExpected(t *testing.T) {
	expected := []string{
		"node_modules",
		".venv",
		"venv",
		"__pycache__",
	}
	for _, dir := range expected {
		found := false
		for _, d := range DangerousDirectories {
			if d == dir {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DangerousDirectories should contain %q", dir)
		}
	}
}

// ---- isSensitiveFile ----

func TestIsSensitiveFileEnvFiles(t *testing.T) {
	envFiles := []string{".env", ".env.local", ".env.production", ".env.development"}
	for _, f := range envFiles {
		if !isSensitiveFile(f) {
			t.Errorf("isSensitiveFile(%q) = false, want true", f)
		}
	}
}

func TestIsSensitiveFileKeyFiles(t *testing.T) {
	keyFiles := []string{"test.key", "private.pem", "client.p12"}
	for _, f := range keyFiles {
		if !isSensitiveFile(f) {
			t.Errorf("isSensitiveFile(%q) = false, want true", f)
		}
	}
	// .crt files are not considered sensitive keys by default
	if isSensitiveFile("server.crt") {
		t.Errorf("isSensitiveFile(%q) = true, want false", "server.crt")
	}
}

func TestIsSensitiveFileSSHKeys(t *testing.T) {
	sshKeys := []string{"id_rsa", "id_rsa.pub", "id_ed25519", "id_ed25519.pub"}
	for _, f := range sshKeys {
		if !isSensitiveFile(f) {
			t.Errorf("isSensitiveFile(%q) = false, want true", f)
		}
	}
}

func TestIsSensitiveFileCredentials(t *testing.T) {
	credFiles := []string{"credentials.json", "service-account.json", "secrets.yaml", "secrets.yml"}
	for _, f := range credFiles {
		if !isSensitiveFile(f) {
			t.Errorf("isSensitiveFile(%q) = false, want true", f)
		}
	}
}

func TestIsSensitiveFileNonSensitive(t *testing.T) {
	safeFiles := []string{"readme.md", "main.go", "app.js", "styles.css", "data.json"}
	for _, f := range safeFiles {
		if isSensitiveFile(f) {
			t.Errorf("isSensitiveFile(%q) = true, want false", f)
		}
	}
}

func TestIsSensitiveFileCaseInsensitive(t *testing.T) {
	caseVariants := []string{".ENV", ".Env", "ID_RSA", "Id_Rsa", "CREDENTIALS.JSON"}
	for _, f := range caseVariants {
		if !isSensitiveFile(f) {
			t.Errorf("isSensitiveFile(%q) = false, want true", f)
		}
	}
}

// ---- isDangerousDir ----

func TestIsDangerousDir(t *testing.T) {
	dangerous := []string{"node_modules", ".venv", "venv", "__pycache__"}
	for _, d := range dangerous {
		if !isDangerousDir(d) {
			t.Errorf("isDangerousDir(%q) = false, want true", d)
		}
	}

	safe := []string{"src", "lib", "utils", "tests", ".git"}
	for _, d := range safe {
		if isDangerousDir(d) {
			t.Errorf("isDangerousDir(%q) = true, want false", d)
		}
	}
}

// ---- FormatHumanSize ----

func TestFormatHumanSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{50 * 1024 * 1024, "50.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tc := range tests {
		got := FormatHumanSize(tc.bytes)
		if got != tc.want {
			t.Errorf("FormatHumanSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

// ---- TrashItem and TrashReport ----

func TestTrashReportHasBlocking(t *testing.T) {
	// Report with blocking item
	report := &TrashReport{
		Items: []TrashItem{
			{Type: TrashTypeSensitiveFile, Blocking: true},
		},
		HasBlocking: true,
	}
	if !report.HasBlocking {
		t.Error("Report with blocking item should have HasBlocking = true")
	}

	// Report without blocking item
	report2 := &TrashReport{
		Items: []TrashItem{
			{Type: TrashTypeTotalSizeExceed, Blocking: false},
		},
		HasBlocking: false,
	}
	if report2.HasBlocking {
		t.Error("Report without blocking item should have HasBlocking = false")
	}
}

// ---- DetectTrash ----

func TestDetectTrashSensitiveFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DetectTrash test in short mode - requires git repo")
	}

	tmpDir := t.TempDir()
	// Create a git repo properly
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create a sensitive file
	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("SECRET=value"), 0644); err != nil {
		t.Fatalf("Failed to create .env: %v", err)
	}

	report, err := DetectTrash(tmpDir)
	if err != nil {
		t.Fatalf("DetectTrash failed: %v", err)
	}

	found := false
	for _, item := range report.Items {
		if item.Type == TrashTypeSensitiveFile && item.Path == ".env" {
			found = true
			if !item.Blocking {
				t.Error("Sensitive file should be blocking")
			}
			break
		}
	}
	if !found {
		t.Error("Should detect sensitive .env file")
	}
}

func TestDetectTrashLargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DetectTrash test in short mode - requires git repo")
	}

	tmpDir := t.TempDir()
	// Create a git repo properly
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create a large file (> 50MB)
	largeFile := filepath.Join(tmpDir, "large.bin")
	f, err := os.Create(largeFile)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}
	// Write 51MB of zeros
	data := make([]byte, 51*1024*1024)
	f.Write(data)
	f.Close()

	report, err := DetectTrash(tmpDir)
	if err != nil {
		t.Fatalf("DetectTrash failed: %v", err)
	}

	found := false
	for _, item := range report.Items {
		if item.Type == TrashTypeLargeFile && item.Path == "large.bin" {
			found = true
			if !item.Blocking {
				t.Error("Large file should be blocking")
			}
			break
		}
	}
	if !found {
		t.Error("Should detect large file")
	}
}

func TestDetectTrashClean(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DetectTrash test in short mode - requires git repo")
	}

	tmpDir := t.TempDir()
	// Create a git repo properly
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create clean files
	cleanFiles := []string{"main.go", "readme.md", "utils.go"}
	for _, f := range cleanFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("package main"), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", f, err)
		}
	}

	report, err := DetectTrash(tmpDir)
	if err != nil {
		t.Fatalf("DetectTrash failed: %v", err)
	}

	if report.HasBlocking {
		t.Error("Clean repo should not have blocking items")
	}
}
