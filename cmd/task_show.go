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

var taskShowAliasCmd = &cobra.Command{
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

		body := output.FormatKeyValues([]output.KeyValue{
			{Key: "ID", Value: fmt.Sprintf("%d", task.ID)},
			{Key: "Title", Value: task.Title},
			{Key: "Description", Value: task.Description},
			{Key: "Done", Value: strconv.FormatBool(task.Done)},
			{Key: "Due", Value: output.FormatDateText(task.DueDate, now)},
			{Key: "Project", Value: fmt.Sprintf("%d", task.ProjectID)},
			{Key: "Priority", Value: fmt.Sprintf("%d", task.Priority)},
			{Key: "Favorite", Value: strconv.FormatBool(task.IsFavorite)},
			{Key: "Labels", Value: joinTaskLabelTitles(task.Labels)},
			{Key: "Created", Value: output.FormatDateText(task.Created, now)},
			{Key: "Updated", Value: output.FormatDateText(task.Updated, now)},
		})

		if body == "" {
			continue
		}

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), body); err != nil {
			return err
		}
	}

	return nil
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
