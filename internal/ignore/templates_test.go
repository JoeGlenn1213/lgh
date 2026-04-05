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
	"path/filepath"
	"strings"
	"testing"
)

// ---- GetTemplate ----

func TestGetTemplateGo(t *testing.T) {
	template := GetTemplate(ProjectTypeGo)
	if !strings.Contains(template, "# Binaries for programs and plugins") {
		t.Error("Go template should contain binary patterns")
	}
	if !strings.Contains(template, "*.exe") {
		t.Error("Go template should contain *.exe")
	}
}

func TestGetTemplatePython(t *testing.T) {
	template := GetTemplate(ProjectTypePython)
	if !strings.Contains(template, "__pycache__/") {
		t.Error("Python template should contain __pycache__")
	}
	if !strings.Contains(template, "venv/") {
		t.Error("Python template should contain venv")
	}
}

func TestGetTemplateNode(t *testing.T) {
	template := GetTemplate(ProjectTypeNode)
	if !strings.Contains(template, "node_modules/") {
		t.Error("Node template should contain node_modules")
	}
	if !strings.Contains(template, "dist/") {
		t.Error("Node template should contain dist")
	}
}

func TestGetTemplateJava(t *testing.T) {
	template := GetTemplate(ProjectTypeJava)
	if !strings.Contains(template, "*.class") {
		t.Error("Java template should contain *.class")
	}
	if !strings.Contains(template, "target/") {
		t.Error("Java template should contain target")
	}
}

func TestGetTemplateRust(t *testing.T) {
	template := GetTemplate(ProjectTypeRust)
	if !strings.Contains(template, "debug/") {
		t.Error("Rust template should contain debug/")
	}
	if !strings.Contains(template, "target/") {
		t.Error("Rust template should contain target")
	}
}

func TestGetTemplateAI(t *testing.T) {
	template := GetTemplate(ProjectTypeAI)
	if !strings.Contains(template, ".ipynb_checkpoints") {
		t.Error("AI template should contain notebook checkpoints pattern")
	}
	if !strings.Contains(template, "models/") {
		t.Error("AI template should contain models/ directory")
	}
	if !strings.Contains(template, "*.pt") {
		t.Error("AI template should contain model file patterns")
	}
}

func TestGetTemplateUnknown(t *testing.T) {
	template := GetTemplate(ProjectTypeUnknown)
	if template != universalTemplate {
		t.Error("Unknown project type should return universal template")
	}
}

// ---- Template contents ----

func TestUniversalTemplateContainsOSFiles(t *testing.T) {
	if !strings.Contains(universalTemplate, ".DS_Store") {
		t.Error("universalTemplate should contain .DS_Store")
	}
	if !strings.Contains(universalTemplate, "ehthumbs.db") {
		t.Error("universalTemplate should contain ehthumbs.db")
	}
}

func TestUniversalTemplateContainsEditorFiles(t *testing.T) {
	if !strings.Contains(universalTemplate, ".idea/") {
		t.Error("universalTemplate should contain .idea/")
	}
	if !strings.Contains(universalTemplate, ".vscode/") {
		t.Error("universalTemplate should contain .vscode/")
	}
}

func TestUniversalTemplateContainsSecrets(t *testing.T) {
	if !strings.Contains(universalTemplate, ".env") {
		t.Error("universalTemplate should contain .env")
	}
	if !strings.Contains(universalTemplate, "*.key") {
		t.Error("universalTemplate should contain *.key")
	}
}

func TestUniversalTemplateContainsLogs(t *testing.T) {
	if !strings.Contains(universalTemplate, "*.log") {
		t.Error("universalTemplate should contain *.log")
	}
}

// ---- GenerateGitignore ----

func TestGenerateGitignoreGo(t *testing.T) {
	tmpDir := t.TempDir()
	err := GenerateGitignore(tmpDir, ProjectTypeGo)
	if err != nil {
		t.Fatalf("GenerateGitignore failed: %v", err)
	}

	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read generated .gitignore: %v", err)
	}

	content := string(data)
	// Should contain universal template
	if !strings.Contains(content, "# OS generated files") {
		t.Error("Generated .gitignore should contain universal template header")
	}
	// Should contain Go-specific content
	if !strings.Contains(content, "# Binaries for programs and plugins") {
		t.Error("Generated .gitignore should contain Go template")
	}
	// Should have project type title
	if !strings.Contains(content, "# Go specific") {
		t.Error("Generated .gitignore should have Go specific section")
	}
}

func TestGenerateGitignorePython(t *testing.T) {
	tmpDir := t.TempDir()
	err := GenerateGitignore(tmpDir, ProjectTypePython)
	if err != nil {
		t.Fatalf("GenerateGitignore failed: %v", err)
	}

	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read generated .gitignore: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Python specific") {
		t.Error("Generated .gitignore should have Python specific section")
	}
}

func TestGenerateGitignoreUnknown(t *testing.T) {
	tmpDir := t.TempDir()
	err := GenerateGitignore(tmpDir, ProjectTypeUnknown)
	if err != nil {
		t.Fatalf("GenerateGitignore failed: %v", err)
	}

	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read generated .gitignore: %v", err)
	}

	content := string(data)
	// Should only contain universal template, no project-specific section
	if strings.Contains(content, " specific") {
		t.Error("Unknown project should not have project-specific section")
	}
}

// ---- EnsureGitignore ----

func TestEnsureGitignoreExists(t *testing.T) {
	tmpDir := t.TempDir()
	// Pre-existing .gitignore
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	projectType, err := EnsureGitignore(tmpDir)
	if err != nil {
		t.Fatalf("EnsureGitignore failed: %v", err)
	}
	if projectType != ProjectTypeUnknown {
		t.Errorf("EnsureGitignore() Type = %v, want %v", projectType, ProjectTypeUnknown)
	}

	// Should not overwrite existing
	data, _ := os.ReadFile(gitignorePath)
	if string(data) != "existing content" {
		t.Error("Should not overwrite existing .gitignore")
	}
}

func TestEnsureGitignoreNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a go.mod to trigger Go detection
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	projectType, err := EnsureGitignore(tmpDir)
	if err != nil {
		t.Fatalf("EnsureGitignore failed: %v", err)
	}
	if projectType != ProjectTypeGo {
		t.Errorf("EnsureGitignore() Type = %v, want %v", projectType, ProjectTypeGo)
	}

	// Should create .gitignore
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error("EnsureGitignore should create .gitignore")
	}
}

// ---- Template is not empty ----

func TestAllTemplatesNonEmpty(t *testing.T) {
	templates := []string{
		pythonTemplate,
		goTemplate,
		nodeTemplate,
		javaTemplate,
		rustTemplate,
		aiTemplate,
		universalTemplate,
	}
	names := []string{
		"pythonTemplate",
		"goTemplate",
		"nodeTemplate",
		"javaTemplate",
		"rustTemplate",
		"aiTemplate",
		"universalTemplate",
	}

	for i, tmpl := range templates {
		if len(tmpl) == 0 {
			t.Errorf("%s is empty", names[i])
		}
	}
}
