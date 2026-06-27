package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var taskDeferSetDue bool

var taskDeferCmd = &cobra.Command{
	Use:   "defer <id...> <duration>",
	Short: "Defer due date and reminder",
	Long: `Push the due date (and reminder) of one or more tasks forward by a duration.

The duration is applied relative to now for any date that is already in the
past, otherwise it is added to the existing date. Durations look like: 1d, 2h30m, 1w.

Tasks without a due date are skipped unless --set-due is given, in which case
the duration is treated as "due <duration> from now".`,
	Args: cobra.MinimumNArgs(2),
	RunE: runTaskDefer,
}

func init() {
	taskDeferCmd.Flags().BoolVar(&taskDeferSetDue, "set-due", false, "Set a due date (now + duration) when none exists")
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
	ctx := context.Background()
	for _, id := range ids {
		task, err := client.GetTask(ctx, id)
		if err != nil {
			return err
		}

		hasDue := task.DueDate.Valid
		hasShiftableReminder := hasShiftableReminder(task)

		if !hasDue && !hasShiftableReminder {
			if !taskDeferSetDue {
				return fmt.Errorf("task #%d has no due date or reminders; nothing to defer. Set one with: vja edit %d -d <date>, or re-run with --set-due", id, id)
			}
			// Bootstrap a due date from now + delta.
			task.DueDate = model.NewNullableTime(now.Add(delta))
		} else if hasDue {
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

		if flagDryRun {
			if err := output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "[dry-run] would defer task #%d by %s (due: %s)\n", id, delta, output.FormatDateRich(task.DueDate, now)); err != nil {
				return err
			}
			updated = append(updated, task)
			continue
		}

		payload := model.TaskUpdatePayload{
			DueDate:   &task.DueDate,
			Reminders: &task.Reminders,
		}

		newTask, err := client.UpdateTask(ctx, id, payload)
		if err != nil {
			return err
		}

		updated = append(updated, newTask)
	}

	return writeTaskOutput(cmd, updated)
}

func hasShiftableReminder(task model.Task) bool {
	for _, r := range task.Reminders {
		if !r.Reminder.Valid {
			continue
		}
		if r.RelativePeriod != 0 || r.RelativeTo != "" {
			continue
		}
		return true
	}
	return false
}
