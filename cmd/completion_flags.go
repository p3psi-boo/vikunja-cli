package cmd

import (
	"context"
	"strconv"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/spf13/cobra"
)

func registerProjectFlagCompletion(cmd *cobra.Command, flagName string) {
	if cmd.Flags().Lookup(flagName) == nil {
		return
	}

	_ = cmd.RegisterFlagCompletionFunc(flagName, completeProjectFlag)
}

func registerLabelFlagCompletion(cmd *cobra.Command, flagName string) {
	if cmd.Flags().Lookup(flagName) == nil {
		return
	}

	_ = cmd.RegisterFlagCompletionFunc(flagName, completeLabelFlag)
}

func completeProjectFlag(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projects, err := projectsForCompletion()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	suggestions := make([]string, 0, len(projects))
	for _, project := range projects {
		if project.ID <= 0 {
			continue
		}

		id := strconv.FormatInt(project.ID, 10)
		if toComplete != "" && !strings.HasPrefix(id, toComplete) {
			continue
		}

		title := strings.TrimSpace(project.Title)
		if title == "" {
			suggestions = append(suggestions, id)
			continue
		}

		suggestions = append(suggestions, id+"\t"+title)
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

func completeLabelFlag(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	labels, err := labelsForCompletion()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	suggestions := make([]string, 0, len(labels))
	for _, label := range labels {
		if label.ID <= 0 {
			continue
		}

		id := strconv.FormatInt(label.ID, 10)
		if toComplete != "" && !strings.HasPrefix(id, toComplete) {
			continue
		}

		title := strings.TrimSpace(label.Title)
		if title == "" {
			suggestions = append(suggestions, id)
			continue
		}

		suggestions = append(suggestions, id+"\t"+title)
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

func projectsForCompletion() ([]apiProject, error) {
	client, err := completionClient()
	if err != nil {
		return nil, err
	}

	projects, err := client.GetProjects(context.Background())
	if err != nil {
		return nil, err
	}

	out := make([]apiProject, 0, len(projects))
	for _, project := range projects {
		out = append(out, apiProject{ID: project.ID, Title: project.Title})
	}

	return out, nil
}

func labelsForCompletion() ([]apiLabel, error) {
	client, err := completionClient()
	if err != nil {
		return nil, err
	}

	labels, err := client.GetLabels(context.Background())
	if err != nil {
		return nil, err
	}

	out := make([]apiLabel, 0, len(labels))
	for _, label := range labels {
		out = append(out, apiLabel{ID: label.ID, Title: label.Title})
	}

	return out, nil
}

func completionClient() (*api.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	return api.NewClient(cfg)
}

type apiProject struct {
	ID    int64
	Title string
}

type apiLabel struct {
	ID    int64
	Title string
}
