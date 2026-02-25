package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagJSON    bool
	flagQuiet   bool
	flagVerbose bool
	flagVersion bool

	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "vja",
	Short: "A stateless CLI for Vikunja",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagVersion {
			fmt.Fprintln(cmd.OutOrStdout(), version)
			return nil
		}

		return cmd.Help()
	},
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.BoolVarP(&flagJSON, "json", "j", false, "Output as JSON")
	flags.BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress informational messages")
	flags.BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose logs")
	flags.BoolVar(&flagVersion, "version", false, "Print version")
}

func Execute() error {
	return rootCmd.Execute()
}

func SetVersion(v string) {
	version = v
}

func JSONMode() bool {
	return flagJSON
}
