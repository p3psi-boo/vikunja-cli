package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrMissingAPIURL  = errors.New("server.api_url is required")
)

type Config struct {
	Path     string         `toml:"-"`
	Server   ServerConfig   `toml:"server"`
	Defaults DefaultsConfig `toml:"defaults"`
	Output   OutputConfig   `toml:"output"`
}

type ServerConfig struct {
	APIURL      string `toml:"api_url"`
	FrontendURL string `toml:"frontend_url"`
	APIToken    string `toml:"api_token"`
}

type DefaultsConfig struct {
	Project ProjectRef `toml:"project"`
}

type ProjectRef struct {
	ID   *int64
	Name string
}

func (p *ProjectRef) UnmarshalTOML(value any) error {
	switch v := value.(type) {
	case nil:
		p.ID = nil
		p.Name = ""
		return nil
	case string:
		p.ID = nil
		p.Name = strings.TrimSpace(v)
		return nil
	case int64:
		id := v
		p.ID = &id
		p.Name = ""
		return nil
	case int:
		id := int64(v)
		p.ID = &id
		p.Name = ""
		return nil
	case float64:
		if v != float64(int64(v)) {
			return fmt.Errorf("defaults.project must be string or integer")
		}
		id := int64(v)
		p.ID = &id
		p.Name = ""
		return nil
	default:
		return fmt.Errorf("defaults.project must be string or integer")
	}
}

type OutputConfig struct {
	Format string `toml:"format"`
}

func Load() (*Config, error) {
	path, searched, err := findConfigPath()
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			return nil, fmt.Errorf("%w (searched: %s)", ErrConfigNotFound, strings.Join(searched, ", "))
		}
		return nil, err
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	applyEnvOverrides(&cfg)

	if strings.TrimSpace(cfg.Server.APIURL) == "" {
		return nil, fmt.Errorf("%w in %s", ErrMissingAPIURL, path)
	}

	cfg.Path = path
	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if apiURL, ok := os.LookupEnv("VJA_API_URL"); ok {
		cfg.Server.APIURL = apiURL
	}

	if apiToken, ok := os.LookupEnv("VJA_API_TOKEN"); ok {
		cfg.Server.APIToken = apiToken
	}
}

func findConfigPath() (string, []string, error) {
	paths := configCandidates()
	for _, p := range paths {
		info, err := os.Stat(p)
		if err == nil {
			if info.IsDir() {
				continue
			}
			return p, paths, nil
		}

		if !os.IsNotExist(err) {
			return "", paths, fmt.Errorf("check config %q: %w", p, err)
		}
	}

	return "", paths, ErrConfigNotFound
}

func configCandidates() []string {
	paths := make([]string, 0, 3)

	if cfgDir, ok := os.LookupEnv("VJA_CONFIG_DIR"); ok && strings.TrimSpace(cfgDir) != "" {
		paths = append(paths, filepath.Join(cfgDir, "config.toml"))
	}

	if xdgConfigHome, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok && strings.TrimSpace(xdgConfigHome) != "" {
		paths = append(paths, filepath.Join(xdgConfigHome, "vja", "config.toml"))
	}

	if home, ok := os.LookupEnv("HOME"); ok && strings.TrimSpace(home) != "" {
		paths = append(paths, filepath.Join(home, ".config", "vja", "config.toml"))
	}

	return uniqueNonEmpty(paths)
}

func uniqueNonEmpty(paths []string) []string {
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

func (p ProjectRef) IsSet() bool {
	return p.ID != nil || p.Name != ""
}

func (p ProjectRef) String() string {
	if p.Name != "" {
		return p.Name
	}

	if p.ID == nil {
		return ""
	}

	return strconv.FormatInt(*p.ID, 10)
}
