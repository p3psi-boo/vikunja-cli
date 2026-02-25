package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrTokenNotFound = errors.New("token file not found")

type tokenFile struct {
	Token string `json:"token"`
}

type loginRequest struct {
	LongToken bool   `json:"long_token"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	TOTP      string `json:"totp_passcode,omitempty"`
}

func tokenPathCandidates() []string {
	paths := make([]string, 0, 3)

	if cfgDir, ok := os.LookupEnv("VJA_CONFIG_DIR"); ok && strings.TrimSpace(cfgDir) != "" {
		paths = append(paths, filepath.Join(cfgDir, "token.json"))
	}

	if xdgStateHome, ok := os.LookupEnv("XDG_STATE_HOME"); ok && strings.TrimSpace(xdgStateHome) != "" {
		paths = append(paths, filepath.Join(xdgStateHome, "vja", "token.json"))
	}

	if home, ok := os.LookupEnv("HOME"); ok && strings.TrimSpace(home) != "" {
		paths = append(paths, filepath.Join(home, ".local", "state", "vja", "token.json"))
	}

	return dedupeNonEmpty(paths)
}

func dedupeNonEmpty(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))

	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, exists := seen[p]; exists {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	return out
}

func preferredTokenPath() (string, error) {
	paths := tokenPathCandidates()
	if len(paths) == 0 {
		return "", fmt.Errorf("cannot resolve token path from VJA_CONFIG_DIR, XDG_STATE_HOME or HOME")
	}

	return paths[0], nil
}

func LoadToken() (string, error) {
	for _, path := range tokenPathCandidates() {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("read token file %q: %w", path, err)
		}

		var payload tokenFile
		if err := json.Unmarshal(data, &payload); err != nil {
			return "", fmt.Errorf("parse token file %q: %w", path, err)
		}

		token := strings.TrimSpace(payload.Token)
		if token == "" {
			return "", fmt.Errorf("token file %q is missing token", path)
		}

		return token, nil
	}

	return "", ErrTokenNotFound
}

func SaveToken(token string) error {
	path, err := preferredTokenPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create token directory %q: %w", filepath.Dir(path), err)
	}

	data, err := json.Marshal(tokenFile{Token: strings.TrimSpace(token)})
	if err != nil {
		return fmt.Errorf("encode token file: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write token file %q: %w", path, err)
	}

	return nil
}

func DeleteTokenFile() error {
	var firstErr error
	for _, path := range tokenPathCandidates() {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			if firstErr == nil {
				firstErr = fmt.Errorf("remove token file %q: %w", path, err)
			}
		}
	}

	return firstErr
}

func (c *Client) Login(ctx context.Context, username, password, totp string) (string, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	totp = strings.TrimSpace(totp)

	if username == "" || password == "" {
		return "", fmt.Errorf("username and password are required")
	}

	payload := loginRequest{
		LongToken: true,
		Username:  username,
		Password:  password,
		TOTP:      totp,
	}

	var response tokenFile
	if err := c.doJSON(ctx, "POST", "/login", payload, &response, requestOptions{withAuth: false, retryOn401: false}); err != nil {
		return "", err
	}

	token := strings.TrimSpace(response.Token)
	if token == "" {
		return "", fmt.Errorf("login response did not include token")
	}

	c.setToken(token)
	if c.staticToken == "" {
		if err := SaveToken(token); err != nil {
			return "", err
		}
	}

	return token, nil
}

func (c *Client) RefreshToken(ctx context.Context) error {
	if c.staticToken != "" {
		return fmt.Errorf("static API token is configured; refresh is unavailable")
	}

	if c.token == "" {
		token, err := LoadToken()
		if err != nil {
			return fmt.Errorf("refresh failed: %w", err)
		}
		c.token = token
	}

	var response tokenFile
	if err := c.doJSON(ctx, "POST", "/user/token/refresh", nil, &response, requestOptions{withAuth: true, retryOn401: false}); err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}

	token := strings.TrimSpace(response.Token)
	if token == "" {
		return fmt.Errorf("refresh failed: response did not include token")
	}

	c.setToken(token)
	if err := SaveToken(token); err != nil {
		return err
	}

	return nil
}
