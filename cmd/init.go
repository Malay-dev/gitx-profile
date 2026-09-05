package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Malay-dev/gitx-profile/internal/git"
	"github.com/Malay-dev/gitx-profile/internal/profile"
	"github.com/Malay-dev/gitx-profile/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up a profile for the current repository",
	Long: `Interactively configure the identity for the current repository.

This command inspects the repo's current config and remote, then:
  - If a matching profile exists → offers to apply it
  - If the current identity is unrecognized → offers to save it as a new profile
  - If no identity is set → prompts you to pick or create one

This is the recommended way to set up a repo after cloning.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.IsInsideWorkTree() {
			ui.Error("Not inside a Git repository.")
			return fmt.Errorf("not a git repository")
		}

		reader := bufio.NewReader(os.Stdin)

		store, err := profile.LoadStore()
		if err != nil {
			return fmt.Errorf("failed to load profiles: %w", err)
		}

		repoName, _ := git.GetRepoName()
		remote, _ := git.GetRemoteURL("origin")
		currentName, _ := git.GetConfig("user.name")
		currentEmail, _ := git.GetConfig("user.email")

		fmt.Printf("Repository: %s\n", repoName)
		if remote != "" {
			fmt.Printf("Remote:     %s\n", remote)
		}
		fmt.Println()

		profiles := store.List()

		// Case 1: Current email already matches a profile
		if currentEmail != "" {
			matched := store.MatchByEmail(currentEmail)
			if matched != nil {
				ui.Success("This repo already uses profile %q (%s)", matched.Name, matched.UserEmail)

				// Check if it's set locally or just inherited
				localEmail, _ := git.GetConfigLocal("user.email")
				if localEmail == "" {
					fmt.Println()
					ui.Info("Identity is inherited from global config (not pinned to this repo).")
					fmt.Printf("Pin profile %q to this repo? (y/N): ", matched.Name)
					answer, _ := reader.ReadString('\n')
					answer = strings.TrimSpace(strings.ToLower(answer))
					if answer == "y" || answer == "yes" {
						return applyProfile(matched, git.ScopeLocal, repoName)
					}
				}
				return nil
			}
		}

		// Case 2: Identity exists but doesn't match any profile
		if currentEmail != "" {
			ui.Warn("Current identity doesn't match any saved profile.")
			fmt.Println()
			ui.Field("Name", currentName)
			ui.Field("Email", currentEmail)
			fmt.Println()

			fmt.Println("What would you like to do?")
			fmt.Println("  [1] Save current identity as a new profile")
			if len(profiles) > 0 {
				fmt.Println("  [2] Apply an existing profile instead")
			}
			fmt.Println("  [n] Skip")
			fmt.Println()
			fmt.Print("Choice: ")

			choice, _ := reader.ReadString('\n')
			choice = strings.TrimSpace(choice)

			switch choice {
			case "1":
				return saveCurrentAsProfile(store, currentName, currentEmail, reader)
			case "2":
				if len(profiles) > 0 {
					return pickAndApply(store, profiles, reader, repoName)
				}
				ui.Info("No profiles available. Create one first with 'git profile add'.")
				return nil
			default:
				ui.Info("Skipped.")
				return nil
			}
		}

		// Case 3: No identity configured at all
		ui.Warn("No Git identity configured for this repository.")
		fmt.Println()

		if len(profiles) == 0 {
			ui.Info("No profiles saved yet. Let's create one.")
			fmt.Println()
			return createProfileInteractive(store, reader, repoName)
		}

		fmt.Println("Pick a profile to apply, or create a new one:")
		fmt.Println()
		for i, p := range profiles {
			fmt.Printf("  [%d] %s <%s>\n", i+1, p.Name, p.UserEmail)
		}
		fmt.Printf("  [n] Create new profile\n")
		fmt.Println()
		fmt.Print("Choice: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		if choice == "n" || choice == "new" {
			return createProfileInteractive(store, reader, repoName)
		}

		// Parse numeric choice
		var idx int
		_, err = fmt.Sscanf(choice, "%d", &idx)
		if err != nil || idx < 1 || idx > len(profiles) {
			ui.Error("Invalid choice.")
			return fmt.Errorf("invalid choice")
		}

		selected := &profiles[idx-1]
		return applyProfile(selected, git.ScopeLocal, repoName)
	},
}

// saveCurrentAsProfile saves the repo's current identity as a new profile.
func saveCurrentAsProfile(store *profile.Store, name, email string, reader *bufio.Reader) error {
	fmt.Print("Profile name: ")
	profileName, _ := reader.ReadString('\n')
	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		return fmt.Errorf("profile name is required")
	}

	if store.Exists(profileName) {
		ui.Error("Profile %q already exists.", profileName)
		return fmt.Errorf("profile already exists")
	}

	p := profile.Profile{
		Name:      profileName,
		UserName:  name,
		UserEmail: email,
	}

	if err := store.Add(p); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	ui.Success("Profile %q created from current identity", profileName)
	fmt.Println()
	ui.Field("Name", p.UserName)
	ui.Field("Email", p.UserEmail)

	return nil
}

// pickAndApply shows a list and lets the user choose.
func pickAndApply(store *profile.Store, profiles []profile.Profile, reader *bufio.Reader, repoName string) error {
	fmt.Println()
	for i, p := range profiles {
		fmt.Printf("  [%d] %s <%s>\n", i+1, p.Name, p.UserEmail)
	}
	fmt.Println()
	fmt.Print("Choice: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var idx int
	_, err := fmt.Sscanf(choice, "%d", &idx)
	if err != nil || idx < 1 || idx > len(profiles) {
		ui.Error("Invalid choice.")
		return fmt.Errorf("invalid choice")
	}

	selected := &profiles[idx-1]
	return applyProfile(selected, git.ScopeLocal, repoName)
}

// createProfileInteractive walks through creating a new profile and applying it.
func createProfileInteractive(store *profile.Store, reader *bufio.Reader, repoName string) error {
	fmt.Print("Profile name: ")
	profileName, _ := reader.ReadString('\n')
	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		return fmt.Errorf("profile name is required")
	}

	fmt.Print("Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}

	p := profile.Profile{
		Name:      profileName,
		UserName:  name,
		UserEmail: email,
	}

	if err := store.Add(p); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	ui.Success("Profile %q created", profileName)

	// Apply immediately
	return applyProfile(&p, git.ScopeLocal, repoName)
}

// applyProfile writes a profile to git config and confirms.
func applyProfile(p *profile.Profile, scope git.Scope, repoName string) error {
	if err := git.SetConfig("user.name", p.UserName, scope); err != nil {
		return fmt.Errorf("failed to set user.name: %w", err)
	}
	if err := git.SetConfig("user.email", p.UserEmail, scope); err != nil {
		return fmt.Errorf("failed to set user.email: %w", err)
	}
	if p.SigningKey != "" {
		if err := git.SetConfig("user.signingkey", p.SigningKey, scope); err != nil {
			return fmt.Errorf("failed to set user.signingkey: %w", err)
		}
	}

	fmt.Println()
	ui.Success("Applied profile %q to repository", p.Name)
	ui.Field("Repository", repoName)
	ui.Field("Name", p.UserName)
	ui.Field("Email", p.UserEmail)

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
