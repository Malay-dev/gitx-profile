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

var addName string
var addEmail string
var addSigningKey string
var addSSHKey string
var addFrom string

var addCmd = &cobra.Command{
	Use:   "add <profile-name>",
	Short: "Create a new profile",
	Long: `Add a new named profile with Git identity information.

If --name and --email are not provided, you will be prompted interactively.

Use --from to inherit values from an existing source:
  --from global       Inherit from your global git config (~/.gitconfig)
  --from <profile>    Clone values from an existing profile

When using --from global, you will be asked to confirm the imported values
before the profile is created. You can override any field with explicit flags.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		store, err := profile.LoadStore()
		if err != nil {
			return fmt.Errorf("failed to load profiles: %w", err)
		}

		// Check if profile already exists
		if store.Exists(profileName) {
			ui.Error("Profile %q already exists. Use 'git profile remove %s' first.", profileName, profileName)
			return fmt.Errorf("profile already exists")
		}

		reader := bufio.NewReader(os.Stdin)

		// Handle --from flag
		if addFrom != "" {
			if err := inheritFrom(addFrom, store, reader); err != nil {
				return err
			}
		}

		// Interactive prompts for any remaining empty fields
		if addName == "" {
			fmt.Print("Name: ")
			addName, _ = reader.ReadString('\n')
			addName = strings.TrimSpace(addName)
		}
		if addName == "" {
			return fmt.Errorf("name is required")
		}

		if addEmail == "" {
			fmt.Print("Email: ")
			addEmail, _ = reader.ReadString('\n')
			addEmail = strings.TrimSpace(addEmail)
		}
		if addEmail == "" {
			return fmt.Errorf("email is required")
		}

		if addSigningKey == "" && addFrom == "" {
			fmt.Print("Signing key (optional, press Enter to skip): ")
			addSigningKey, _ = reader.ReadString('\n')
			addSigningKey = strings.TrimSpace(addSigningKey)
		}

		if addSSHKey == "" && addFrom == "" {
			fmt.Print("SSH key path (optional, press Enter to skip): ")
			addSSHKey, _ = reader.ReadString('\n')
			addSSHKey = strings.TrimSpace(addSSHKey)
		}

		p := profile.Profile{
			Name:       profileName,
			UserName:   addName,
			UserEmail:  addEmail,
			SigningKey:  addSigningKey,
			SSHKeyPath: addSSHKey,
		}

		if err := store.Add(p); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}

		ui.Success("Profile %q created", profileName)
		fmt.Println()
		ui.Field("Name", p.UserName)
		ui.Field("Email", p.UserEmail)
		if p.SigningKey != "" {
			ui.Field("Signing Key", p.SigningKey)
		}
		if p.SSHKeyPath != "" {
			ui.Field("SSH Key", p.SSHKeyPath)
		}

		return nil
	},
}

// inheritFrom populates the add flags from an existing source.
// source can be "global" or an existing profile name.
func inheritFrom(source string, store *profile.Store, reader *bufio.Reader) error {
	if strings.EqualFold(source, "global") {
		return inheritFromGlobal(reader)
	}
	return inheritFromProfile(source, store)
}

// inheritFromGlobal reads the global git config and asks for confirmation.
func inheritFromGlobal(reader *bufio.Reader) error {
	globalName, _ := git.GetConfigGlobal("user.name")
	globalEmail, _ := git.GetConfigGlobal("user.email")
	globalSigningKey, _ := git.GetConfigGlobal("user.signingkey")

	if globalName == "" && globalEmail == "" {
		ui.Error("No identity found in global git config (~/.gitconfig).")
		return fmt.Errorf("global config has no identity")
	}

	// Show what will be imported and ask for confirmation
	ui.Warn("Importing identity from global git config (~/.gitconfig)")
	fmt.Println()
	ui.Field("Name", valueOrEmpty(globalName))
	ui.Field("Email", valueOrEmpty(globalEmail))
	if globalSigningKey != "" {
		ui.Field("Signing Key", globalSigningKey)
	}
	fmt.Println()

	fmt.Print("Use these values? (y/N): ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		ui.Info("Cancelled. You can provide values manually or with flags.")
		return fmt.Errorf("user cancelled global import")
	}

	// Only set fields that aren't already overridden by explicit flags
	if addName == "" {
		addName = globalName
	}
	if addEmail == "" {
		addEmail = globalEmail
	}
	if addSigningKey == "" {
		addSigningKey = globalSigningKey
	}

	return nil
}

// inheritFromProfile copies values from an existing profile.
func inheritFromProfile(source string, store *profile.Store) error {
	srcProfile, err := store.Get(source)
	if err != nil {
		ui.Error("Source profile %q not found.", source)
		suggestions := store.SuggestSimilar(source)
		if len(suggestions) > 0 {
			ui.Info("Did you mean: %s?", suggestions[0])
		}
		return fmt.Errorf("source profile not found: %s", source)
	}

	ui.Info("Cloning values from profile %q", source)
	fmt.Println()
	ui.Field("Name", srcProfile.UserName)
	ui.Field("Email", srcProfile.UserEmail)
	if srcProfile.SigningKey != "" {
		ui.Field("Signing Key", srcProfile.SigningKey)
	}
	if srcProfile.SSHKeyPath != "" {
		ui.Field("SSH Key", srcProfile.SSHKeyPath)
	}
	fmt.Println()

	// Only set fields that aren't already overridden by explicit flags
	if addName == "" {
		addName = srcProfile.UserName
	}
	if addEmail == "" {
		addEmail = srcProfile.UserEmail
	}
	if addSigningKey == "" {
		addSigningKey = srcProfile.SigningKey
	}
	if addSSHKey == "" {
		addSSHKey = srcProfile.SSHKeyPath
	}

	return nil
}

func valueOrEmpty(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

func init() {
	addCmd.Flags().StringVar(&addName, "name", "", "Git user.name for this profile")
	addCmd.Flags().StringVar(&addEmail, "email", "", "Git user.email for this profile")
	addCmd.Flags().StringVar(&addSigningKey, "signing-key", "", "GPG/SSH signing key (optional)")
	addCmd.Flags().StringVar(&addSSHKey, "ssh-key", "", "Path to SSH private key (optional)")
	addCmd.Flags().StringVar(&addFrom, "from", "", "Inherit values from 'global' config or an existing profile name")
	rootCmd.AddCommand(addCmd)
}
