package cmd

import "github.com/spf13/cobra"

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
}

// registerTaskShortcut registers a top-level command (e.g. `vja ls`) that
// mirrors one of the `vja task <sub>` commands. It binds the same flags by
// applying the subcommand's flag set, so there is exactly one source of truth
// for each command's flags (no drift between the alias and the canonical form).
func registerTaskShortcut(name, short string, sub *cobra.Command) *cobra.Command {
	shortcut := &cobra.Command{
		Use:   name,
		Short: short,
		Args:  sub.Args,
		RunE:  sub.RunE,
	}
	// Flag binding is deferred to registerTaskShortcuts() because Go runs this
	// file's init() before the per-command files' init() (alphabetical order),
	// so the subcommand flags are not populated yet at init time.
	rootCmd.AddCommand(shortcut)
	return shortcut
}

// taskShortcuts maps a top-level name to the canonical task subcommand and its
// description. The flags are wired in registerTaskShortcuts() once all init()
// functions have populated the subcommand flag sets.
var taskShortcuts = []struct {
	name, short string
	sub         *cobra.Command
}{
	{"ls", "List tasks", nil},
	{"show", "Show one or more tasks", nil},
	{"add", "Create a task", nil},
	{"edit", "Edit one or more tasks", nil},
	{"done", "Mark tasks as done", nil},
	{"undone", "Mark tasks as not done", nil},
	{"toggle", "Toggle task done state", nil},
	{"check", "Toggle task done state", nil},
	{"rm", "Delete one or more tasks", nil},
	{"defer", "Defer due date and reminder", nil},
	{"clone", "Clone a task", nil},
	{"open", "Open tasks in the browser", nil},
}

var taskShortcutCmds []*cobra.Command

func init() {
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.AddCommand(taskDoneCmd)
	taskCmd.AddCommand(taskUndoneCmd)
	taskCmd.AddCommand(taskToggleCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskDeferCmd)
	taskCmd.AddCommand(taskCloneCmd)
	taskCmd.AddCommand(taskOpenCmd)

	rootCmd.AddCommand(taskCmd)

	// Register the top-level shortcut commands. Their flags are attached in
	// registerTaskShortcuts() to avoid depending on init() ordering.
	taskShortcuts[0].sub = taskListCmd
	taskShortcuts[1].sub = taskShowCmd
	taskShortcuts[2].sub = taskAddCmd
	taskShortcuts[3].sub = taskEditCmd
	taskShortcuts[4].sub = taskDoneCmd
	taskShortcuts[5].sub = taskUndoneCmd
	taskShortcuts[6].sub = taskToggleCmd
	taskShortcuts[7].sub = taskToggleCmd
	taskShortcuts[8].sub = taskDeleteCmd
	taskShortcuts[9].sub = taskDeferCmd
	taskShortcuts[10].sub = taskCloneCmd
	taskShortcuts[11].sub = taskOpenCmd

	for _, s := range taskShortcuts {
		taskShortcutCmds = append(taskShortcutCmds, registerTaskShortcut(s.name, s.short, s.sub))
	}
}

// registerTaskShortcuts copies each canonical subcommand's flag set onto its
// top-level shortcut. Called from Execute() so every per-command init() has
// already run and the flag sets are fully populated.
func registerTaskShortcuts() {
	for i, s := range taskShortcuts {
		taskShortcutCmds[i].Flags().AddFlagSet(s.sub.Flags())
	}
}

