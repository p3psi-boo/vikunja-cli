package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var (
	taskDeleteYes bool
)

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <id...>",
	Short: "Delete one or more tasks",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskDelete,
}

func init() {
	taskDeleteCmd.Flags().BoolVarP(&taskDeleteYes, "yes", "y", false, "Skip confirmation prompt")
}

func runTaskDelete(cmd *cobra.Command, args []string) error {
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

	// Confirm before destructive action. Only prompt when stdin is a TTY and
	// the user did not opt out with --yes, so scripted usage stays non-interactive.
	if !flagJSON && !flagQuiet && isTTY(os.Stdin) && !taskDeleteYes {
		if confirmed, err := confirmDelete(cmd, client, ids); err != nil {
			return err
		} else if !confirmed {
			return output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "Aborted.\n")
		}
	}

	if flagDryRun {
		for _, id := range ids {
			if err := output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "[dry-run] would delete task #%d\n", id); err != nil {
				return err
			}
		}
		return nil
	}

	for _, id := range ids {
		if err := client.DeleteTask(context.Background(), id); err != nil {
			return err
		}

		if !flagJSON {
			if err := output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "Deleted task #%d\n", id); err != nil {
				return err
			}
		}
	}

	if flagJSON {
		return output.WriteJSONSingle(cmd.OutOrStdout(), map[string]any{"deleted": true, "ids": ids})
	}

	return nil
}

// confirmDelete fetches the task titles so the prompt is informative, then asks
// for an explicit yes. Anything other than y/Y cancels.
func confirmDelete(cmd *cobra.Command, client *api.Client, ids []int64) (bool, error) {
	ctx := context.Background()
	labels := make([]string, 0, len(ids))
	for _, id := range ids {
		task, err := client.GetTask(ctx, id)
		if err != nil {
			// Fall back to a plain id prompt; a bad id will surface as a real
			// error from DeleteTask shortly after.
			labels = append(labels, fmt.Sprintf("#%d", id))
			continue
		}
		labels = append(labels, fmt.Sprintf("#%d %q", id, task.Title))
	}

	var prompt string
	if len(labels) == 1 {
		prompt = fmt.Sprintf("Delete task %s? [y/N] ", labels[0])
	} else {
		prompt = fmt.Sprintf("Delete %d tasks (%s)? [y/N] ", len(labels), strings.Join(labels, ", "))
	}

	if _, err := fmt.Fprint(cmd.ErrOrStderr(), prompt); err != nil {
		return false, err
	}

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
