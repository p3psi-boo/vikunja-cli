package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/dateparse"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

func parseTaskIDs(args []string) ([]int64, error) {
	ids := make([]int64, 0, len(args))
	for _, arg := range args {
		id, err := strconv.ParseInt(strings.TrimSpace(arg), 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid task id %q", arg)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func parseProjectID(ctx context.Context, client *api.Client, value string) (int64, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return 0, fmt.Errorf("--project is required")
	}

	return resolveProjectID(ctx, client, raw)
}

func resolveProjectID(ctx context.Context, client *api.Client, raw string) (int64, error) {
	if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
		return id, nil
	}

	projects, err := client.GetProjects(ctx)
	if err != nil {
		return 0, err
	}

	var matchID int64
	matches := 0
	for _, project := range projects {
		if project.ID <= 0 {
			continue
		}

		if strings.TrimSpace(project.Title) != raw {
			continue
		}

		matchID = project.ID
		matches++
	}

	if matches == 1 {
		return matchID, nil
	}

	if matches > 1 {
		return 0, fmt.Errorf("multiple projects found with title %q; use project ID", raw)
	}

	return 0, fmt.Errorf("project %q not found", raw)
}

func parseLabelIDs(ctx context.Context, client *api.Client, values []string) ([]int64, error) {
	if len(values) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(values))
	var labels []model.Label
	labelsLoaded := false

	for _, value := range values {
		raw := strings.TrimSpace(value)
		if raw == "" {
			continue
		}

		if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
			ids = append(ids, id)
			continue
		}

		if !labelsLoaded {
			var err error
			labels, err = client.GetLabels(ctx)
			if err != nil {
				return nil, err
			}
			labelsLoaded = true
		}

		id, err := resolveLabelID(raw, labels)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func resolveLabelID(raw string, labels []model.Label) (int64, error) {
	var matchID int64
	matches := 0

	for _, label := range labels {
		if label.ID <= 0 {
			continue
		}

		if strings.TrimSpace(label.Title) != raw {
			continue
		}

		matchID = label.ID
		matches++
	}

	if matches == 1 {
		return matchID, nil
	}

	if matches > 1 {
		return 0, fmt.Errorf("multiple labels found with title %q; use label ID", raw)
	}

	return 0, fmt.Errorf("label %q not found", raw)
}

func parseTaskDate(raw string) (model.NullableTime, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return model.NullableTime{}, nil
	}

	parsed, err := dateparse.ParseDateExpr(time.Now(), value)
	if err != nil {
		return model.NullableTime{}, err
	}

	return model.NewNullableTime(parsed), nil
}

func parseDeferDuration(raw string) (time.Duration, error) {
	return dateparse.ParseDurationExpr(raw)
}

func reminderFromFlag(raw string, due model.NullableTime) ([]model.TaskReminder, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return []model.TaskReminder{}, nil
	}

	if strings.EqualFold(value, "due") {
		if !due.Valid {
			return nil, fmt.Errorf("--reminder=due requires a due date")
		}

		return []model.TaskReminder{{Reminder: due}}, nil
	}

	reminder, err := parseTaskDate(value)
	if err != nil {
		return nil, err
	}

	if !reminder.Valid {
		return []model.TaskReminder{}, nil
	}

	return []model.TaskReminder{{Reminder: reminder}}, nil
}

func taskLabelIDs(task model.Task) []int64 {
	if len(task.Labels) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(task.Labels))
	for _, label := range task.Labels {
		if label.ID > 0 {
			ids = append(ids, label.ID)
		}
	}

	return ids
}

func toggleTaskLabelIDs(current []int64, toggles []int64) []int64 {
	if len(toggles) == 0 {
		return append([]int64(nil), current...)
	}

	ids := make([]int64, 0, len(current))
	present := make(map[int64]struct{}, len(current))
	for _, id := range current {
		if id <= 0 {
			continue
		}
		if _, exists := present[id]; exists {
			continue
		}
		present[id] = struct{}{}
		ids = append(ids, id)
	}

	for _, id := range toggles {
		if id <= 0 {
			continue
		}

		if _, exists := present[id]; exists {
			delete(present, id)
			for i := range ids {
				if ids[i] != id {
					continue
				}
				ids = append(ids[:i], ids[i+1:]...)
				break
			}
			continue
		}

		present[id] = struct{}{}
		ids = append(ids, id)
	}

	return ids
}

func writeTaskOutput(cmd *cobra.Command, tasks []model.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	if flagJSON {
		if len(tasks) == 1 {
			return output.WriteJSONSingle(cmd.OutOrStdout(), tasks[0])
		}
		return output.WriteJSONList(cmd.OutOrStdout(), tasks)
	}

	now := time.Now()
	for i, task := range tasks {
		if i > 0 {
			if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
				return err
			}
		}

		text := output.FormatKeyValues([]output.KeyValue{
			{Key: "ID", Value: strconv.FormatInt(task.ID, 10)},
			{Key: "Title", Value: task.Title},
			{Key: "Done", Value: strconv.FormatBool(task.Done)},
			{Key: "Due", Value: output.FormatDateText(task.DueDate, now)},
			{Key: "Project", Value: strconv.FormatInt(task.ProjectID, 10)},
			{Key: "Priority", Value: strconv.Itoa(task.Priority)},
			{Key: "Favorite", Value: strconv.FormatBool(task.IsFavorite)},
		})

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), text); err != nil {
			return err
		}
	}

	return nil
}

func shiftFromNowIfPast(current model.NullableTime, now time.Time, delta time.Duration) model.NullableTime {
	if !current.Valid {
		return current
	}

	base := current.Time
	if base.Before(now) {
		base = now
	}

	return model.NewNullableTime(base.Add(delta))
}
