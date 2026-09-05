package cmd

import (
	"fmt"
	"os"

	"github.com/Malay-dev/gitx-profile/internal/git"
	"github.com/Malay-dev/gitx-profile/internal/profile"
	"github.com/Malay-dev/gitx-profile/internal/ui"
	"github.com/spf13/cobra"
)

var checkQuiet bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify the current identity matches a known profile",
	Long: `Check whether the current Git identity matches any known profile.
Also evaluates policies if configured. Useful in pre-commit hooks.

Exit codes:
  0 — identity matches a known profile
  1 — no match or policy violation`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := git.GetConfig("user.email")
		if email == "" {
			if !checkQuiet {
				ui.Error("No Git identity configured.")
			}
			os.Exit(1)
		}

		store, err := profile.LoadStore()
		if err != nil {
			if !checkQuiet {
				ui.Error("Failed to load profiles: %v", err)
			}
			os.Exit(1)
		}

		matched := store.MatchByEmail(email)
		if matched == nil {
			if !checkQuiet {
				ui.Error("Current email %q does not match any profile.", email)
				// Suggest closest match
				suggestions := store.SuggestByEmail(email)
				if len(suggestions) > 0 {
					ui.Info("Suggested: git profile use %s", suggestions[0])
				}
			}
			os.Exit(1)
		}

		if !checkQuiet {
			ui.Success("Identity matches profile %q", matched.Name)
			fmt.Println()
			ui.Field("Name", matched.UserName)
			ui.Field("Email", matched.UserEmail)
		}

		return nil
	},
}

func init() {
	checkCmd.Flags().BoolVar(&checkQuiet, "quiet", false, "Suppress output (exit code only)")
	rootCmd.AddCommand(checkCmd)
}
