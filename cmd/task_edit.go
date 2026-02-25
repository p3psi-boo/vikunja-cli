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
	taskEditProject    string
	taskEditPriority   int
	taskEditDue        string
	taskEditReminder   string
	taskEditNote       string
	taskEditNoteAppend string
	taskEditLabels     []string
	taskEditFavorite   bool
	taskEditTitle      string
	taskEditDone       bool
)

var taskEditCmd = &cobra.Command{
	Use:   "edit <id...>",
	Short: "Edit one or more tasks",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskEdit,
}

var taskEditAliasCmd = &cobra.Command{
	Use:   "edit <id...>",
	Short: "Edit one or more tasks",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskEdit,
}

func init() {
	flags := taskEditCmd.Flags()
	flags.StringVarP(&taskEditProject, "project", "p", "", "Target project ID or title")
	flags.IntVar(&taskEditPriority, "priority", 0, "Task priority")
	flags.IntVar(&taskEditPriority, "prio", 0, "Task priority")
	flags.StringVarP(&taskEditDue, "due", "d", "", "Due date")
	flags.StringVarP(&taskEditReminder, "reminder", "r", "", "Reminder date or 'due'")
	flags.Lookup("reminder").NoOptDefVal = "due"
	flags.StringVarP(&taskEditNote, "note", "n", "", "Task note")
	flags.StringArrayVarP(&taskEditLabels, "label", "l", nil, "Label ID or title (repeatable, toggles on edit)")
	flags.BoolVarP(&taskEditFavorite, "favorite", "f", false, "Mark as favorite")
	flags.StringVarP(&taskEditTitle, "title", "t", "", "Task title")
	flags.BoolVar(&taskEditDone, "done", false, "Set done status")
	flags.StringVar(&taskEditNoteAppend, "note-append", "", "Append text to the note")

	taskEditAliasCmd.Flags().AddFlagSet(flags)

	registerProjectFlagCompletion(taskEditCmd, "project")
	registerProjectFlagCompletion(taskEditAliasCmd, "project")
	registerLabelFlagCompletion(taskEditCmd, "label")
	registerLabelFlagCompletion(taskEditAliasCmd, "label")
}

func runTaskEdit(cmd *cobra.Command, args []string) error {
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

	ctx := context.Background()

	var projectID int64
	if cmd.Flags().Changed("project") {
		projectID, err = parseProjectID(ctx, client, taskEditProject)
		if err != nil {
			return err
		}
	}

	var labelIDs []int64
	if cmd.Flags().Changed("label") {
		labelIDs, err = parseLabelIDs(ctx, client, taskEditLabels)
		if err != nil {
			return err
		}
	}

	updated := make([]model.Task, 0, len(ids))
	for _, id := range ids {
		task, err := client.GetTask(ctx, id)
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("title") {
			title := strings.TrimSpace(taskEditTitle)
			if title == "" {
				return fmt.Errorf("--title cannot be empty")
			}
			task.Title = title
		}

		if cmd.Flags().Changed("note") {
			task.Description = taskEditNote
		}

		if cmd.Flags().Changed("note-append") {
			if task.Description == "" {
				task.Description = taskEditNoteAppend
			} else {
				task.Description += "\n" + taskEditNoteAppend
			}
		}

		if cmd.Flags().Changed("project") {
			task.ProjectID = projectID
		}

		if cmd.Flags().Changed("priority") || cmd.Flags().Changed("prio") {
			task.Priority = taskEditPriority
		}

		if cmd.Flags().Changed("favorite") {
			task.IsFavorite = taskEditFavorite
		}

		if cmd.Flags().Changed("done") {
			task.Done = taskEditDone
		}

		if cmd.Flags().Changed("due") {
			due, err := parseTaskDate(taskEditDue)
			if err != nil {
				return err
			}
			task.DueDate = due
		}

		if cmd.Flags().Changed("reminder") {
			reminders, err := reminderFromFlag(taskEditReminder, task.DueDate)
			if err != nil {
				return err
			}
			task.Reminders = reminders
		}

		payload := model.TaskUpdatePayload{}
		payload.Title = &task.Title
		payload.Description = &task.Description
		payload.Done = &task.Done
		payload.DueDate = &task.DueDate
		payload.ProjectID = &task.ProjectID
		payload.Priority = &task.Priority
		payload.IsFavorite = &task.IsFavorite
		payload.DoneAt = &task.DoneAt

		reminders := append([]model.TaskReminder(nil), task.Reminders...)
		payload.Reminders = &reminders

		if cmd.Flags().Changed("label") {
			toggledLabelIDs := toggleTaskLabelIDs(taskLabelIDs(task), labelIDs)
			payload.LabelIDs = &toggledLabelIDs
		} else {
			taskLabels := taskLabelIDs(task)
			payload.LabelIDs = &taskLabels
		}

		newTask, err := client.UpdateTask(ctx, id, payload)
		if err != nil {
			return err
		}

		updated = append(updated, newTask)
	}

	return writeTaskOutput(cmd, updated)
}
