package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), formatVersion())
		return err
	},
}

func formatVersion() string {
	return fmt.Sprintf("vja %s (commit %s, built %s)", version, commit, date)
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
