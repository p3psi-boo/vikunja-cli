package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// projectConfigName is the per-project override file, discovered by walking up
// from the current working directory. It layers on top of the XDG config and
// is overridden by environment variables.
const projectConfigName = ".vja.yaml"

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrMissingAPIURL  = errors.New("server.api_url is required")
	ErrNoConfigHome   = errors.New("cannot resolve config path from VJA_CONFIG_DIR, XDG_CONFIG_HOME or HOME")
)

type Config struct {
	Path              string         `toml:"-" yaml:"-"`
	ProjectConfigPath string         `toml:"-" yaml:"-"`
	Server            ServerConfig   `toml:"server"   yaml:"server"`
	Defaults          DefaultsConfig `toml:"defaults" yaml:"defaults"`
	Output            OutputConfig   `toml:"output"   yaml:"output"`
}

type ServerConfig struct {
	APIURL      string `toml:"api_url"      yaml:"api_url"`
	FrontendURL string `toml:"frontend_url" yaml:"frontend_url"`
	APIToken    string `toml:"api_token"    yaml:"api_token"`
}

type DefaultsConfig struct {
	Project ProjectRef `toml:"project" yaml:"project"`
}

type ProjectRef struct {
	ID   *int64
	Name string
}

func (p *ProjectRef) UnmarshalTOML(value any) error {
	return p.assignFromValue(value)
}

// UnmarshalYAML lets .vja.yaml set defaults.project as a string (project name)
// or an integer (project id), mirroring the TOML semantics.
func (p *ProjectRef) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		switch value.Tag {
		case "!!int":
			n, err := strconv.ParseInt(value.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("defaults.project: invalid integer %q: %w", value.Value, err)
			}
			p.ID = &n
			p.Name = ""
			return nil
		case "!!str", "!!null":
			p.ID = nil
			p.Name = strings.TrimSpace(value.Value)
			return nil
		}
	}
	// Fallback: decode into any and reuse the same value-type logic as TOML.
	var raw any
	if err := value.Decode(&raw); err != nil {
		return fmt.Errorf("defaults.project: %w", err)
	}
	return p.assignFromValue(raw)
}

// assignFromValue is the shared parser for a project reference value. Both the
// TOML and YAML unmarshalers route through it so the accepted shapes stay in
// lockstep: nil, string (name), or integer (id).
func (p *ProjectRef) assignFromValue(value any) error {
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
	Format string `toml:"format" yaml:"format"`
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

	// Layer any project-local .vja.yaml on top of the XDG config before env
	// overrides are applied, so precedence is: flag > env > project > XDG.
	if projPath, err := findProjectConfig(); err != nil {
		return nil, err
	} else if projPath != "" {
		if err := mergeProjectConfig(&cfg, projPath); err != nil {
			return nil, err
		}
		cfg.ProjectConfigPath = projPath
	}

	applyEnvOverrides(&cfg)

	if strings.TrimSpace(cfg.Server.APIURL) == "" {
		return nil, fmt.Errorf("%w in %s", ErrMissingAPIURL, path)
	}

	cfg.Path = path
	return &cfg, nil
}

// PreferredConfigPath returns the path where a new config file should be written
// (the first config candidate). It mirrors the lookup order used when loading.
func PreferredConfigPath() (string, error) {
	paths := configCandidates()
	if len(paths) == 0 {
		return "", ErrNoConfigHome
	}
	return paths[0], nil
}

// Save writes the config to the given path (or the preferred path when empty),
// creating the parent directory with 0o700 permissions.
func Save(cfg *Config, path string) error {
	if strings.TrimSpace(path) == "" {
		p, err := PreferredConfigPath()
		if err != nil {
			return err
		}
		path = p
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory %q: %w", filepath.Dir(path), err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open config %q: %w", path, err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("encode config %q: %w", path, err)
	}

	cfg.Path = path
	return nil
}

// findProjectConfig walks up from the current working directory looking for a
// project-local override file (.vja.yaml). It returns the first match, or ""
// when none is found.
func findProjectConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, projectConfigName)
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("check project config %q: %w", candidate, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

// mergeProjectConfig decodes the project-local YAML file and overlays its
// non-empty fields onto base. Empty/omitted fields are left untouched, and
// base.Path is never rewritten (it always points at the XDG config so that
// `vja login` writes back to the right place).
func mergeProjectConfig(base *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read project config %q: %w", path, err)
	}

	var overlay Config
	if err := yaml.Unmarshal(data, &overlay); err != nil {
		return fmt.Errorf("parse project config %q: %w", path, err)
	}

	if strings.TrimSpace(overlay.Server.APIURL) != "" {
		base.Server.APIURL = overlay.Server.APIURL
	}
	if strings.TrimSpace(overlay.Server.FrontendURL) != "" {
		base.Server.FrontendURL = overlay.Server.FrontendURL
	}
	if strings.TrimSpace(overlay.Server.APIToken) != "" {
		base.Server.APIToken = overlay.Server.APIToken
	}
	if overlay.Defaults.Project.IsSet() {
		base.Defaults.Project = overlay.Defaults.Project
	}
	if strings.TrimSpace(overlay.Output.Format) != "" {
		base.Output.Format = overlay.Output.Format
	}
	return nil
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
