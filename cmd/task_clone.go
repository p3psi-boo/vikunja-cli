package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var taskCloneCmd = &cobra.Command{
	Use:   "clone <id> [title]",
	Short: "Clone a task",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runTaskClone,
}

func runTaskClone(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(strings.TrimSpace(args[0]), 10, 64)
	if err != nil || id <= 0 {
		return fmt.Errorf("invalid task id %q", args[0])
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	task, err := client.GetTask(context.Background(), id)
	if err != nil {
		return err
	}

	title := task.Title
	if len(args) == 2 {
		title = strings.TrimSpace(args[1])
		if title == "" {
			return fmt.Errorf("clone title cannot be empty")
		}
	}

	projectID := task.ProjectID
	payload := model.TaskCreatePayload{
		Title:       title,
		Description: task.Description,
		Done:        task.Done,
		ProjectID:   &projectID,
		Reminders:   append([]model.TaskReminder(nil), task.Reminders...),
		LabelIDs:    taskLabelIDs(task),
	}

	if task.DueDate.Valid {
		due := task.DueDate
		payload.DueDate = &due
	}

	priority := task.Priority
	payload.Priority = &priority

	favorite := task.IsFavorite
	payload.IsFavorite = &favorite

	if flagDryRun {
		return output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "[dry-run] would clone task #%d as %q in project #%d\n", id, title, projectID)
	}

	created, err := client.CreateTask(context.Background(), payload)
	if err != nil {
		return err
	}

	return writeTaskOutput(cmd, []model.Task{created})
}
