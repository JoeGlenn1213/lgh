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
