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
	"github.com/spf13/pflag"
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

	registerProjectFlagCompletion(taskEditCmd, "project")
	registerLabelFlagCompletion(taskEditCmd, "label")
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

		if flagDryRun {
			describe := describeEditChanges(cmd.Flags())
			if err := output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "[dry-run] would update task #%d: %s\n", id, describe); err != nil {
				return err
			}
			updated = append(updated, task)
			continue
		}

		newTask, err := client.UpdateTask(ctx, id, payload)
		if err != nil {
			return err
		}

		updated = append(updated, newTask)
	}

	return writeTaskOutput(cmd, updated)
}

// describeEditChanges produces a short human summary of the flags the user
// passed to `vja edit`, for use in --dry-run previews.
func describeEditChanges(flags *pflag.FlagSet) string {
	var parts []string
	add := func(name, value string) {
		parts = append(parts, name+"="+value)
	}
	if flags.Changed("title") {
		add("title", fmt.Sprintf("%q", taskEditTitle))
	}
	if flags.Changed("note") {
		add("note", fmt.Sprintf("%q", taskEditNote))
	}
	if flags.Changed("note-append") {
		add("note-append", fmt.Sprintf("%q", taskEditNoteAppend))
	}
	if flags.Changed("project") {
		add("project", taskEditProject)
	}
	if flags.Changed("priority") || flags.Changed("prio") {
		add("priority", strconv.Itoa(taskEditPriority))
	}
	if flags.Changed("due") {
		add("due", taskEditDue)
	}
	if flags.Changed("reminder") {
		add("reminder", taskEditReminder)
	}
	if flags.Changed("label") {
		add("labels", fmt.Sprintf("%v", taskEditLabels))
	}
	if flags.Changed("favorite") {
		add("favorite", strconv.FormatBool(taskEditFavorite))
	}
	if flags.Changed("done") {
		add("done", strconv.FormatBool(taskEditDone))
	}
	if len(parts) == 0 {
		return "no changes"
	}
	return strings.Join(parts, ", ")
}
