package cmd

import (
	"fmt"

	"github.com/Malay-dev/gitx-profile/internal/git"
	"github.com/Malay-dev/gitx-profile/internal/policy"
	"github.com/Malay-dev/gitx-profile/internal/profile"
	"github.com/Malay-dev/gitx-profile/internal/ui"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Discover the identity configuration of the current repository",
	Long: `Inspect the current repository's Git identity setup.

Shows:
  - Repository name and remote
  - Active identity (name, email, signing key)
  - Where the identity is configured (local vs global)
  - Matched profile (if any)
  - Policy evaluation result
  - Suggested profile based on remote URL

This is useful when you cd into a repo and want to know
"who am I here and is it correct?"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.IsInsideWorkTree() {
			ui.Error("Not inside a Git repository.")
			ui.Info("cd into a repository and run this command again.")
			return fmt.Errorf("not a git repository")
		}

		// --- Repository ---
		fmt.Println("=== Repository ===")
		fmt.Println()

		repoName, _ := git.GetRepoName()
		ui.Field("Name", repoName)

		remote, _ := git.GetRemoteURL("origin")
		if remote != "" {
			ui.Field("Remote (origin)", remote)
		} else {
			ui.Field("Remote (origin)", "(not configured)")
		}

		branch, _ := git.GetCurrentBranch()
		if branch != "" {
			ui.Field("Branch", branch)
		}
		fmt.Println()

		// --- Identity ---
		fmt.Println("=== Identity ===")
		fmt.Println()

		// Effective values
		name, _ := git.GetConfig("user.name")
		email, _ := git.GetConfig("user.email")
		signingKey, _ := git.GetConfig("user.signingkey")

		ui.Field("user.name", valueOrNotSet(name))
		ui.Field("user.email", valueOrNotSet(email))
		if signingKey != "" {
			ui.Field("user.signingkey", signingKey)
		}

		// Source detection
		localName, _ := git.GetConfigLocal("user.name")
		localEmail, _ := git.GetConfigLocal("user.email")
		globalName, _ := git.GetConfigGlobal("user.name")
		globalEmail, _ := git.GetConfigGlobal("user.email")

		fmt.Println()
		fmt.Println("  Source breakdown:")
		if localName != "" {
			ui.Field("  user.name", fmt.Sprintf("%s (local — .git/config)", localName))
		} else if globalName != "" {
			ui.Field("  user.name", fmt.Sprintf("%s (global — ~/.gitconfig)", globalName))
		} else {
			ui.Field("  user.name", "(not set anywhere)")
		}

		if localEmail != "" {
			ui.Field("  user.email", fmt.Sprintf("%s (local — .git/config)", localEmail))
		} else if globalEmail != "" {
			ui.Field("  user.email", fmt.Sprintf("%s (global — ~/.gitconfig)", globalEmail))
		} else {
			ui.Field("  user.email", "(not set anywhere)")
		}
		fmt.Println()

		// --- Profile Match ---
		fmt.Println("=== Profile ===")
		fmt.Println()

		store, err := profile.LoadStore()
		if err != nil {
			ui.Warn("Could not load profiles: %v", err)
			return nil
		}

		if email != "" {
			matched := store.MatchByEmail(email)
			if matched != nil {
				ui.Success("Matched: %s", matched.Name)
			} else {
				ui.Warn("No profile matches current email %q", email)
				// Suggest all available
				profiles := store.List()
				if len(profiles) > 0 {
					names := make([]string, len(profiles))
					for i, p := range profiles {
						names[i] = fmt.Sprintf("%s <%s>", p.Name, p.UserEmail)
					}
					ui.Info("Available profiles:")
					for _, n := range names {
						fmt.Fprintf(cmd.ErrOrStderr(), "    %s\n", n)
					}
				}
			}
		} else {
			ui.Warn("No email configured — cannot match profiles.")
		}
		fmt.Println()

		// --- Remote-based suggestion ---
		if remote != "" {
			profiles := store.List()
			rules := loadRemoteRules()
			if len(rules) > 0 {
				result := profile.MatchByRemote(remote, profiles, rules)
				if result.Confidence != "none" && result.Profile != nil {
					fmt.Println("=== Suggestion ===")
					fmt.Println()
					ui.Info("Based on remote URL, recommended profile: %s", result.Profile.Name)
					ui.Field("Reason", result.Reason)
					fmt.Println()
				}
			}
		}

		// --- Policy ---
		policies := policy.LoadPolicies()
		if len(policies) > 0 && email != "" {
			fmt.Println("=== Policy ===")
			fmt.Println()

			// Find which profile name we're currently using
			currentProfileName := ""
			if matched := store.MatchByEmail(email); matched != nil {
				currentProfileName = matched.Name
			}

			if currentProfileName != "" && remote != "" {
				violation := policy.Evaluate(currentProfileName, remote)
				if violation != nil {
					ui.Error("Policy violation: %s", violation.Message)
					ui.Field("Policy", violation.PolicyName)
					ui.Field("Mode", violation.Mode)
					ui.Field("Allowed", fmt.Sprintf("%v", violation.AllowedProfiles))
				} else {
					ui.Success("Policy check passed for profile %q", currentProfileName)
				}
			} else if currentProfileName == "" {
				ui.Warn("Cannot evaluate policy — current email doesn't match any profile.")
			}
			fmt.Println()
		}

		return nil
	},
}

// loadRemoteRules builds RemoteRules from loaded policies.
func loadRemoteRules() []profile.RemoteRule {
	policies := policy.LoadPolicies()
	var rules []profile.RemoteRule
	for _, p := range policies {
		for _, pattern := range p.RemotePatterns {
			if pattern == "*" {
				continue // skip catch-all repo policies
			}
			for _, allowed := range p.AllowedProfiles {
				rules = append(rules, profile.RemoteRule{
					Pattern:     pattern,
					ProfileName: allowed,
				})
			}
		}
	}
	return rules
}

func valueOrNotSet(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

func init() {
	rootCmd.AddCommand(showCmd)
}
