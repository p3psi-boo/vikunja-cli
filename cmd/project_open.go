package cmd

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/spf13/cobra"
)

var projectOpenCmd = &cobra.Command{
	Use:   "open <id>",
	Short: "Open a project in the browser",
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

		frontendURL := strings.TrimSpace(cfg.Server.FrontendURL)
		if frontendURL == "" {
			return fmt.Errorf("server.frontend_url is required for `vja project open`")
		}

		projectURL, err := buildProjectURL(frontendURL, id)
		if err != nil {
			return err
		}

		if err := openURL(projectURL); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectOpenCmd)
}

func buildProjectURL(frontendURL string, id int64) (string, error) {
	base, err := url.Parse(frontendURL)
	if err != nil {
		return "", fmt.Errorf("invalid server.frontend_url: %w", err)
	}

	base.Path = strings.TrimRight(base.Path, "/") + "/projects/" + strconv.FormatInt(id, 10)
	return base.String(), nil
}

func openURL(rawURL string) error {
	var command string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		command = "open"
		args = []string{rawURL}
	case "windows":
		command = "cmd"
		args = []string{"/c", "start", "", rawURL}
	default:
		command = "xdg-open"
		args = []string{rawURL}
	}

	if err := exec.Command(command, args...).Run(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}

	return nil
}
