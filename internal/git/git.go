package git

import (
	"os/exec"
	"strings"
)

// Scope represents where git config is written.
type Scope int

const (
	ScopeLocal  Scope = iota // .git/config
	ScopeGlobal              // ~/.gitconfig
)

// IsInsideWorkTree returns true if the current directory is inside a git repo.
func IsInsideWorkTree() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// GetConfig reads a git config value (effective value from all scopes).
func GetConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", key)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetConfigLocal reads a git config value from local repo scope only.
func GetConfigLocal(key string) (string, error) {
	cmd := exec.Command("git", "config", "--local", key)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetConfigGlobal reads a git config value from global scope only (~/.gitconfig).
func GetConfigGlobal(key string) (string, error) {
	cmd := exec.Command("git", "config", "--global", key)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// SetConfig writes a git config value at the specified scope.
func SetConfig(key, value string, scope Scope) error {
	args := []string{"config"}
	if scope == ScopeGlobal {
		args = append(args, "--global")
	}
	args = append(args, key, value)

	cmd := exec.Command("git", args...)
	return cmd.Run()
}

// GetRemoteURL returns the URL of the named remote.
func GetRemoteURL(name string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", name)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRepoName returns the name of the current repository (basename of root).
func GetRepoName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	topLevel := strings.TrimSpace(string(out))
	parts := strings.Split(topLevel, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}
	return topLevel, nil
}

// GetCurrentBranch returns the current branch name.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Version returns the installed git version string.
func Version() (string, error) {
	cmd := exec.Command("git", "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
