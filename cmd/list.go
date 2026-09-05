package cmd

import (
	"fmt"

	"github.com/Malay-dev/gitx-profile/internal/git"
	"github.com/Malay-dev/gitx-profile/internal/profile"
	"github.com/Malay-dev/gitx-profile/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved profiles",
	Long:  `Display all profiles stored in ~/.gitprofiles. The currently active profile (if any) is marked with an asterisk.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := profile.LoadStore()
		if err != nil {
			return fmt.Errorf("failed to load profiles: %w", err)
		}

		profiles := store.List()
		if len(profiles) == 0 {
			ui.Info("No profiles configured. Run 'git profile add <name>' to create one.")
			return nil
		}

		// Detect current identity for matching
		currentEmail, _ := git.GetConfig("user.email")

		for _, p := range profiles {
			marker := "  "
			if p.UserEmail == currentEmail {
				marker = "* "
			}
			fmt.Printf("%s%s <%s>\n", marker, p.Name, p.UserEmail)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
