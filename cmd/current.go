package cmd

import (
	"fmt"

	"github.com/Malay-dev/gitx-profile/internal/git"
	"github.com/Malay-dev/gitx-profile/internal/profile"
	"github.com/Malay-dev/gitx-profile/internal/ui"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the active Git identity",
	Long:  `Display the currently active Git identity and match it against known profiles.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := git.GetConfig("user.name")
		email, _ := git.GetConfig("user.email")

		if name == "" && email == "" {
			ui.Warn("No Git identity configured in current scope.")
			return nil
		}

		// Try to match against known profiles
		store, err := profile.LoadStore()
		if err != nil {
			// Non-fatal — still show raw config
			ui.Info("Active identity (no profile match):")
			ui.Field("Name", name)
			ui.Field("Email", email)
			return nil
		}

		matched := store.MatchByEmail(email)

		if matched != nil {
			ui.Success("Profile: %s", matched.Name)
		} else {
			ui.Warn("No matching profile found")
		}
		fmt.Println()
		ui.Field("Name", name)
		ui.Field("Email", email)

		// Show scope info
		if git.IsInsideWorkTree() {
			repoName, _ := git.GetRepoName()
			if repoName != "" {
				ui.Field("Repository", repoName)
			}

			// Check if identity is local or inherited
			localEmail, _ := git.GetConfigLocal("user.email")
			if localEmail != "" {
				ui.Field("Scope", "Local (.git/config)")
			} else {
				ui.Field("Scope", "Inherited (global)")
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
