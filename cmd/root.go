package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-profile",
	Short: "Manage multiple Git identities",
	Long: `git-profile is a Git extension that manages multiple Git identities.

Switch between work, personal, and open-source profiles without
manually editing git config. Part of the gitx ecosystem.

Usage:
  git profile <command> [flags]

Examples:
  git profile add work --name "Malay Kumar" --email "malay@company.com"
  git profile use work
  git profile current
  git profile list`,
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
