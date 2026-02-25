package cmd

import "github.com/spf13/cobra"

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
}

func init() {
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.AddCommand(taskDoneCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskDeferCmd)
	taskCmd.AddCommand(taskCloneCmd)
	taskCmd.AddCommand(taskOpenCmd)

	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(taskListAliasCmd)
	rootCmd.AddCommand(taskShowAliasCmd)
	rootCmd.AddCommand(taskAddAliasCmd)
	rootCmd.AddCommand(taskEditAliasCmd)
	rootCmd.AddCommand(taskDoneAliasCmd)
	rootCmd.AddCommand(taskCheckAliasCmd)
	rootCmd.AddCommand(taskDeleteAliasCmd)
	rootCmd.AddCommand(taskDeferAliasCmd)
	rootCmd.AddCommand(taskCloneAliasCmd)
	rootCmd.AddCommand(taskOpenAliasCmd)
}
