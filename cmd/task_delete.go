package cmd

import (
	"context"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <id...>",
	Short: "Delete one or more tasks",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskDelete,
}

var taskDeleteAliasCmd = &cobra.Command{
	Use:   "rm <id...>",
	Short: "Delete one or more tasks",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskDelete,
}

func runTaskDelete(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ids, err := parseTaskIDs(args)
	if err != nil {
		return err
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	for _, id := range ids {
		if err := client.DeleteTask(context.Background(), id); err != nil {
			return err
		}

		if !flagJSON {
			if err := output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "Deleted task #%d", id); err != nil {
				return err
			}
		}
	}

	if flagJSON {
		return output.WriteJSONSingle(cmd.OutOrStdout(), map[string]any{"deleted": true, "ids": ids})
	}

	return nil
}
