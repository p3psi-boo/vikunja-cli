package cmd

import (
	"strings"

	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

func projectJSONOutput(cfg *config.Config) bool {
	if flagJSON {
		return true
	}

	return strings.EqualFold(strings.TrimSpace(cfg.Output.Format), "json")
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
