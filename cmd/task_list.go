package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var (
	taskListAll        bool
	taskListProject    string
	taskListLabels     []string
	taskListPriority   string
	taskListDue        string
	taskListFavorite   bool
	taskListFilters    []string
	taskListSort       string
	taskListLimit      int
	taskListAbsolute   bool
	taskListNoSummary  bool
)

var taskListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List tasks",
	Long: `List tasks, defaulting to incomplete ones.

Examples:
  # All open tasks
  vja ls

  # Tasks in a project, with a label, limited to 20
  vja ls -p 1 -l backend -n 20

  # Include completed tasks
  vja ls -a

  # Raw Vikunja filter expression (repeatable)
  vja ls --filter 'due_date < now'

  # Sort: priority descending then due ascending
  vja ls -s '-priority,due_date'

See the Vikunja API docs for the full filter and sort syntax.`,
	Args: cobra.NoArgs,
	RunE: runTaskList,
}

func init() {
	flags := taskListCmd.Flags()
	flags.BoolVarP(&taskListAll, "all", "a", false, "Include completed tasks")
	flags.StringVarP(&taskListProject, "project", "p", "", "Filter by project")
	flags.StringArrayVarP(&taskListLabels, "label", "l", nil, "Filter by label (repeatable)")
	flags.StringVar(&taskListPriority, "priority", "", "Filter by priority")
	flags.StringVarP(&taskListDue, "due", "d", "", "Filter by due date")
	flags.BoolVarP(&taskListFavorite, "favorite", "f", false, "Show favorite tasks only")
	flags.StringArrayVar(&taskListFilters, "filter", nil, "Raw Vikunja filter expression, e.g. 'due_date < now' (repeatable)")
	flags.StringVarP(&taskListSort, "sort", "s", "", "Sort expression, e.g. '-priority,due_date'")
	flags.IntVarP(&taskListLimit, "limit", "n", 0, "Limit number of tasks")
	flags.BoolVar(&taskListAbsolute, "absolute", false, "Show absolute due dates alongside relative ones")
	flags.BoolVar(&taskListNoSummary, "no-summary", false, "Hide the task count summary line")

	registerProjectFlagCompletion(taskListCmd, "project")
	registerLabelFlagCompletion(taskListCmd, "label")
}

func runTaskList(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	filter := api.TaskFilter{
		All:      taskListAll,
		Project:  taskListProject,
		Labels:   taskListLabels,
		Priority: taskListPriority,
		Due:      taskListDue,
		Favorite: taskListFavorite,
		Filters:  taskListFilters,
		Sort:     taskListSort,
		Limit:    taskListLimit,
	}

	ctx := context.Background()
	tasks, err := client.GetTasks(ctx, filter)
	if err != nil {
		return err
	}

	if taskListLimit > 0 && len(tasks) > taskListLimit {
		tasks = tasks[:taskListLimit]
	}

	if flagJSON {
		return output.WriteJSONList(cmd.OutOrStdout(), tasks)
	}

	if len(tasks) == 0 {
		return output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "%s\n", output.EmptyMessage("tasks", listHasFilters()))
	}

	// Build the project id->title lookup so the Project column is populated.
	// Failures are non-fatal: we just show an empty project name.
	projectByID := map[int64]string{}
	if projects, perr := client.GetProjects(ctx); perr == nil {
		for _, p := range projects {
			projectByID[p.ID] = p.Title
		}
	} else if flagVerbose {
		_ = output.PrintInfo(cmd.ErrOrStderr(), false, "warning: could not load projects: %v\n", perr)
	}

	now := time.Now()
	if !taskListNoSummary {
		if err := printTaskSummary(cmd, tasks, now); err != nil {
			return err
		}
	}

	table := output.FormatTaskTableWithOptions(tasks, projectByID, now, output.TaskTableOptions{
		AbsoluteDate: taskListAbsolute,
	})

	_, err = fmt.Fprintln(cmd.OutOrStdout(), table)
	return err
}

func listHasFilters() bool {
	return taskListProject != "" || len(taskListLabels) > 0 || taskListPriority != "" ||
		taskListDue != "" || taskListFavorite || len(taskListFilters) > 0 || taskListSort != ""
}

// printTaskSummary prints a one-line tally such as "12 tasks (3 overdue, 2 done)".
func printTaskSummary(cmd *cobra.Command, tasks []model.Task, now time.Time) error {
	total := len(tasks)
	done, overdue := 0, 0
	for _, t := range tasks {
		if t.Done {
			done++
			continue
		}
		if output.ClassifyDue(t.DueDate, now) == output.DueStateOverdue {
			overdue++
		}
	}

	parts := []string{fmt.Sprintf("%d %s", total, pluralWord(total, "task", "tasks"))}
	if overdue > 0 {
		parts = append(parts, fmt.Sprintf("%d overdue", overdue))
	}
	if done > 0 {
		parts = append(parts, fmt.Sprintf("%d done", done))
	}

	line := output.Muted(strings.Join(parts, ", "))
	_, err := fmt.Fprintln(cmd.OutOrStdout(), line)
	return err
}

func pluralWord(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
