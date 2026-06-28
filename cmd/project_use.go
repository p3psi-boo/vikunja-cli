package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)
var projectUseUnset bool

var projectUseCmd = &cobra.Command{
	Use:   "use [<project>]",
	Short: "Set the default project for this working directory",
	Long: `Pin a default project for the current working directory by writing
defaults.project to a project-local .vja.yaml.

The argument may be a project ID or a title. Titles are validated against the
server (and must match a single project) so the stored reference is stable.
A bare integer is stored as an ID without contacting the server.

Run without an argument and with --unset to remove the default project again.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load is required even for --unset so that an absent project override
		// does not fail silently on a misconfigured environment.
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var ref config.ProjectRef
		if projectUseUnset {
			if len(args) > 0 {
				return fmt.Errorf("--unset takes no project argument")
			}
			// ref stays as the zero value; SaveProjectDefault drops the section.
		} else {
			if len(args) == 0 {
				return fmt.Errorf("a project ID or title is required (or pass --unset)")
			}
			raw := strings.TrimSpace(args[0])
			if raw == "" {
				return fmt.Errorf("project argument is required")
			}

			// A pure positive integer is stored as an ID without an API call;
			// anything else is resolved (and validated) by title.
			if id, perr := strconv.ParseInt(raw, 10, 64); perr == nil && id > 0 {
				ref = config.ProjectRef{ID: &id}
			} else {
				client, cerr := api.NewClient(cfg)
				if cerr != nil {
					return cerr
				}
				if _, rerr := resolveProjectID(context.Background(), client, raw); rerr != nil {
					return rerr
				}
				ref = config.ProjectRef{Name: raw}
			}
		}

		label := ref.String()
		if projectUseUnset {
			label = "(unset)"
		}

		if flagDryRun {
			return output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "[dry-run] would write defaults.project=%s to .vja.yaml\n", label)
		}

		path, err := config.SaveProjectDefault(ref)
		if err != nil {
			return err
		}

		if projectUseUnset {
			return output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "Cleared default project (%s)\n", path)
		}
		return output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "Set default project %s (%s)\n", label, path)
	},
}

func init() {
	projectUseCmd.Flags().BoolVar(&projectUseUnset, "unset", false, "Remove the default project for this working directory")
	projectUseCmd.RegisterFlagCompletionFunc("unset", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	projectUseCmd.ValidArgsFunction = completeProjectArg
	projectCmd.AddCommand(projectUseCmd)
}

// completeProjectArg suggests projects for positional arguments that take a
// project reference (id or title), e.g. `vja project use <project>`. Unlike
// completeProjectFlag (which matches on id only), this matches the partial
// input against both id and title so users can complete by name.
func completeProjectArg(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projects, err := projectsForCompletion()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	needle := strings.ToLower(toComplete)
	suggestions := make([]string, 0, len(projects))
	for _, project := range projects {
		if project.ID <= 0 {
			continue
		}

		id := strconv.FormatInt(project.ID, 10)
		title := strings.TrimSpace(project.Title)

		if toComplete != "" {
			if !strings.HasPrefix(id, toComplete) && !strings.HasPrefix(strings.ToLower(title), needle) {
				continue
			}
		}

		if title == "" {
			suggestions = append(suggestions, id)
			continue
		}
		suggestions = append(suggestions, id+"\t"+title)
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}
