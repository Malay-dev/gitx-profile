package cmd

import (
	"fmt"

	"github.com/Malay-dev/gitx-profile/internal/git"
	"github.com/Malay-dev/gitx-profile/internal/policy"
	"github.com/Malay-dev/gitx-profile/internal/profile"
	"github.com/Malay-dev/gitx-profile/internal/ui"
	"github.com/spf13/cobra"
)

var useGlobal bool
var useForce bool

var useCmd = &cobra.Command{
	Use:   "use <profile-name>",
	Short: "Switch to a named profile",
	Long: `Apply a profile's identity to the current repository (default)
or globally with --global.

The profile's user.name and user.email will be written to
the appropriate git config scope.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		// Load profiles
		store, err := profile.LoadStore()
		if err != nil {
			return fmt.Errorf("failed to load profiles: %w", err)
		}

		// Find the requested profile
		p, err := store.Get(profileName)
		if err != nil {
			// Suggest similar names
			suggestions := store.SuggestSimilar(profileName)
			if len(suggestions) > 0 {
				ui.Error("Profile %q not found. Did you mean: %s?", profileName, suggestions[0])
			} else {
				ui.Error("Profile %q not found. Run 'git profile list' to see available profiles.", profileName)
			}
			return err
		}

		// Determine scope
		scope := git.ScopeLocal
		if useGlobal {
			scope = git.ScopeGlobal
		}

		// Check if we're in a git repo (required for local scope)
		if scope == git.ScopeLocal && !git.IsInsideWorkTree() {
			ui.Warn("Not inside a Git repository.")
			ui.Info("Use --global to apply globally, or cd into a repository.")
			return fmt.Errorf("not a git repository")
		}

		// Evaluate policies (unless --force)
		if !useForce && scope == git.ScopeLocal {
			remote, _ := git.GetRemoteURL("origin")
			violation := policy.Evaluate(profileName, remote)
			if violation != nil {
				ui.Error("Policy violation: %s", violation.Message)
				ui.Info("Allowed profiles for this repository: %v", violation.AllowedProfiles)
				ui.Info("Use --force to override.")
				return fmt.Errorf("policy violation")
			}
		}

		// Apply the profile
		if err := git.SetConfig("user.name", p.UserName, scope); err != nil {
			return fmt.Errorf("failed to set user.name: %w", err)
		}
		if err := git.SetConfig("user.email", p.UserEmail, scope); err != nil {
			return fmt.Errorf("failed to set user.email: %w", err)
		}

		// Set signing key if present
		if p.SigningKey != "" {
			if err := git.SetConfig("user.signingkey", p.SigningKey, scope); err != nil {
				return fmt.Errorf("failed to set user.signingkey: %w", err)
			}
		}

		// Display result
		scopeLabel := "Local"
		repoName := ""
		if scope == git.ScopeLocal {
			repoName, _ = git.GetRepoName()
		}
		if scope == git.ScopeGlobal {
			scopeLabel = "Global"
		}

		ui.Success("Switched to profile %q", profileName)
		fmt.Println()
		ui.Field("Scope", scopeLabel)
		if repoName != "" {
			ui.Field("Repository", repoName)
		}
		ui.Field("Name", p.UserName)
		ui.Field("Email", p.UserEmail)

		return nil
	},
}

func init() {
	useCmd.Flags().BoolVar(&useGlobal, "global", false, "Apply profile to global git config")
	useCmd.Flags().BoolVar(&useForce, "force", false, "Override policy restrictions")
	rootCmd.AddCommand(useCmd)
}
