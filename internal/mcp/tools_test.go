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

package mcp

import (
	"testing"
)

// ---- getString ----

func TestGetString(t *testing.T) {
	args := map[string]interface{}{
		"name":    "test-repo",
		"empty":   "",
		"invalid": 123,
		"missing": nil,
	}

	if got := getString(args, "name"); got != "test-repo" {
		t.Errorf("getString(name) = %q, want %q", got, "test-repo")
	}
	if got := getString(args, "empty"); got != "" {
		t.Errorf("getString(empty) = %q, want empty string", got)
	}
	if got := getString(args, "invalid"); got != "" {
		t.Errorf("getString(invalid) = %q, want empty string", got)
	}
	if got := getString(args, "missing"); got != "" {
		t.Errorf("getString(missing) = %q, want empty string", got)
	}
}

// ---- getBool ----

func TestGetBool(t *testing.T) {
	args := map[string]interface{}{
		"trueVal":  true,
		"falseVal": false,
		"invalid":  "not bool",
		"missing":  nil,
	}

	if got := getBool(args, "trueVal"); !got {
		t.Error("getBool(trueVal) = false, want true")
	}
	if got := getBool(args, "falseVal"); got {
		t.Error("getBool(falseVal) = true, want false")
	}
	if got := getBool(args, "invalid"); got {
		t.Error("getBool(invalid) = true, want false")
	}
	if got := getBool(args, "missing"); got {
		t.Error("getBool(missing) = true, want false")
	}
}

// ---- getFloat ----

func TestGetFloat(t *testing.T) {
	args := map[string]interface{}{
		"intVal":   float64(42),
		"floatVal": float64(3.14),
		"invalid":  "not float",
		"missing":  nil,
	}

	if got := getFloat(args, "intVal"); got != 42 {
		t.Errorf("getFloat(intVal) = %v, want 42", got)
	}
	if got := getFloat(args, "floatVal"); got != 3.14 {
		t.Errorf("getFloat(floatVal) = %v, want 3.14", got)
	}
	if got := getFloat(args, "invalid"); got != 0 {
		t.Errorf("getFloat(invalid) = %v, want 0", got)
	}
	if got := getFloat(args, "missing"); got != 0 {
		t.Errorf("getFloat(missing) = %v, want 0", got)
	}
}
