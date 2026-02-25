package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/p3psi-boo/vikunja-cli/cmd"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)

	if err := cmd.Execute(); err != nil {
		exitCode := cmd.ExitCode(err)
		if cmd.JSONMode() {
			payload := map[string]any{"error": err.Error(), "code": exitCode}
			encoded, marshalErr := json.Marshal(payload)
			if marshalErr == nil {
				fmt.Fprintln(os.Stderr, string(encoded))
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(exitCode)
	}
}
