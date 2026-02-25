package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Show current user",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client, err := api.NewClient(cfg)
		if err != nil {
			return err
		}

		var user model.User
		if err := client.GetJSON(context.Background(), "/user", &user); err != nil {
			return err
		}

		if flagJSON {
			return output.WriteJSONSingle(cmd.OutOrStdout(), user)
		}

		now := time.Now()

		text := output.FormatKeyValues([]output.KeyValue{
			{Key: "ID", Value: strconv.FormatInt(user.ID, 10)},
			{Key: "Username", Value: user.Username},
			{Key: "Name", Value: user.Name},
			{Key: "Email", Value: user.Email},
			{Key: "Created", Value: output.FormatDateText(user.Created, now)},
			{Key: "Updated", Value: output.FormatDateText(user.Updated, now)},
		})
		_, err = fmt.Fprintln(cmd.OutOrStdout(), text)
		return err
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
}
