package cmd

import (
	"context"
	"fmt"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var projectListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client, err := api.NewClient(cfg)
		if err != nil {
			return err
		}

		projects, err := client.GetProjects(context.Background())
		if err != nil {
			return err
		}

		if projectJSONOutput(cfg) {
			return output.WriteJSONList(cmd.OutOrStdout(), projects)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), output.FormatProjectTable(projects))
		return err
	},
}

func init() {
	projectCmd.AddCommand(projectListCmd)
}
