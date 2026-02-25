package cmd

import (
	"context"
	"time"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/spf13/cobra"
)

var taskDeferCmd = &cobra.Command{
	Use:   "defer <id...> <duration>",
	Short: "Defer due date and reminder",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runTaskDefer,
}

var taskDeferAliasCmd = &cobra.Command{
	Use:   "defer <id...> <duration>",
	Short: "Defer due date and reminder",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runTaskDefer,
}

func runTaskDefer(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	delta, err := parseDeferDuration(args[len(args)-1])
	if err != nil {
		return err
	}

	ids, err := parseTaskIDs(args[:len(args)-1])
	if err != nil {
		return err
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	now := time.Now()

	updated := make([]model.Task, 0, len(ids))
	for _, id := range ids {
		task, err := client.GetTask(context.Background(), id)
		if err != nil {
			return err
		}

		if task.DueDate.Valid {
			task.DueDate = shiftFromNowIfPast(task.DueDate, now, delta)
		}

		for i := range task.Reminders {
			if !task.Reminders[i].Reminder.Valid {
				continue
			}
			if task.Reminders[i].RelativePeriod != 0 || task.Reminders[i].RelativeTo != "" {
				continue
			}

			task.Reminders[i].Reminder = shiftFromNowIfPast(task.Reminders[i].Reminder, now, delta)
			break
		}

		payload := model.TaskUpdatePayload{
			DueDate:   &task.DueDate,
			Reminders: &task.Reminders,
		}

		newTask, err := client.UpdateTask(context.Background(), id, payload)
		if err != nil {
			return err
		}

		updated = append(updated, newTask)
	}

	return writeTaskOutput(cmd, updated)
}
