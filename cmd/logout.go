package cmd

import (
	"fmt"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.Server.APIToken != "" {
			if !flagQuiet {
				fmt.Fprintln(cmd.OutOrStdout(), "Static API token is configured; no token file to remove")
			}
			return nil
		}

		if err := api.DeleteTokenFile(); err != nil {
			return err
		}

		if !flagQuiet {
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
