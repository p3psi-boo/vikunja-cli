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
	"github.com/p3psi-boo/vikunja-cli/model"
	"github.com/spf13/cobra"
)

var (
	loginAPIURL   string
	loginUsername string
	loginPassword string
	loginTOTP     string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and store a token",
	Long: `Authenticate against a Vikunja server and store the resulting token.

When run interactively in a TTY without the relevant flags, login guides you
through providing the API URL (saving it to the config when none is set yet),
username, password, and an optional TOTP passcode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load existing config if present; an interactive flow can bootstrap one
		// when the config file or api_url is missing.
		cfg, cfgErr := config.Load()

		interactive := !flagQuiet && isTTY(os.Stdin)
		reader := bufio.NewReader(os.Stdin)

		apiURL := strings.TrimSpace(loginAPIURL)
		if apiURL == "" && cfg != nil {
			apiURL = strings.TrimSpace(cfg.Server.APIURL)
		}

		// Guide the user to set the API URL when it is not configured.
		if apiURL == "" {
			if !interactive {
				if cfgErr != nil {
					return fmt.Errorf("no api url configured: %w (provide --api-url, or run interactively in a TTY without --quiet)", cfgErr)
				}
				return fmt.Errorf("no api url configured (provide --api-url, or run interactively in a TTY without --quiet)")
			}

			prompted, err := promptWithDefault(reader, cmd.ErrOrStderr(), "API URL", "")
			if err != nil {
				return err
			}
			apiURL = strings.TrimSpace(prompted)
			if apiURL == "" {
				return fmt.Errorf("api url is required")
			}
		}

		// Defer persisting the API URL until after a successful login, so a
		// failed login does not rewrite an existing config with a bad URL.
		if cfg == nil {
			cfg = &config.Config{}
		}
		originalAPIURL := cfg.Server.APIURL
		cfg.Server.APIURL = apiURL

		client, err := api.NewClient(cfg)
		if err != nil {
			cfg.Server.APIURL = originalAPIURL
			return err
		}

		username := strings.TrimSpace(loginUsername)
		password := strings.TrimSpace(loginPassword)
		totp := strings.TrimSpace(loginTOTP)

		if interactive {
			if username == "" {
				username, err = prompt(reader, cmd.ErrOrStderr(), "Username")
				if err != nil {
					return err
				}
			}
			if password == "" {
				password, err = promptHidden(cmd.ErrOrStderr(), "Password")
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
			cfg.Server.APIURL = originalAPIURL
			return err
		}

		// Persist the api url only now that login succeeded.
		if strings.TrimSpace(originalAPIURL) == "" || !strings.EqualFold(strings.TrimRight(originalAPIURL, "/"), strings.TrimRight(apiURL, "/")) {
			if err := config.Save(cfg, cfg.Path); err != nil {
				return err
			}
		}

		if !flagQuiet {
			who := loginIdentity(client)
			if who != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Login successful: "+who)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Login successful")
			}
		}

		return nil
	},
}

// loginIdentity fetches a short description of the logged-in user (e.g.
// "alice (alice@example.com)"). Returns an empty string on any failure so the
// caller can fall back to a plain success message.
func loginIdentity(client *api.Client) string {
	var user model.User
	if err := client.GetJSON(context.Background(), "/user", &user); err != nil {
		return ""
	}
	name := strings.TrimSpace(user.Name)
	switch {
	case user.Username != "" && name != "":
		return user.Username + " (" + name + ")"
	case user.Username != "" && user.Email != "":
		return user.Username + " (" + user.Email + ")"
	case user.Username != "":
		return user.Username
	default:
		return ""
	}
}

func init() {
	loginCmd.Flags().StringVar(&loginAPIURL, "api-url", "", "Vikunja API URL (saved to config when not yet set)")
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
	return promptWithDefault(reader, out, label, "")
}

func promptWithDefault(reader *bufio.Reader, out io.Writer, label, def string) (string, error) {
	if strings.TrimSpace(def) != "" {
		if _, err := fmt.Fprintf(out, "%s [%s]: ", label, def); err != nil {
			return "", err
		}
	} else {
		if _, err := fmt.Fprintf(out, "%s: ", label); err != nil {
			return "", err
		}
	}

	value, err := reader.ReadString('\n')
	if err != nil && value == "" {
		return "", err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		value = strings.TrimSpace(def)
	}
	return value, nil
}

// promptHidden reads a line from stdin for password entry, echoing an asterisk
// per typed character. It reads directly from the terminal file descriptor
// rather than the provided bufio.Reader: that reader may have buffered bytes
// already, and switching the tty to raw mode underneath it leaves those bytes
// stranded.
//
// On platforms where raw mode is unavailable it falls back to a normal echoed
// prompt.
func promptHidden(out io.Writer, label string) (string, error) {
	if value, ok, err := readPassword(int(os.Stdin.Fd()), out, label); ok {
		return value, err
	}

	// Fallback: no raw-mode support. Re-prompt on a fresh reader so we don't
	// trip over any bytes the caller's reader already buffered.
	if _, err := fmt.Fprintf(out, "%s: ", label); err != nil {
		return "", err
	}
	reader := bufio.NewReader(os.Stdin)
	value, readErr := reader.ReadString('\n')
	if readErr != nil && value == "" {
		return "", readErr
	}
	return strings.TrimSpace(value), nil
}
