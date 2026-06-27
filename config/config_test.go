package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// writeXDGConfig stages a fake XDG config in a temp dir and points the loader
// at it via VJA_CONFIG_DIR. It returns the config file path and a restore
// function. Pair with `defer restore()` in each test.
func writeXDGConfig(t *testing.T, body string) (string, func()) {
	t.Helper()
	xdg := t.TempDir()
	path := filepath.Join(xdg, "config.toml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write xdg config: %v", err)
	}
	t.Setenv("VJA_CONFIG_DIR", xdg)
	// Ignore HOME/XDG_CONFIG_HOME so only VJA_CONFIG_DIR resolves.
	t.Setenv("XDG_CONFIG_HOME", "")
	return path, func() {}
}

// chdir changes into dir for the duration of the test.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %q: %v", dir, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(orig)
	})
}

const xdgBase = `[server]
api_url = "https://xdg.example.com/api/v1"
api_token = "xdg-token"
frontend_url = "https://xdg.example.com"

[defaults]
project = "xdg-project"

[output]
format = "text"
`

func TestLoad_XDGOnly(t *testing.T) {
	path, _ := writeXDGConfig(t, xdgBase)
	chdir(t, t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.APIURL != "https://xdg.example.com/api/v1" {
		t.Fatalf("api_url = %q", cfg.Server.APIURL)
	}
	if cfg.Defaults.Project.Name != "xdg-project" {
		t.Fatalf("project = %q", cfg.Defaults.Project.Name)
	}
	if cfg.Path != path {
		t.Fatalf("Path = %q, want %q", cfg.Path, path)
	}
	if cfg.ProjectConfigPath != "" {
		t.Fatalf("ProjectConfigPath = %q, want empty", cfg.ProjectConfigPath)
	}
}

func TestLoad_ProjectOverridesDefaultProject(t *testing.T) {
	writeXDGConfig(t, xdgBase)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".vja.yaml"), []byte(
		"defaults:\n  project: work-project\n",
	), 0o600); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	chdir(t, root)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Defaults.Project.Name != "work-project" {
		t.Fatalf("project = %q, want work-project", cfg.Defaults.Project.Name)
	}
	// Untouched XDG fields survive.
	if cfg.Server.APIURL != "https://xdg.example.com/api/v1" {
		t.Fatalf("api_url = %q", cfg.Server.APIURL)
	}
}

func TestLoad_ProjectServerOverride(t *testing.T) {
	writeXDGConfig(t, xdgBase)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".vja.yaml"), []byte(
		"server:\n  api_url: https://corp.example.com/api/v1\n",
	), 0o600); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	chdir(t, root)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.APIURL != "https://corp.example.com/api/v1" {
		t.Fatalf("api_url = %q", cfg.Server.APIURL)
	}
	// Token still comes from XDG.
	if cfg.Server.APIToken != "xdg-token" {
		t.Fatalf("token = %q", cfg.Server.APIToken)
	}
}

func TestLoad_EnvOverridesProject(t *testing.T) {
	writeXDGConfig(t, xdgBase)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".vja.yaml"), []byte(
		"server:\n  api_url: https://corp.example.com/api/v1\n",
	), 0o600); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	chdir(t, root)
	t.Setenv("VJA_API_URL", "https://env.example.com/api/v1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.APIURL != "https://env.example.com/api/v1" {
		t.Fatalf("api_url = %q, want env value", cfg.Server.APIURL)
	}
}

func TestLoad_ProjectConfigDiscoveredFromSubdir(t *testing.T) {
	writeXDGConfig(t, xdgBase)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".vja.yaml"), []byte(
		"defaults:\n  project: nested-project\n",
	), 0o600); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	chdir(t, sub)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Defaults.Project.Name != "nested-project" {
		t.Fatalf("project = %q, want nested-project", cfg.Defaults.Project.Name)
	}
	if filepath.Base(cfg.ProjectConfigPath) != ".vja.yaml" {
		t.Fatalf("ProjectConfigPath = %q", cfg.ProjectConfigPath)
	}
}

func TestLoad_ProjectPathNotRewritten(t *testing.T) {
	path, _ := writeXDGConfig(t, xdgBase)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".vja.yaml"), []byte(
		"defaults:\n  project: work-project\n",
	), 0o600); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	chdir(t, root)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Path != path {
		t.Fatalf("Path = %q, want XDG path %q", cfg.Path, path)
	}
}

func TestLoad_NoXDGConfigReturnsNotFound(t *testing.T) {
	// XDG dir with no config file.
	t.Setenv("VJA_CONFIG_DIR", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".vja.yaml"), []byte(
		"server:\n  api_url: https://corp.example.com/api/v1\n",
	), 0o600); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	chdir(t, root)

	_, err := Load()
	if !errors.Is(err, ErrConfigNotFound) {
		t.Fatalf("err = %v, want ErrConfigNotFound", err)
	}
}

func TestProjectRef_UnmarshalYAML_StringAndInt(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var cfg Config
		if err := yaml.Unmarshal([]byte("defaults:\n  project: my-name\n"), &cfg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if cfg.Defaults.Project.Name != "my-name" || cfg.Defaults.Project.ID != nil {
			t.Fatalf("project = %+v", cfg.Defaults.Project)
		}
	})
	t.Run("int", func(t *testing.T) {
		var cfg Config
		if err := yaml.Unmarshal([]byte("defaults:\n  project: 42\n"), &cfg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if cfg.Defaults.Project.ID == nil || *cfg.Defaults.Project.ID != 42 || cfg.Defaults.Project.Name != "" {
			t.Fatalf("project = %+v", cfg.Defaults.Project)
		}
	})
}
