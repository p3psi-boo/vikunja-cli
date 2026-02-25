package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/spf13/cobra"
)

var (
	taskAddProject  string
	taskAddPriority int
	taskAddDue      string
	taskAddReminder string
	taskAddNote     string
	taskAddLabels   []string
	taskAddFavorite bool
)

var taskAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Create a task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskAdd,
}

var taskAddAliasCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Create a task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskAdd,
}

func init() {
	flags := taskAddCmd.Flags()
	flags.StringVarP(&taskAddProject, "project", "p", "", "Target project ID or title")
	flags.IntVar(&taskAddPriority, "priority", 0, "Task priority")
	flags.IntVar(&taskAddPriority, "prio", 0, "Task priority")
	flags.StringVarP(&taskAddDue, "due", "d", "", "Due date")
	flags.StringVarP(&taskAddReminder, "reminder", "r", "", "Reminder date or 'due'")
	flags.Lookup("reminder").NoOptDefVal = "due"
	flags.StringVarP(&taskAddNote, "note", "n", "", "Task note")
	flags.StringArrayVarP(&taskAddLabels, "label", "l", nil, "Label ID or title (repeatable)")
	flags.BoolVarP(&taskAddFavorite, "favorite", "f", false, "Mark as favorite")

	taskAddAliasCmd.Flags().AddFlagSet(flags)

	registerProjectFlagCompletion(taskAddCmd, "project")
	registerProjectFlagCompletion(taskAddAliasCmd, "project")
	registerLabelFlagCompletion(taskAddCmd, "label")
	registerLabelFlagCompletion(taskAddAliasCmd, "label")
}

func runTaskAdd(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	projectRef := taskAddProject
	if !cmd.Flags().Changed("project") {
		projectRef = cfg.Defaults.Project.String()
	}

	if strings.TrimSpace(projectRef) == "" {
		return fmt.Errorf("--project is required or set defaults.project")
	}

	projectID, err := parseProjectID(ctx, client, projectRef)
	if err != nil {
		return err
	}

	labelIDs, err := parseLabelIDs(ctx, client, taskAddLabels)
	if err != nil {
		return err
	}

	title := strings.TrimSpace(args[0])
	if title == "" {
		return fmt.Errorf("task title is required")
	}

	payload := model.TaskCreatePayload{
		Title:       title,
		Description: taskAddNote,
		ProjectID:   &projectID,
		LabelIDs:    labelIDs,
	}

	if cmd.Flags().Changed("priority") || cmd.Flags().Changed("prio") {
		priority := taskAddPriority
		payload.Priority = &priority
	}

	if cmd.Flags().Changed("favorite") {
		favorite := taskAddFavorite
		payload.IsFavorite = &favorite
	}

	if cmd.Flags().Changed("due") {
		due, err := parseTaskDate(taskAddDue)
		if err != nil {
			return err
		}
		payload.DueDate = &due
	}

	if cmd.Flags().Changed("reminder") {
		dueValue := model.NullableTime{}
		if payload.DueDate != nil {
			dueValue = *payload.DueDate
		}

		reminders, err := reminderFromFlag(taskAddReminder, dueValue)
		if err != nil {
			return err
		}
		payload.Reminders = reminders
	}

	task, err := client.CreateTask(ctx, payload)
	if err != nil {
		return err
	}

	return writeTaskOutput(cmd, []model.Task{task})
}
