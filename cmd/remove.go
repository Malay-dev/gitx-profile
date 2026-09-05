package cmd

import (
	"fmt"

	"github.com/Malay-dev/gitx-profile/internal/profile"
	"github.com/Malay-dev/gitx-profile/internal/ui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <profile-name>",
	Aliases: []string{"rm"},
	Short:   "Remove a saved profile",
	Long:    `Delete a profile from ~/.gitprofiles. This does not change any existing git config.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		store, err := profile.LoadStore()
		if err != nil {
			return fmt.Errorf("failed to load profiles: %w", err)
		}

		if !store.Exists(profileName) {
			ui.Error("Profile %q not found.", profileName)
			return fmt.Errorf("profile not found")
		}

		if err := store.Remove(profileName); err != nil {
			return fmt.Errorf("failed to remove profile: %w", err)
		}

		ui.Success("Profile %q removed.", profileName)
		ui.Info("Note: Existing git configs using this identity are unchanged.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
