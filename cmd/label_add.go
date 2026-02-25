package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var labelAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Create a label",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client, err := api.NewClient(cfg)
		if err != nil {
			return err
		}

		label, err := client.CreateLabel(context.Background(), args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			return output.WriteJSONSingle(cmd.OutOrStdout(), label)
		}

		text := output.FormatKeyValues([]output.KeyValue{
			{Key: "ID", Value: strconv.FormatInt(label.ID, 10)},
			{Key: "Title", Value: label.Title},
			{Key: "Color", Value: label.HexColor},
		})
		_, err = fmt.Fprintln(cmd.OutOrStdout(), text)
		return err
	},
}

func init() {
	labelCmd.AddCommand(labelAddCmd)
}
