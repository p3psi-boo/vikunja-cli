package cmd

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/spf13/cobra"
)

var taskOpenCmd = &cobra.Command{
	Use:   "open [id...]",
	Short: "Open tasks in the browser",
	Args:  cobra.ArbitraryArgs,
	RunE:  runTaskOpen,
}

var taskOpenAliasCmd = &cobra.Command{
	Use:   "open [id...]",
	Short: "Open tasks in the browser",
	Args:  cobra.ArbitraryArgs,
	RunE:  runTaskOpen,
}

func runTaskOpen(_ *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	frontendURL := strings.TrimSpace(cfg.Server.FrontendURL)
	if frontendURL == "" {
		return fmt.Errorf("server.frontend_url is required for `vja task open`")
	}

	if len(args) == 0 {
		if err := openURL(frontendURL); err != nil {
			return err
		}
		return nil
	}

	for _, arg := range args {
		id, err := strconv.ParseInt(arg, 10, 64)
		if err != nil || id <= 0 {
			return fmt.Errorf("task id must be a positive integer")
		}

		taskURL, err := buildTaskURL(frontendURL, id)
		if err != nil {
			return err
		}

		if err := openURL(taskURL); err != nil {
			return err
		}
	}

	return nil
}

func buildTaskURL(frontendURL string, id int64) (string, error) {
	base, err := url.Parse(frontendURL)
	if err != nil {
		return "", fmt.Errorf("invalid server.frontend_url: %w", err)
	}

	base.Path = strings.TrimRight(base.Path, "/") + "/tasks/" + strconv.FormatInt(id, 10)
	return base.String(), nil
}
