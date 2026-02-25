package cmd

import (
	"context"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/spf13/cobra"
)

var taskDoneCmd = &cobra.Command{
	Use:   "done <id...>",
	Short: "Toggle task done state",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskDone,
}

var taskDoneAliasCmd = &cobra.Command{
	Use:   "done <id...>",
	Short: "Toggle task done state",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskDone,
}

var taskCheckAliasCmd = &cobra.Command{
	Use:   "check <id...>",
	Short: "Toggle task done state",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskDone,
}

func runTaskDone(cmd *cobra.Command, args []string) error {
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

	updated := make([]model.Task, 0, len(ids))
	for _, id := range ids {
		task, err := client.GetTask(context.Background(), id)
		if err != nil {
			return err
		}

		nextDone := !task.Done
		payload := model.TaskUpdatePayload{Done: &nextDone}

		newTask, err := client.UpdateTask(context.Background(), id, payload)
		if err != nil {
			return err
		}

		updated = append(updated, newTask)
	}

	return writeTaskOutput(cmd, updated)
}
