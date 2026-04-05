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

package registry

import (
	"path/filepath"
	"testing"
	"time"
)

// ---- RepoMapping ----

func TestRepoMappingFields(t *testing.T) {
	now := time.Now()
	rm := RepoMapping{
		Name:       "test-repo",
		SourcePath: "/path/to/source",
		BarePath:   "/path/to/bare",
		CreatedAt:  now,
	}

	if rm.Name != "test-repo" {
		t.Errorf("Name = %q, want %q", rm.Name, "test-repo")
	}
	if rm.SourcePath != "/path/to/source" {
		t.Errorf("SourcePath = %q, want %q", rm.SourcePath, "/path/to/source")
	}
	if rm.BarePath != "/path/to/bare" {
		t.Errorf("BarePath = %q, want %q", rm.BarePath, "/path/to/bare")
	}
	if !rm.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", rm.CreatedAt, now)
	}
}

// ---- New ----

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New() returned nil")
	}
	if r.path == "" {
		t.Error("New() should set a default path")
	}
}

// ---- NewWithPath ----

func TestNewWithPath(t *testing.T) {
	customPath := "/custom/path/mappings.yaml"
	r := NewWithPath(customPath)
	if r == nil {
		t.Fatal("NewWithPath() returned nil")
	}
	if r.path != customPath {
		t.Errorf("path = %q, want %q", r.path, customPath)
	}
}

// ---- Registry Add/List ----

func TestRegistryAddAndList(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	err := r.Add("test-repo", "/source/path", "/bare/path")
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	repos, err := r.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("len(repos) = %d, want 1", len(repos))
	}

	if repos[0].Name != "test-repo" {
		t.Errorf("repos[0].Name = %q, want %q", repos[0].Name, "test-repo")
	}
	if repos[0].SourcePath != "/source/path" {
		t.Errorf("repos[0].SourcePath = %q, want %q", repos[0].SourcePath, "/source/path")
	}
}

func TestRegistryAddDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	err := r.Add("test-repo", "/source/path", "/bare/path")
	if err != nil {
		t.Fatalf("First Add() failed: %v", err)
	}

	err = r.Add("test-repo", "/another/path", "/another/bare")
	if err == nil {
		t.Fatal("Add() should fail for duplicate name")
	}
}

// ---- Registry Remove ----

func TestRegistryRemove(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	r.Add("repo1", "/source1", "/bare1")
	r.Add("repo2", "/source2", "/bare2")

	err := r.Remove("repo1")
	if err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	repos, _ := r.List()
	if len(repos) != 1 {
		t.Errorf("len(repos) = %d, want 1", len(repos))
	}
	if repos[0].Name != "repo2" {
		t.Errorf("repos[0].Name = %q, want %q", repos[0].Name, "repo2")
	}
}

func TestRegistryRemoveNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	err := r.Remove("nonexistent")
	if err == nil {
		t.Fatal("Remove() should fail for nonexistent repo")
	}
}

// ---- Registry Find ----

func TestRegistryFind(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	r.Add("test-repo", "/source/path", "/bare/path")

	repo, err := r.Find("test-repo")
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}
	if repo.Name != "test-repo" {
		t.Errorf("repo.Name = %q, want %q", repo.Name, "test-repo")
	}
}

func TestRegistryFindNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	_, err := r.Find("nonexistent")
	if err == nil {
		t.Fatal("Find() should fail for nonexistent repo")
	}
}

// ---- Registry FindBySourcePath ----

func TestRegistryFindBySourcePath(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	r.Add("test-repo", "/source/path", "/bare/path")

	repo, err := r.FindBySourcePath("/source/path")
	if err != nil {
		t.Fatalf("FindBySourcePath() failed: %v", err)
	}
	if repo.Name != "test-repo" {
		t.Errorf("repo.Name = %q, want %q", repo.Name, "test-repo")
	}
}

// ---- Registry Exists ----

func TestRegistryExists(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	r.Add("test-repo", "/source/path", "/bare/path")

	if !r.Exists("test-repo") {
		t.Error("Exists() = false, want true for existing repo")
	}
	if r.Exists("nonexistent") {
		t.Error("Exists() = true, want false for nonexistent repo")
	}
}

// ---- Registry Count ----

func TestRegistryCount(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	count, err := r.Count()
	if err != nil {
		t.Fatalf("Count() failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0 for empty registry", count)
	}

	r.Add("repo1", "/source1", "/bare1")
	r.Add("repo2", "/source2", "/bare2")

	count, _ = r.Count()
	if count != 2 {
		t.Errorf("Count() = %d, want 2", count)
	}
}

// ---- Registry Clear ----

func TestRegistryClear(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	r.Add("repo1", "/source1", "/bare1")
	r.Add("repo2", "/source2", "/bare2")

	err := r.Clear()
	if err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	count, _ := r.Count()
	if count != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", count)
	}
}

// ---- Registry empty file ----

func TestRegistryEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	mappingsFile := filepath.Join(tmpDir, "mappings.yaml")
	r := NewWithPath(mappingsFile)

	// File doesn't exist yet, should return empty list
	repos, err := r.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("len(repos) = %d, want 0 for nonexistent file", len(repos))
	}
}

// ---- Mappings struct ----

func TestMappingsYAML(t *testing.T) {
	m := Mappings{
		Repos: []RepoMapping{
			{Name: "repo1", SourcePath: "/path1", BarePath: "/bare1"},
			{Name: "repo2", SourcePath: "/path2", BarePath: "/bare2"},
		},
	}

	if len(m.Repos) != 2 {
		t.Errorf("len(m.Repos) = %d, want 2", len(m.Repos))
	}
}
