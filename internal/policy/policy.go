package policy

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Malay-dev/gitx-profile/internal/config"
	"gopkg.in/yaml.v3"
)

// Violation represents a policy check failure.
type Violation struct {
	Message         string
	AllowedProfiles []string
	PolicyName      string
	Mode            string // "strict", "warning", "advisory"
}

// Policy defines which profiles are allowed for matching remotes.
type Policy struct {
	Name            string   `yaml:"name"`
	RemotePatterns  []string `yaml:"remotePatterns"`
	AllowedProfiles []string `yaml:"allowedProfiles"`
	Mode            string   `yaml:"mode"` // "strict", "warning", "advisory"
}

// PolicyFile represents the top-level structure of a policies YAML file.
type PolicyFile struct {
	Policies []Policy `yaml:"policies"`
}

// RepoPolicyFile represents the .gitprofile.yml in a repository root.
type RepoPolicyFile struct {
	Organization    string   `yaml:"organization"`
	AllowedProfiles []string `yaml:"allowedProfiles"`
	Mode            string   `yaml:"mode"`
	Enforce         struct {
		SignedCommits bool `yaml:"signedCommits"`
	} `yaml:"enforce"`
}

// Evaluate checks if using a given profile in a repo with the given remote
// violates any policies. Returns nil if no violation.
func Evaluate(profileName string, remote string) *Violation {
	if remote == "" {
		return nil
	}

	policies := LoadPolicies()
	if len(policies) == 0 {
		return nil
	}

	for _, p := range policies {
		if matchesRemote(remote, p.RemotePatterns) {
			if !isAllowed(profileName, p.AllowedProfiles) {
				return &Violation{
					Message:         "Profile \"" + profileName + "\" is not allowed by policy \"" + p.Name + "\"",
					AllowedProfiles: p.AllowedProfiles,
					PolicyName:      p.Name,
					Mode:            p.Mode,
				}
			}
		}
	}

	return nil
}

// LoadPolicies loads policies from all sources in precedence order:
// 1. Repo-level (.gitprofile.yml) — highest priority
// 2. User-level (~/.git-profile/policies.yaml)
func LoadPolicies() []Policy {
	var policies []Policy

	// Load user-level policies
	userPolicies := loadUserPolicies()
	policies = append(policies, userPolicies...)

	// Load repo-level policy (overrides user-level if present)
	repoPolicies := loadRepoPolicies()
	policies = append(repoPolicies, policies...) // repo-level first = higher priority

	return policies
}

// loadUserPolicies reads ~/.git-profile/policies.yaml
func loadUserPolicies() []Policy {
	path := config.PoliciesFilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var pf PolicyFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil
	}

	// Default mode to "strict" if not set
	for i := range pf.Policies {
		if pf.Policies[i].Mode == "" {
			pf.Policies[i].Mode = "strict"
		}
	}

	return pf.Policies
}

// loadRepoPolicies reads .gitprofile.yml from the current repo root.
func loadRepoPolicies() []Policy {
	// Find repo root
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		return nil
	}

	path := filepath.Join(repoRoot, config.RepoPolicyFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var rpf RepoPolicyFile
	if err := yaml.Unmarshal(data, &rpf); err != nil {
		return nil
	}

	if len(rpf.AllowedProfiles) == 0 {
		return nil
	}

	mode := rpf.Mode
	if mode == "" {
		mode = "strict"
	}

	name := "repo"
	if rpf.Organization != "" {
		name = rpf.Organization
	}

	// Repo policy matches ALL remotes (it applies to this repo regardless)
	return []Policy{
		{
			Name:            name,
			RemotePatterns:  []string{"*"},
			AllowedProfiles: rpf.AllowedProfiles,
			Mode:            mode,
		},
	}
}

// findRepoRoot walks up from the current directory looking for .git/
func findRepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// matchesRemote checks if a remote URL matches any of the patterns.
func matchesRemote(remote string, patterns []string) bool {
	normalized := NormalizeRemote(remote)
	for _, pattern := range patterns {
		if pattern == "*" {
			return true
		}
		if MatchPattern(normalized, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isAllowed checks if a profile name is in the allowed list.
func isAllowed(profileName string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(a, profileName) {
			return true
		}
	}
	return false
}

// NormalizeRemote strips protocol and normalizes a remote URL.
//
//	"git@github.com:AcmeCorp/repo.git"     → "github.com/acmecorp/repo"
//	"https://github.com/AcmeCorp/repo.git" → "github.com/acmecorp/repo"
func NormalizeRemote(remote string) string {
	r := remote
	if strings.Contains(r, "@") && strings.Contains(r, ":") {
		parts := strings.SplitN(r, "@", 2)
		if len(parts) == 2 {
			r = parts[1]
			r = strings.Replace(r, ":", "/", 1)
		}
	}
	r = strings.TrimPrefix(r, "https://")
	r = strings.TrimPrefix(r, "http://")
	r = strings.TrimSuffix(r, ".git")
	return strings.ToLower(r)
}

// MatchPattern does simple wildcard matching (trailing /* only).
func MatchPattern(value, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(value, prefix+"/")
	}
	return value == pattern
}
