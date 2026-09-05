package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Malay-dev/gitx-profile/internal/config"
	"github.com/Malay-dev/gitx-profile/internal/policy"
	"github.com/Malay-dev/gitx-profile/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage identity policies",
	Long: `Policies control which profiles are allowed in which repositories.

Policies can be defined at two levels:
  1. User-level:  ~/.git-profile/policies.yaml
  2. Repo-level:  .gitprofile.yml (in repository root)

Repository-level policies take precedence over user-level policies.`,
}

var policyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active policies",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		policies := policy.LoadPolicies()

		if len(policies) == 0 {
			ui.Info("No policies configured.")
			ui.Info("Run 'git profile policy create' to create one.")
			return nil
		}

		for _, p := range policies {
			fmt.Printf("  %s\n", p.Name)
			ui.Field("  Remote Patterns", strings.Join(p.RemotePatterns, ", "))
			ui.Field("  Allowed Profiles", strings.Join(p.AllowedProfiles, ", "))
			ui.Field("  Mode", p.Mode)
			fmt.Println()
		}

		return nil
	},
}

var policyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new policy (interactive)",
	Long:  `Interactively create a policy and save it to ~/.git-profile/policies.yaml.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		// Policy name
		fmt.Print("Policy name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("policy name is required")
		}

		// Remote patterns
		fmt.Print("Remote patterns (comma-separated, e.g., github.com/AcmeCorp/*): ")
		patternsRaw, _ := reader.ReadString('\n')
		patternsRaw = strings.TrimSpace(patternsRaw)
		if patternsRaw == "" {
			return fmt.Errorf("at least one remote pattern is required")
		}
		patterns := splitAndTrim(patternsRaw)

		// Allowed profiles
		fmt.Print("Allowed profiles (comma-separated, e.g., work,personal): ")
		profilesRaw, _ := reader.ReadString('\n')
		profilesRaw = strings.TrimSpace(profilesRaw)
		if profilesRaw == "" {
			return fmt.Errorf("at least one allowed profile is required")
		}
		profiles := splitAndTrim(profilesRaw)

		// Mode
		fmt.Print("Mode (strict/warning/advisory) [strict]: ")
		mode, _ := reader.ReadString('\n')
		mode = strings.TrimSpace(mode)
		if mode == "" {
			mode = "strict"
		}
		if mode != "strict" && mode != "warning" && mode != "advisory" {
			return fmt.Errorf("invalid mode: %s (must be strict, warning, or advisory)", mode)
		}

		newPolicy := policy.Policy{
			Name:            name,
			RemotePatterns:  patterns,
			AllowedProfiles: profiles,
			Mode:            mode,
		}

		// Load existing policies
		if err := appendUserPolicy(newPolicy); err != nil {
			return fmt.Errorf("failed to save policy: %w", err)
		}

		ui.Success("Policy %q created", name)
		fmt.Println()
		ui.Field("Remote Patterns", strings.Join(patterns, ", "))
		ui.Field("Allowed Profiles", strings.Join(profiles, ", "))
		ui.Field("Mode", mode)
		ui.Info("Saved to %s", config.PoliciesFilePath())

		return nil
	},
}

// appendUserPolicy reads existing policies, appends the new one, and writes back.
func appendUserPolicy(p policy.Policy) error {
	path := config.PoliciesFilePath()

	var pf policy.PolicyFile

	// Try to load existing file
	data, err := os.ReadFile(path)
	if err == nil {
		_ = yaml.Unmarshal(data, &pf)
	}

	// Check for duplicate name
	for _, existing := range pf.Policies {
		if strings.EqualFold(existing.Name, p.Name) {
			return fmt.Errorf("policy %q already exists", p.Name)
		}
	}

	pf.Policies = append(pf.Policies, p)

	// Ensure directory exists
	dir := config.PoliciesDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", dir, err)
	}

	out, err := yaml.Marshal(&pf)
	if err != nil {
		return fmt.Errorf("cannot marshal policies: %w", err)
	}

	return os.WriteFile(path, out, 0o644)
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func init() {
	policyCmd.AddCommand(policyListCmd)
	policyCmd.AddCommand(policyCreateCmd)
	rootCmd.AddCommand(policyCmd)
}
