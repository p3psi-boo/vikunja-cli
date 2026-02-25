package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var projectShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show project details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil || id <= 0 {
			return fmt.Errorf("project id must be a positive integer")
		}

		client, err := api.NewClient(cfg)
		if err != nil {
			return err
		}

		project, err := client.GetProject(context.Background(), id)
		if err != nil {
			return err
		}

		if projectJSONOutput(cfg) {
			return output.WriteJSONSingle(cmd.OutOrStdout(), project)
		}

		parent := ""
		if project.ParentProjectID != nil {
			parent = strconv.FormatInt(*project.ParentProjectID, 10)
		}

		lines := []output.KeyValue{
			{Key: "ID", Value: strconv.FormatInt(project.ID, 10)},
			{Key: "Title", Value: project.Title},
			{Key: "Description", Value: project.Description},
			{Key: "Parent", Value: parent},
			{Key: "Color", Value: project.HexColor},
			{Key: "Favorite", Value: strconv.FormatBool(project.IsFavorite)},
			{Key: "Created", Value: output.FormatDateJSON(project.Created)},
			{Key: "Updated", Value: output.FormatDateJSON(project.Updated)},
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), output.FormatKeyValues(lines))
		return err
	},
}

func init() {
	projectCmd.AddCommand(projectShowCmd)
}
