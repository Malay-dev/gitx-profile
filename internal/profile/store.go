package profile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Malay-dev/gitx-profile/internal/config"
)

// Store manages profile persistence and lookup.
type Store struct {
	profiles []Profile
	filePath string
}

// LoadStore reads profiles from the config file (~/.gitprofiles).
func LoadStore() (*Store, error) {
	path := config.ProfilesFilePath()

	s := &Store{
		filePath: path,
		profiles: []Profile{},
	}

	// If file doesn't exist, return empty store
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return s, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var current *Profile

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header: [profile "name"]
		if strings.HasPrefix(line, "[profile") && strings.HasSuffix(line, "]") {
			if current != nil {
				s.profiles = append(s.profiles, *current)
			}
			name := extractProfileName(line)
			current = &Profile{Name: name}
			continue
		}

		// Key-value pairs
		if current != nil {
			key, value := parseKeyValue(line)
			switch key {
			case "name":
				current.UserName = value
			case "email":
				current.UserEmail = value
			case "signingkey":
				current.SigningKey = value
			case "sshkey":
				current.SSHKeyPath = value
			}
		}
	}

	// Don't forget the last profile
	if current != nil {
		s.profiles = append(s.profiles, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	return s, nil
}

// List returns all profiles.
func (s *Store) List() []Profile {
	return s.profiles
}

// Get returns a profile by name.
func (s *Store) Get(name string) (*Profile, error) {
	for i := range s.profiles {
		if strings.EqualFold(s.profiles[i].Name, name) {
			return &s.profiles[i], nil
		}
	}
	return nil, fmt.Errorf("profile %q not found", name)
}

// Exists checks if a profile name is taken.
func (s *Store) Exists(name string) bool {
	_, err := s.Get(name)
	return err == nil
}

// Add appends a new profile and saves to disk.
func (s *Store) Add(p Profile) error {
	if s.Exists(p.Name) {
		return fmt.Errorf("profile %q already exists", p.Name)
	}
	s.profiles = append(s.profiles, p)
	return s.save()
}

// Remove deletes a profile by name and saves to disk.
func (s *Store) Remove(name string) error {
	idx := -1
	for i := range s.profiles {
		if strings.EqualFold(s.profiles[i].Name, name) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("profile %q not found", name)
	}
	s.profiles = append(s.profiles[:idx], s.profiles[idx+1:]...)
	return s.save()
}

// MatchByEmail finds the first profile matching the given email.
func (s *Store) MatchByEmail(email string) *Profile {
	for i := range s.profiles {
		if strings.EqualFold(s.profiles[i].UserEmail, email) {
			return &s.profiles[i]
		}
	}
	return nil
}

// SuggestSimilar returns profile names that are similar to the input.
// Simple implementation: prefix match or contains.
func (s *Store) SuggestSimilar(name string) []string {
	var suggestions []string
	lower := strings.ToLower(name)
	for _, p := range s.profiles {
		pLower := strings.ToLower(p.Name)
		if strings.HasPrefix(pLower, lower) || strings.Contains(pLower, lower) {
			suggestions = append(suggestions, p.Name)
		}
	}
	return suggestions
}

// SuggestByEmail returns profile names as suggestions when email doesn't match.
func (s *Store) SuggestByEmail(email string) []string {
	var suggestions []string
	for _, p := range s.profiles {
		suggestions = append(suggestions, p.Name)
	}
	return suggestions
}

// save writes all profiles back to disk in INI format.
func (s *Store) save() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", dir, err)
	}

	file, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("cannot write %s: %w", s.filePath, err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for i, p := range s.profiles {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "[profile %q]\n", p.Name)
		fmt.Fprintf(w, "  name = %s\n", p.UserName)
		fmt.Fprintf(w, "  email = %s\n", p.UserEmail)
		if p.SigningKey != "" {
			fmt.Fprintf(w, "  signingkey = %s\n", p.SigningKey)
		}
		if p.SSHKeyPath != "" {
			fmt.Fprintf(w, "  sshkey = %s\n", p.SSHKeyPath)
		}
	}

	return w.Flush()
}

// extractProfileName parses: [profile "work"] → "work"
func extractProfileName(line string) string {
	line = strings.TrimPrefix(line, "[profile")
	line = strings.TrimSuffix(line, "]")
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "\"")
	return line
}

// parseKeyValue splits "key = value" into (key, value).
func parseKeyValue(line string) (string, string) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(strings.ToLower(parts[0])), strings.TrimSpace(parts[1])
}
