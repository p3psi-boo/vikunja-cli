package cmd

import (
	"fmt"
	"os"

	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var (
	flagJSON    bool
	flagQuiet   bool
	flagVerbose bool
	flagVersion bool
	flagColor   string
	flagDryRun  bool

	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "vja",
	Short: "A stateless CLI for Vikunja",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Apply the user's color preference for every subcommand. The target
		// writer is stdout; auto-mode falls back to plain text when stdout is
		// piped or when NO_COLOR is set.
		mode, ok := output.ParseColorMode(flagColor)
		if !ok {
			return fmt.Errorf("invalid value %q for --color (expected auto, always or never)", flagColor)
		}
		output.SetColorMode(mode)
		return nil
	},
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
	flags.StringVar(&flagColor, "color", "auto", "Color output: auto, always, or never (also honors NO_COLOR)")
	flags.BoolVar(&flagDryRun, "dry-run", false, "Preview the action without making any changes")

	// NO_COLOR is honored before any command-specific PreRun applies the flag,
	// so even commands that skip PersistentPreRun get the right default.
	output.InitColorFromEnv()
	output.SetColorTarget(os.Stdout)
}

func Execute() error {
	// Wire top-level shortcut flags now that every per-command init() has run.
	registerTaskShortcuts()
	return rootCmd.Execute()
}

func SetVersion(v string) {
	version = v
}

// SetBuildInfo records build-time metadata surfaced by `vja version`.
func SetBuildInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

func JSONMode() bool {
	return flagJSON
}

// DryRunMode reports whether --dry-run was requested.
func DryRunMode() bool {
	return flagDryRun
}
