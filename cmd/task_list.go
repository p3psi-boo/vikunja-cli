package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var (
	taskListAll      bool
	taskListProject  string
	taskListLabels   []string
	taskListPriority string
	taskListDue      string
	taskListFavorite bool
	taskListFilters  []string
	taskListSort     string
	taskListLimit    int
)

var taskListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List tasks",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE:    runTaskList,
}

var taskListAliasCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks",
	Args:  cobra.NoArgs,
	RunE:  runTaskList,
}

func init() {
	flags := taskListCmd.Flags()
	flags.BoolVarP(&taskListAll, "all", "a", false, "Include completed tasks")
	flags.StringVarP(&taskListProject, "project", "p", "", "Filter by project")
	flags.StringArrayVarP(&taskListLabels, "label", "l", nil, "Filter by label (repeatable)")
	flags.StringVar(&taskListPriority, "priority", "", "Filter by priority")
	flags.StringVarP(&taskListDue, "due", "d", "", "Filter by due date")
	flags.BoolVarP(&taskListFavorite, "favorite", "f", false, "Show favorite tasks only")
	flags.StringArrayVar(&taskListFilters, "filter", nil, "Add raw filter (repeatable)")
	flags.StringVarP(&taskListSort, "sort", "s", "", "Sort expression")
	flags.IntVarP(&taskListLimit, "limit", "n", 0, "Limit number of tasks")

	taskListAliasCmd.Flags().AddFlagSet(flags)

	registerProjectFlagCompletion(taskListCmd, "project")
	registerProjectFlagCompletion(taskListAliasCmd, "project")
	registerLabelFlagCompletion(taskListCmd, "label")
	registerLabelFlagCompletion(taskListAliasCmd, "label")
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

	tasks, err := client.GetTasks(context.Background(), filter)
	if err != nil {
		return err
	}

	if taskListLimit > 0 && len(tasks) > taskListLimit {
		tasks = tasks[:taskListLimit]
	}

	if flagJSON {
		return output.WriteJSONList(cmd.OutOrStdout(), tasks)
	}

	table := output.FormatTaskTable(tasks, nil, time.Now())
	if table == "" {
		return nil
	}

	_, err = fmt.Fprintln(cmd.OutOrStdout(), table)
	return err
}
