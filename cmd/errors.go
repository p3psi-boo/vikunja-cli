package cmd

import (
	"errors"
	"strings"

	"github.com/p3psi-boo/vikunja-cli/api"
)

const (
	ExitCodeOK       = 0
	ExitCodeError    = 1
	ExitCodeUsage    = 2
	ExitCodeAuth     = 3
	ExitCodeNotFound = 4
)

func ExitCode(err error) int {
	if err == nil {
		return ExitCodeOK
	}

	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 401:
			return ExitCodeAuth
		case 404:
			return ExitCodeNotFound
		}
	}

	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(msg, "authentication failed") || strings.Contains(msg, "run `vja login`") {
		return ExitCodeAuth
	}
	if strings.Contains(msg, "username and password are required") {
		return ExitCodeAuth
	}

	if isResourceNotFoundMessage(msg) {
		return ExitCodeNotFound
	}

	if isUsageMessage(msg) {
		return ExitCodeUsage
	}

	return ExitCodeError
}

func isUsageMessage(msg string) bool {
	usageHints := []string{
		"unknown flag",
		"accepts ",
		"requires at least",
		"requires at most",
		"requires exactly",
		"must be",
		"invalid ",
		" is required",
		"unsupported",
	}

	for _, hint := range usageHints {
		if strings.Contains(msg, hint) {
			return true
		}
	}

	return false
}

func isResourceNotFoundMessage(msg string) bool {
	resourceHints := []string{
		"task ",
		"project ",
		"label ",
	}

	if !strings.Contains(msg, " not found") {
		return false
	}

	for _, hint := range resourceHints {
		if strings.Contains(msg, hint) {
			return true
		}
	}

	return false
}
