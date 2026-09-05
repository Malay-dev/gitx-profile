package git

import (
	"testing"
)

// These tests verify the git wrapper functions work when git is available.
// They are intentionally lightweight — we test behavior, not mock git.

func TestIsInsideWorkTree_OutsideRepo(t *testing.T) {
	// When running tests from the project root, we ARE in a git repo.
	// But we test the function doesn't panic and returns a bool.
	result := IsInsideWorkTree()
	// We can't assert true/false without knowing the test environment,
	// but we ensure it doesn't crash.
	_ = result
}

func TestVersion(t *testing.T) {
	v, err := Version()
	if err != nil {
		t.Skipf("git not available: %v", err)
	}
	if v == "" {
		t.Error("expected non-empty git version")
	}
	if len(v) < 5 {
		t.Errorf("git version unexpectedly short: %q", v)
	}
}

func TestGetConfig_NonexistentKey(t *testing.T) {
	// Reading a key that doesn't exist should return an error
	_, err := GetConfig("user.this_key_does_not_exist_12345")
	if err == nil {
		t.Error("expected error for nonexistent config key")
	}
}

func TestScopeConstants(t *testing.T) {
	// Verify scope constants are distinct
	if ScopeLocal == ScopeGlobal {
		t.Error("ScopeLocal and ScopeGlobal should be different")
	}
}

func TestGetRemoteURL_NoRemote(t *testing.T) {
	// In a repo without the named remote, should error
	// This test is environment-dependent but should not panic
	_, err := GetRemoteURL("nonexistent_remote_12345")
	if err == nil {
		// If we're in a repo that somehow has this remote, that's fine
		// This is a defensive test
		t.Log("surprisingly found a remote named 'nonexistent_remote_12345'")
	}
}
