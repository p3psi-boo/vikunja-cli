package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/api"
	"github.com/p3psi-boo/vikunja-cli/config"
	"github.com/spf13/cobra"
)

var (
	loginUsername string
	loginPassword string
	loginTOTP     string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and store a token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client, err := api.NewClient(cfg)
		if err != nil {
			return err
		}

		username := strings.TrimSpace(loginUsername)
		password := strings.TrimSpace(loginPassword)
		totp := strings.TrimSpace(loginTOTP)

		if !flagQuiet && isTTY(os.Stdin) {
			reader := bufio.NewReader(os.Stdin)
			if username == "" {
				username, err = prompt(reader, cmd.ErrOrStderr(), "Username")
				if err != nil {
					return err
				}
			}
			if password == "" {
				password, err = prompt(reader, cmd.ErrOrStderr(), "Password")
				if err != nil {
					return err
				}
			}
			if totp == "" {
				totp, err = prompt(reader, cmd.ErrOrStderr(), "TOTP (optional)")
				if err != nil {
					return err
				}
			}
		}

		if username == "" || password == "" {
			return fmt.Errorf("username and password are required (provide flags, or run interactively in a TTY without --quiet)")
		}

		if _, err := client.Login(context.Background(), username, password, totp); err != nil {
			return err
		}

		if !flagQuiet {
			fmt.Fprintln(cmd.OutOrStdout(), "Login successful")
		}

		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginUsername, "username", "", "Username")
	loginCmd.Flags().StringVar(&loginPassword, "password", "", "Password")
	loginCmd.Flags().StringVar(&loginTOTP, "totp", "", "One-time passcode")

	rootCmd.AddCommand(loginCmd)
}

func isTTY(file *os.File) bool {
	if file == nil {
		return false
	}

	stat, err := file.Stat()
	if err != nil {
		return false
	}

	return (stat.Mode() & os.ModeCharDevice) != 0
}

func prompt(reader *bufio.Reader, out io.Writer, label string) (string, error) {
	if _, err := fmt.Fprintf(out, "%s: ", label); err != nil {
		return "", err
	}

	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(value), nil
}
