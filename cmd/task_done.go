package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

// doneAction describes how a command wants to set a task's done state.
type doneAction int

const (
	doneOnly   doneAction = iota // done:   only complete (error if already done)
	undoneOnly                    // undone: only reopen (error if already not done)
	toggleState                   // toggle: flip current state
)

var taskDoneCmd = &cobra.Command{
	Use:   "done <id...>",
	Short: "Mark tasks as done",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskDoneAction(cmd, args, doneOnly)
	},
}

var taskUndoneCmd = &cobra.Command{
	Use:   "undone <id...>",
	Short: "Mark tasks as not done",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskDoneAction(cmd, args, undoneOnly)
	},
}

var taskToggleCmd = &cobra.Command{
	Use:   "toggle <id...>",
	Short: "Toggle task done state",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskDoneAction(cmd, args, toggleState)
	},
}

func runTaskDoneAction(cmd *cobra.Command, args []string, action doneAction) error {
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
	ctx := context.Background()
	for _, id := range ids {
		task, err := client.GetTask(ctx, id)
		if err != nil {
			return err
		}

		nextDone, err := resolveNextDone(task, id, action)
		if err != nil {
			return err
		}

		if nextDone == task.Done {
			// No state change required (e.g. toggle was a no-op after a prior
			// check, or action matched current state). Still echo the task.
			updated = append(updated, task)
			continue
		}

		if flagDryRun {
			updated = append(updated, task)
			continue
		}

		payload := model.TaskUpdatePayload{Done: &nextDone}
		newTask, err := client.UpdateTask(ctx, id, payload)
		if err != nil {
			return err
		}
		updated = append(updated, newTask)
	}

	return writeTaskDoneOutput(cmd, updated, action)
}

func resolveNextDone(task model.Task, id int64, action doneAction) (bool, error) {
	switch action {
	case doneOnly:
		if task.Done {
			return false, fmt.Errorf("task #%d is already done; use `vja undone %d` to reopen", id, id)
		}
		return true, nil
	case undoneOnly:
		if !task.Done {
			return false, fmt.Errorf("task #%d is not done; use `vja done %d` to complete it", id, id)
		}
		return false, nil
	default: // toggleState
		return !task.Done, nil
	}
}

func writeTaskDoneOutput(cmd *cobra.Command, tasks []model.Task, action doneAction) error {
	if flagJSON {
		return output.WriteJSONList(cmd.OutOrStdout(), tasks)
	}

	for _, task := range tasks {
		verb := describeDoneOutcome(task, action)
		if err := output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "%s\n", verb); err != nil {
			return err
		}
	}

	return writeTaskOutput(cmd, tasks)
}

func describeDoneOutcome(task model.Task, action doneAction) string {
	id := strconv.FormatInt(task.ID, 10)
	prefix := ""
	if flagDryRun {
		prefix = "[dry-run] would "
	}
	switch action {
	case doneOnly:
		if task.Done {
			return prefix + "Marked #" + id + " as done"
		}
		return prefix + "Mark #" + id + " as done"
	case undoneOnly:
		if !task.Done {
			return prefix + "Reopened #" + id
		}
		return prefix + "Reopen #" + id
	default:
		state := "done"
		if !task.Done {
			state = "not done"
		}
		if flagDryRun {
			return prefix + "toggle #" + id + " -> " + state
		}
		return "Toggled #" + id + " (now " + state + ")"
	}
}
