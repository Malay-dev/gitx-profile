package config

import (
	"os"
	"path/filepath"
)

// ProfilesFilePath returns the path to the profiles store file.
// Default: ~/.gitprofiles
func ProfilesFilePath() string {
	if env := os.Getenv("GITX_PROFILES_PATH"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gitprofiles"
	}
	return filepath.Join(home, ".gitprofiles")
}

// PoliciesFilePath returns the path to the user-level policies file.
// Default: ~/.git-profile/policies.yaml
func PoliciesFilePath() string {
	if env := os.Getenv("GITX_POLICIES_PATH"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".git-profile/policies.yaml"
	}
	return filepath.Join(home, ".git-profile", "policies.yaml")
}

// PoliciesDir returns the directory containing policy files.
// Default: ~/.git-profile/
func PoliciesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".git-profile"
	}
	return filepath.Join(home, ".git-profile")
}

// RepoPolicyFileName is the name of the per-repository policy file.
const RepoPolicyFileName = ".gitprofile.yml"
