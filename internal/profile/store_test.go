package profile

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Store tests ---

func TestLoadStore_EmptyFile(t *testing.T) {
	// Create a temp profiles file
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitprofiles")
	os.WriteFile(path, []byte(""), 0o644)

	t.Setenv("GITX_PROFILES_PATH", path)

	store, err := LoadStore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.List()) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(store.List()))
	}
}

func TestLoadStore_FileNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent")

	t.Setenv("GITX_PROFILES_PATH", path)

	store, err := LoadStore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.List()) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(store.List()))
	}
}

func TestLoadStore_ParsesProfiles(t *testing.T) {
	content := `[profile "work"]
  name = Malay Kumar
  email = malay@company.com
  signingkey = ABC123

[profile "personal"]
  name = Malay
  email = malay@gmail.com
  sshkey = ~/.ssh/personal
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitprofiles")
	os.WriteFile(path, []byte(content), 0o644)

	t.Setenv("GITX_PROFILES_PATH", path)

	store, err := LoadStore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profiles := store.List()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	// First profile
	if profiles[0].Name != "work" {
		t.Errorf("expected name 'work', got %q", profiles[0].Name)
	}
	if profiles[0].UserName != "Malay Kumar" {
		t.Errorf("expected userName 'Malay Kumar', got %q", profiles[0].UserName)
	}
	if profiles[0].UserEmail != "malay@company.com" {
		t.Errorf("expected email 'malay@company.com', got %q", profiles[0].UserEmail)
	}
	if profiles[0].SigningKey != "ABC123" {
		t.Errorf("expected signingKey 'ABC123', got %q", profiles[0].SigningKey)
	}

	// Second profile
	if profiles[1].Name != "personal" {
		t.Errorf("expected name 'personal', got %q", profiles[1].Name)
	}
	if profiles[1].SSHKeyPath != "~/.ssh/personal" {
		t.Errorf("expected sshkey '~/.ssh/personal', got %q", profiles[1].SSHKeyPath)
	}
}

func TestStore_Get(t *testing.T) {
	store := &Store{
		profiles: []Profile{
			{Name: "work", UserName: "Malay", UserEmail: "malay@company.com"},
			{Name: "personal", UserName: "Malay", UserEmail: "malay@gmail.com"},
		},
	}

	p, err := store.Get("work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.UserEmail != "malay@company.com" {
		t.Errorf("expected email 'malay@company.com', got %q", p.UserEmail)
	}

	// Case-insensitive
	p, err = store.Get("WORK")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "work" {
		t.Errorf("expected name 'work', got %q", p.Name)
	}

	// Not found
	_, err = store.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestStore_Exists(t *testing.T) {
	store := &Store{
		profiles: []Profile{
			{Name: "work"},
		},
	}

	if !store.Exists("work") {
		t.Error("expected 'work' to exist")
	}
	if store.Exists("nope") {
		t.Error("expected 'nope' to not exist")
	}
}

func TestStore_AddAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitprofiles")

	t.Setenv("GITX_PROFILES_PATH", path)

	store, _ := LoadStore()

	// Add
	err := store.Add(Profile{Name: "test", UserName: "Test", UserEmail: "test@test.com"})
	if err != nil {
		t.Fatalf("unexpected error adding: %v", err)
	}
	if !store.Exists("test") {
		t.Error("expected profile 'test' to exist after add")
	}

	// Add duplicate
	err = store.Add(Profile{Name: "test", UserName: "Test", UserEmail: "test@test.com"})
	if err == nil {
		t.Error("expected error adding duplicate profile")
	}

	// Remove
	err = store.Remove("test")
	if err != nil {
		t.Fatalf("unexpected error removing: %v", err)
	}
	if store.Exists("test") {
		t.Error("expected profile 'test' to not exist after remove")
	}

	// Remove nonexistent
	err = store.Remove("ghost")
	if err == nil {
		t.Error("expected error removing nonexistent profile")
	}
}

func TestStore_MatchByEmail(t *testing.T) {
	store := &Store{
		profiles: []Profile{
			{Name: "work", UserEmail: "malay@company.com"},
			{Name: "personal", UserEmail: "malay@gmail.com"},
		},
	}

	p := store.MatchByEmail("malay@company.com")
	if p == nil || p.Name != "work" {
		t.Error("expected match for work profile")
	}

	p = store.MatchByEmail("MALAY@COMPANY.COM")
	if p == nil || p.Name != "work" {
		t.Error("expected case-insensitive match")
	}

	p = store.MatchByEmail("unknown@email.com")
	if p != nil {
		t.Error("expected no match for unknown email")
	}
}

func TestStore_SuggestSimilar(t *testing.T) {
	store := &Store{
		profiles: []Profile{
			{Name: "work"},
			{Name: "personal"},
			{Name: "workspace"},
		},
	}

	suggestions := store.SuggestSimilar("wor")
	if len(suggestions) == 0 {
		t.Fatal("expected suggestions for 'wor'")
	}
	// Should match "work" and "workspace"
	found := false
	for _, s := range suggestions {
		if s == "work" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'work' in suggestions, got %v", suggestions)
	}
}

func TestStore_PersistenceRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitprofiles")

	t.Setenv("GITX_PROFILES_PATH", path)

	// Write
	store, _ := LoadStore()
	store.Add(Profile{
		Name:       "work",
		UserName:   "Malay Kumar",
		UserEmail:  "malay@company.com",
		SigningKey:  "DEADBEEF",
		SSHKeyPath: "~/.ssh/work",
	})
	store.Add(Profile{
		Name:      "personal",
		UserName:  "Malay",
		UserEmail: "malay@gmail.com",
	})

	// Reload
	store2, err := LoadStore()
	if err != nil {
		t.Fatalf("unexpected error reloading: %v", err)
	}

	profiles := store2.List()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles after reload, got %d", len(profiles))
	}

	if profiles[0].SigningKey != "DEADBEEF" {
		t.Errorf("expected signingKey 'DEADBEEF', got %q", profiles[0].SigningKey)
	}
	if profiles[0].SSHKeyPath != "~/.ssh/work" {
		t.Errorf("expected sshkey '~/.ssh/work', got %q", profiles[0].SSHKeyPath)
	}
	if profiles[1].Name != "personal" {
		t.Errorf("expected second profile 'personal', got %q", profiles[1].Name)
	}
}

// --- INI parser helper tests ---

func TestExtractProfileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`[profile "work"]`, "work"},
		{`[profile "personal"]`, "personal"},
		{`[profile "my-org"]`, "my-org"},
	}

	for _, tt := range tests {
		got := extractProfileName(tt.input)
		if got != tt.expected {
			t.Errorf("extractProfileName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseKeyValue(t *testing.T) {
	tests := []struct {
		input string
		key   string
		value string
	}{
		{"  name = Malay Kumar", "name", "Malay Kumar"},
		{"email = malay@test.com", "email", "malay@test.com"},
		{"  signingkey = ABC123", "signingkey", "ABC123"},
		{"invalid line", "", ""},
	}

	for _, tt := range tests {
		k, v := parseKeyValue(tt.input)
		if k != tt.key || v != tt.value {
			t.Errorf("parseKeyValue(%q) = (%q, %q), want (%q, %q)", tt.input, k, v, tt.key, tt.value)
		}
	}
}

func TestStore_SkipsCommentsAndEmptyLines(t *testing.T) {
	content := `# This is a comment

[profile "work"]
  name = Malay
  email = malay@company.com

# Another comment
[profile "personal"]
  name = Personal
  email = personal@gmail.com
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitprofiles")
	os.WriteFile(path, []byte(content), 0o644)

	t.Setenv("GITX_PROFILES_PATH", path)

	store, err := LoadStore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.List()) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(store.List()))
	}
}
