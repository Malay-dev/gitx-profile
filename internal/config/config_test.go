package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfilesFilePath_Default(t *testing.T) {
	// Clear any override
	t.Setenv("GITX_PROFILES_PATH", "")

	path := ProfilesFilePath()

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".gitprofiles")

	if path != expected {
		t.Errorf("ProfilesFilePath() = %q, want %q", path, expected)
	}
}

func TestProfilesFilePath_EnvOverride(t *testing.T) {
	t.Setenv("GITX_PROFILES_PATH", "/custom/path/.gitprofiles")

	path := ProfilesFilePath()
	if path != "/custom/path/.gitprofiles" {
		t.Errorf("ProfilesFilePath() = %q, want '/custom/path/.gitprofiles'", path)
	}
}

func TestPoliciesFilePath_Default(t *testing.T) {
	t.Setenv("GITX_POLICIES_PATH", "")

	path := PoliciesFilePath()

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".git-profile", "policies.yaml")

	if path != expected {
		t.Errorf("PoliciesFilePath() = %q, want %q", path, expected)
	}
}

func TestPoliciesFilePath_EnvOverride(t *testing.T) {
	t.Setenv("GITX_POLICIES_PATH", "/tmp/my-policies.yaml")

	path := PoliciesFilePath()
	if path != "/tmp/my-policies.yaml" {
		t.Errorf("PoliciesFilePath() = %q, want '/tmp/my-policies.yaml'", path)
	}
}

func TestPoliciesDir_Default(t *testing.T) {
	path := PoliciesDir()

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".git-profile")

	if path != expected {
		t.Errorf("PoliciesDir() = %q, want %q", path, expected)
	}
}

func TestRepoPolicyFileName(t *testing.T) {
	if RepoPolicyFileName != ".gitprofile.yml" {
		t.Errorf("RepoPolicyFileName = %q, want '.gitprofile.yml'", RepoPolicyFileName)
	}
}

func TestProfilesFilePath_ContainsHomeDir(t *testing.T) {
	t.Setenv("GITX_PROFILES_PATH", "")

	path := ProfilesFilePath()
	home, _ := os.UserHomeDir()

	if !strings.HasPrefix(path, home) {
		t.Errorf("ProfilesFilePath() = %q, expected to start with home dir %q", path, home)
	}
}
