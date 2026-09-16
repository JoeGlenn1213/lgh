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

package main

import "testing"

// cmd/lgh has no other tests yet; this file covers the ActionD endpoint
// override used by `lgh up`'s CI wait (multi-instance setups).
func TestActiondBaseURLFromUp(t *testing.T) {
	t.Setenv("ACTIOND_URL", "")
	if got := actiondBaseURL(); got != "http://localhost:3000" {
		t.Errorf("got %q, want default", got)
	}
	t.Setenv("ACTIOND_URL", "http://127.0.0.1:3999//")
	if got := actiondBaseURL(); got != "http://127.0.0.1:3999" {
		t.Errorf("got %q, want trimmed override", got)
	}
}

// The auto-start opt-out gate must be a pure env read so isolated LGH
// instances never spawn a globally installed ActionD.
func TestTryStartActionDOptOut(t *testing.T) {
	t.Setenv("LGH_AUTOSTART_ACTIOND", "0")
	tryStartActionD() // must return without touching PATH candidates
}
