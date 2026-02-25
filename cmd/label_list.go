package cmd

import (
	"context"
	"fmt"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var labelListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List labels",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client, err := api.NewClient(cfg)
		if err != nil {
			return err
		}

		labels, err := client.GetLabels(context.Background())
		if err != nil {
			return err
		}

		if flagJSON {
			return output.WriteJSONList(cmd.OutOrStdout(), labels)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), output.FormatLabelTable(labels))
		return err
	},
}

func init() {
	labelCmd.AddCommand(labelListCmd)
}
