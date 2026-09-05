package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeRemote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"git@github.com:AcmeCorp/repo.git", "github.com/acmecorp/repo"},
		{"https://github.com/AcmeCorp/repo.git", "github.com/acmecorp/repo"},
		{"http://gitlab.com/org/project.git", "gitlab.com/org/project"},
		{"git@bitbucket.org:team/repo.git", "bitbucket.org/team/repo"},
		{"https://github.com/user/repo", "github.com/user/repo"},
	}

	for _, tt := range tests {
		got := NormalizeRemote(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeRemote(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		value   string
		pattern string
		want    bool
	}{
		{"github.com/acmecorp/repo", "github.com/acmecorp/*", true},
		{"github.com/acmecorp/another", "github.com/acmecorp/*", true},
		{"github.com/other/repo", "github.com/acmecorp/*", false},
		{"github.com/user/repo", "github.com/user/repo", true},
		{"github.com/user/repo", "github.com/user/other", false},
		{"anything", "*", true},
		{"", "*", true},
	}

	for _, tt := range tests {
		got := MatchPattern(tt.value, tt.pattern)
		if got != tt.want {
			t.Errorf("MatchPattern(%q, %q) = %v, want %v", tt.value, tt.pattern, got, tt.want)
		}
	}
}

func TestEvaluate_NoPolicies(t *testing.T) {
	v := Evaluate("personal", "git@github.com:AcmeCorp/repo.git")
	if v != nil {
		t.Errorf("expected no violation with no policies, got: %v", v.Message)
	}
}

func TestEvaluate_EmptyRemote(t *testing.T) {
	v := Evaluate("work", "")
	if v != nil {
		t.Error("expected no violation with empty remote")
	}
}

func TestEvaluate_WithUserPolicies(t *testing.T) {
	dir := t.TempDir()
	policiesDir := filepath.Join(dir, ".git-profile")
	os.MkdirAll(policiesDir, 0o755)

	policiesPath := filepath.Join(policiesDir, "policies.yaml")
	content := `policies:
  - name: acme-corp
    remotePatterns:
      - "github.com/acmecorp/*"
    allowedProfiles:
      - work
    mode: strict
  - name: personal-repos
    remotePatterns:
      - "github.com/johndoe/*"
    allowedProfiles:
      - personal
    mode: warning
`
	os.WriteFile(policiesPath, []byte(content), 0o644)

	t.Setenv("GITX_POLICIES_PATH", policiesPath)

	// Allowed: work profile for AcmeCorp repo
	v := Evaluate("work", "git@github.com:AcmeCorp/inventory-service.git")
	if v != nil {
		t.Errorf("expected no violation for work+acmecorp, got: %s", v.Message)
	}

	// Violation: personal profile for AcmeCorp repo
	v = Evaluate("personal", "git@github.com:AcmeCorp/inventory-service.git")
	if v == nil {
		t.Fatal("expected violation for personal+acmecorp")
	}
	if v.PolicyName != "acme-corp" {
		t.Errorf("expected policy name 'acme-corp', got %q", v.PolicyName)
	}
	if v.Mode != "strict" {
		t.Errorf("expected mode 'strict', got %q", v.Mode)
	}
	if len(v.AllowedProfiles) != 1 || v.AllowedProfiles[0] != "work" {
		t.Errorf("expected allowed profiles [work], got %v", v.AllowedProfiles)
	}

	// Allowed: personal for personal repos
	v = Evaluate("personal", "https://github.com/johndoe/side-project.git")
	if v != nil {
		t.Errorf("expected no violation for personal+johndoe, got: %s", v.Message)
	}

	// Violation: work for personal repos
	v = Evaluate("work", "https://github.com/johndoe/side-project.git")
	if v == nil {
		t.Fatal("expected violation for work+johndoe")
	}
	if v.Mode != "warning" {
		t.Errorf("expected mode 'warning', got %q", v.Mode)
	}
}

func TestEvaluate_UnknownRemote(t *testing.T) {
	dir := t.TempDir()
	policiesDir := filepath.Join(dir, ".git-profile")
	os.MkdirAll(policiesDir, 0o755)

	policiesPath := filepath.Join(policiesDir, "policies.yaml")
	content := `policies:
  - name: acme-corp
    remotePatterns:
      - "github.com/acmecorp/*"
    allowedProfiles:
      - work
`
	os.WriteFile(policiesPath, []byte(content), 0o644)

	t.Setenv("GITX_POLICIES_PATH", policiesPath)

	// Remote that doesn't match any policy → no violation
	v := Evaluate("personal", "git@github.com:random-org/random-repo.git")
	if v != nil {
		t.Errorf("expected no violation for unmatched remote, got: %s", v.Message)
	}
}

func TestIsAllowed(t *testing.T) {
	tests := []struct {
		profile string
		allowed []string
		want    bool
	}{
		{"work", []string{"work", "personal"}, true},
		{"WORK", []string{"work"}, true},
		{"freelance", []string{"work", "personal"}, false},
		{"work", []string{}, false},
	}

	for _, tt := range tests {
		got := isAllowed(tt.profile, tt.allowed)
		if got != tt.want {
			t.Errorf("isAllowed(%q, %v) = %v, want %v", tt.profile, tt.allowed, got, tt.want)
		}
	}
}

func TestMatchesRemote(t *testing.T) {
	tests := []struct {
		remote   string
		patterns []string
		want     bool
	}{
		{"git@github.com:AcmeCorp/repo.git", []string{"github.com/acmecorp/*"}, true},
		{"https://github.com/GlobexInc/repo.git", []string{"github.com/acmecorp/*"}, false},
		{"anything", []string{"*"}, true},
		{"git@github.com:AcmeCorp/repo.git", []string{}, false},
	}

	for _, tt := range tests {
		got := matchesRemote(tt.remote, tt.patterns)
		if got != tt.want {
			t.Errorf("matchesRemote(%q, %v) = %v, want %v", tt.remote, tt.patterns, got, tt.want)
		}
	}
}

func TestLoadUserPolicies_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policies.yaml")
	os.WriteFile(path, []byte("this is not valid yaml: ["), 0o644)

	t.Setenv("GITX_POLICIES_PATH", path)

	policies := loadUserPolicies()
	if len(policies) != 0 {
		t.Errorf("expected 0 policies for malformed YAML, got %d", len(policies))
	}
}

func TestLoadUserPolicies_DefaultMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policies.yaml")
	content := `policies:
  - name: globex
    remotePatterns:
      - "github.com/globexinc/*"
    allowedProfiles:
      - work
`
	os.WriteFile(path, []byte(content), 0o644)

	t.Setenv("GITX_POLICIES_PATH", path)

	policies := loadUserPolicies()
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(policies))
	}
	if policies[0].Mode != "strict" {
		t.Errorf("expected default mode 'strict', got %q", policies[0].Mode)
	}
}
