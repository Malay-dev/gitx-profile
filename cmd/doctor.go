package cmd

import (
	"fmt"

	"github.com/Malay-dev/gitx-profile/internal/git"
	"github.com/Malay-dev/gitx-profile/internal/profile"
	"github.com/Malay-dev/gitx-profile/internal/ui"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose your Git identity configuration",
	Long: `Run a health check on your Git identity setup.

Shows:
  - Current identity and its source
  - Profile match status
  - Remote information
  - Potential issues`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("=== git-profile doctor ===")
		fmt.Println()

		// Git version
		gitVersion, err := git.Version()
		if err != nil {
			ui.Error("Git not found: %v", err)
			return err
		}
		ui.Field("Git Version", gitVersion)
		fmt.Println()

		// Repository status
		fmt.Println("--- Repository ---")
		if git.IsInsideWorkTree() {
			repoName, _ := git.GetRepoName()
			ui.Field("Repository", repoName)

			remote, err := git.GetRemoteURL("origin")
			if err == nil && remote != "" {
				ui.Field("Remote (origin)", remote)
			} else {
				ui.Field("Remote (origin)", "(not set)")
			}
		} else {
			ui.Info("Not inside a Git repository")
		}
		fmt.Println()

		// Current identity
		fmt.Println("--- Identity ---")
		name, _ := git.GetConfig("user.name")
		email, _ := git.GetConfig("user.email")
		ui.Field("user.name", valueOrNone(name))
		ui.Field("user.email", valueOrNone(email))

		if git.IsInsideWorkTree() {
			localName, _ := git.GetConfigLocal("user.name")
			localEmail, _ := git.GetConfigLocal("user.email")
			if localName != "" || localEmail != "" {
				ui.Field("Source", "Local (.git/config)")
			} else {
				ui.Field("Source", "Global (~/.gitconfig)")
			}
		}
		fmt.Println()

		// Profile matching
		fmt.Println("--- Profile Match ---")
		store, err := profile.LoadStore()
		if err != nil {
			ui.Warn("Could not load profiles: %v", err)
		} else {
			profiles := store.List()
			ui.Field("Profiles Found", fmt.Sprintf("%d", len(profiles)))

			if email != "" {
				matched := store.MatchByEmail(email)
				if matched != nil {
					ui.Success("Matched profile: %s", matched.Name)
				} else {
					ui.Warn("No profile matches current email %q", email)
				}
			}
		}
		fmt.Println()

		// Signing
		fmt.Println("--- Signing ---")
		signingKey, _ := git.GetConfig("user.signingkey")
		gpgSign, _ := git.GetConfig("commit.gpgsign")
		ui.Field("user.signingkey", valueOrNone(signingKey))
		ui.Field("commit.gpgsign", valueOrNone(gpgSign))

		return nil
	},
}

func valueOrNone(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
