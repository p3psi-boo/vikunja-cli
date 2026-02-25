package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var projectAddParent string

var projectAddCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Create a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		payload := model.ProjectCreatePayload{Title: strings.TrimSpace(args[0])}
		if payload.Title == "" {
			return fmt.Errorf("project title is required")
		}

		client, err := api.NewClient(cfg)
		if err != nil {
			return err
		}

		parentValue := strings.TrimSpace(projectAddParent)
		if parentValue != "" {
			parentID, err := resolveProjectID(context.Background(), client, parentValue)
			if err != nil {
				return fmt.Errorf("resolve --parent: %w", err)
			}
			payload.ParentProjectID = &parentID
		}

		project, err := client.CreateProject(context.Background(), payload)
		if err != nil {
			return err
		}

		if projectJSONOutput(cfg) {
			return output.WriteJSONSingle(cmd.OutOrStdout(), project)
		}

		return output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "Created project #%d: %s", project.ID, project.Title)
	},
}

func init() {
	projectAddCmd.Flags().StringVarP(&projectAddParent, "parent", "p", "", "Parent project ID or title")
	projectCmd.AddCommand(projectAddCmd)
}
