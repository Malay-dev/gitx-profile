package profile

import (
	"testing"
)

func TestNormalizeRemote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"git@github.com:AcmeCorp/repo.git", "github.com/acmecorp/repo"},
		{"https://github.com/AcmeCorp/repo.git", "github.com/acmecorp/repo"},
		{"http://github.com/GlobexInc/repo.git", "github.com/globexinc/repo"},
		{"git@github.com:johndoe/gitx-profile.git", "github.com/johndoe/gitx-profile"},
		{"https://gitlab.company.com/team/project.git", "gitlab.company.com/team/project"},
		{"git@bitbucket.org:org/repo.git", "bitbucket.org/org/repo"},
		// No .git suffix
		{"https://github.com/user/repo", "github.com/user/repo"},
	}

	for _, tt := range tests {
		got := normalizeRemote(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeRemote(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		value   string
		pattern string
		want    bool
	}{
		// Wildcard matching
		{"github.com/acmecorp/repo", "github.com/acmecorp/*", true},
		{"github.com/acmecorp/another-repo", "github.com/acmecorp/*", true},
		{"github.com/other/repo", "github.com/acmecorp/*", false},
		// Exact matching
		{"github.com/user/repo", "github.com/user/repo", true},
		{"github.com/user/repo", "github.com/user/other", false},
		// Case insensitive
		{"github.com/AcmeCorp/Repo", "github.com/acmecorp/*", true},
	}

	for _, tt := range tests {
		got := matchesPattern(tt.value, tt.pattern)
		if got != tt.want {
			t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.value, tt.pattern, got, tt.want)
		}
	}
}

func TestMatchByRemote(t *testing.T) {
	profiles := []Profile{
		{Name: "work", UserEmail: "alice@acmecorp.com"},
		{Name: "personal", UserEmail: "alice@gmail.com"},
	}

	rules := []RemoteRule{
		{Pattern: "github.com/acmecorp/*", ProfileName: "work"},
		{Pattern: "github.com/johndoe/*", ProfileName: "personal"},
	}

	// SSH remote → work
	result := MatchByRemote("git@github.com:AcmeCorp/inventory.git", profiles, rules)
	if result.Confidence == "none" {
		t.Error("expected match for AcmeCorp remote")
	}
	if result.Profile.Name != "work" {
		t.Errorf("expected 'work' profile, got %q", result.Profile.Name)
	}

	// HTTPS remote → personal
	result = MatchByRemote("https://github.com/johndoe/side-project.git", profiles, rules)
	if result.Confidence == "none" {
		t.Error("expected match for personal remote")
	}
	if result.Profile.Name != "personal" {
		t.Errorf("expected 'personal' profile, got %q", result.Profile.Name)
	}

	// Unknown remote → no match
	result = MatchByRemote("https://github.com/random/repo.git", profiles, rules)
	if result.Confidence != "none" {
		t.Error("expected no match for unknown remote")
	}
}

func TestMatchByRemote_NoRules(t *testing.T) {
	profiles := []Profile{
		{Name: "work", UserEmail: "alice@acmecorp.com"},
	}

	result := MatchByRemote("git@github.com:AcmeCorp/repo.git", profiles, nil)
	if result.Confidence != "none" {
		t.Error("expected no match with no rules")
	}
}

func TestMatchByRemote_ProfileNotFound(t *testing.T) {
	profiles := []Profile{
		{Name: "work", UserEmail: "alice@acmecorp.com"},
	}

	// Rule references a profile that doesn't exist
	rules := []RemoteRule{
		{Pattern: "github.com/acmecorp/*", ProfileName: "nonexistent"},
	}

	result := MatchByRemote("git@github.com:AcmeCorp/repo.git", profiles, rules)
	if result.Confidence != "none" {
		t.Error("expected no match when referenced profile doesn't exist")
	}
}
