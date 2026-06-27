package cmd

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var taskShowCmd = &cobra.Command{
	Use:   "show <id...>",
	Short: "Show one or more tasks",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskShow,
}

func runTaskShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	tasks := make([]model.Task, 0, len(args))
	for _, arg := range args {
		id, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task id %q", arg)
		}

		task, err := client.GetTask(context.Background(), id)
		if err != nil {
			return err
		}

		tasks = append(tasks, task)
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

		header := output.Bold(fmt.Sprintf("── Task #%d ──", task.ID))
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), header); err != nil {
			return err
		}

		body := output.FormatKeyValuesOmitEmpty([]output.KeyValue{
			{Key: "Title", Value: task.Title},
			{Key: "Description", Value: task.Description},
			{Key: "Done", Value: formatBoolMark(task.Done)},
			{Key: "Due", Value: output.FormatDateRich(task.DueDate, now)},
			{Key: "Project", Value: strconv.FormatInt(task.ProjectID, 10)},
			{Key: "Priority", Value: strconv.Itoa(task.Priority)},
			{Key: "Favorite", Value: formatFavoriteMark(task.IsFavorite)},
			{Key: "Labels", Value: joinTaskLabelTitles(task.Labels)},
			{Key: "Created", Value: output.FormatDateText(task.Created, now)},
			{Key: "Updated", Value: output.FormatDateText(task.Updated, now)},
		})

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), body); err != nil {
			return err
		}
	}

	return nil
}

// formatBoolMark renders an affirmative boolean as a check, and an empty value
// (so FormatKeyValuesOmitEmpty drops it) for the "false" case, keeping detail
// views free of a wall of "false" rows.
func formatBoolMark(value bool) string {
	if value {
		return output.CheckMark()
	}
	return ""
}

func formatFavoriteMark(value bool) string {
	if value {
		return output.FavoriteMark()
	}
	return ""
}

func joinTaskLabelTitles(labels []model.Label) string {
	if len(labels) == 0 {
		return ""
	}

	titles := make([]string, 0, len(labels))
	for _, label := range labels {
		titles = append(titles, label.Title)
	}

	sort.Strings(titles)
	return strings.Join(titles, ",")
}
