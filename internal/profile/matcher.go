package profile

import (
	"strings"
)

// MatchResult represents the outcome of matching a remote URL to a profile.
type MatchResult struct {
	Profile    *Profile
	Confidence string // "exact", "pattern", "none"
	Reason     string
}

// MatchByRemote attempts to suggest a profile based on a remote URL.
// This is a simple pattern matcher — no regex in v1.
//
// Pattern examples:
//
//	"github.com/AcmeCorp/*" matches "git@github.com:AcmeCorp/repo.git"
//	"github.com/Malay-dev/*" matches "https://github.com/Malay-dev/repo.git"
func MatchByRemote(remote string, profiles []Profile, rules []RemoteRule) *MatchResult {
	// Normalize the remote URL
	normalized := normalizeRemote(remote)

	for _, rule := range rules {
		if matchesPattern(normalized, rule.Pattern) {
			// Find the profile
			for i := range profiles {
				if strings.EqualFold(profiles[i].Name, rule.ProfileName) {
					return &MatchResult{
						Profile:    &profiles[i],
						Confidence: "pattern",
						Reason:     "Remote matches pattern: " + rule.Pattern,
					}
				}
			}
		}
	}

	return &MatchResult{
		Confidence: "none",
		Reason:     "No matching rule for remote: " + remote,
	}
}

// RemoteRule maps a remote URL pattern to a profile name.
type RemoteRule struct {
	Pattern     string // e.g., "github.com/AcmeCorp/*"
	ProfileName string // e.g., "work"
}

// normalizeRemote converts various remote URL formats to a common form.
//
//	"git@github.com:AcmeCorp/repo.git"     → "github.com/acmecorp/repo"
//	"https://github.com/AcmeCorp/repo.git" → "github.com/acmecorp/repo"
func normalizeRemote(remote string) string {
	r := remote

	// SSH format: git@host:org/repo.git → host/org/repo
	if strings.Contains(r, "@") && strings.Contains(r, ":") {
		parts := strings.SplitN(r, "@", 2)
		if len(parts) == 2 {
			r = parts[1]
			r = strings.Replace(r, ":", "/", 1)
		}
	}

	// HTTPS format: strip protocol
	r = strings.TrimPrefix(r, "https://")
	r = strings.TrimPrefix(r, "http://")

	// Strip .git suffix
	r = strings.TrimSuffix(r, ".git")

	// Lowercase for comparison
	r = strings.ToLower(r)

	return r
}

// matchesPattern does a simple wildcard match.
// Only supports trailing "*" (e.g., "github.com/org/*").
func matchesPattern(value, pattern string) bool {
	pattern = strings.ToLower(pattern)
	value = strings.ToLower(value)

	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(value, prefix+"/")
	}

	// Exact match
	return value == pattern
}
