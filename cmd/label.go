package cmd

import "github.com/spf13/cobra"

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage labels",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(labelCmd)
}
