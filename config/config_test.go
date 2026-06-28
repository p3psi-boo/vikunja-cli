package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	// XDG dir with no config file. Clear HOME and XDG_CONFIG_HOME too so the
	// fallback candidates don't pick up a real config from the dev machine.
	t.Setenv("VJA_CONFIG_DIR", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", t.TempDir())
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

// TestSave_LoadRoundTrip guards against a regression where ProjectRef was
// serialized as a TOML table that could not be read back. Save must emit a
// scalar (or omit the field) for every shape Load accepts.
func TestSave_LoadRoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		project ProjectRef
	}{
		{name: "empty", project: ProjectRef{}},
		{name: "name", project: ProjectRef{Name: "Inbox"}},
		{name: "id", project: ProjectRef{ID: ptrInt64(7)}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xdg := t.TempDir()
			t.Setenv("VJA_CONFIG_DIR", xdg)
			t.Setenv("XDG_CONFIG_HOME", "")

			in := &Config{
				Server:   ServerConfig{APIURL: "https://example.com/api/v1"},
				Defaults: DefaultsConfig{Project: tc.project},
			}
			if err := Save(in, ""); err != nil {
				t.Fatalf("Save: %v", err)
			}

			chdir(t, t.TempDir())
			out, err := Load()
			if err != nil {
				t.Fatalf("Load after Save: %v", err)
			}
			if out.Defaults.Project.Name != tc.project.Name {
				t.Fatalf("project name = %q, want %q", out.Defaults.Project.Name, tc.project.Name)
			}
			if (out.Defaults.Project.ID == nil) != (tc.project.ID == nil) {
				t.Fatalf("project id = %v, want %v", out.Defaults.Project.ID, tc.project.ID)
			}
			if out.Defaults.Project.ID != nil && *out.Defaults.Project.ID != *tc.project.ID {
				t.Fatalf("project id = %d, want %d", *out.Defaults.Project.ID, *tc.project.ID)
			}
		})
	}
}

func ptrInt64(v int64) *int64 { return &v }

func TestProjectRef_MarshalYAML_Scalar(t *testing.T) {
	cases := []struct {
		name    string
		ref     ProjectRef
		want    string
	}{
		{name: "name", ref: ProjectRef{Name: "Inbox"}, want: "Inbox\n"},
		{name: "id", ref: ProjectRef{ID: ptrInt64(7)}, want: "7\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := yaml.Marshal(tc.ref)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(out) != tc.want {
				t.Fatalf("marshal = %q, want %q", string(out), tc.want)
			}

			// Round-trip: what we write must parse back to the same reference.
			var back ProjectRef
			if err := yaml.Unmarshal(out, &back); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if back.Name != tc.ref.Name {
				t.Fatalf("name = %q, want %q", back.Name, tc.ref.Name)
			}
			if (back.ID == nil) != (tc.ref.ID == nil) {
				t.Fatalf("id presence = %v, want %v", back.ID == nil, tc.ref.ID == nil)
			}
			if back.ID != nil && *back.ID != *tc.ref.ID {
				t.Fatalf("id = %d, want %d", *back.ID, *tc.ref.ID)
			}
		})
	}
}

func TestSaveProjectDefault_WritesAndLoads(t *testing.T) {
	writeXDGConfig(t, xdgBase)
	root := t.TempDir()
	chdir(t, root)

	if _, err := SaveProjectDefault(ProjectRef{Name: "work-project"}); err != nil {
		t.Fatalf("SaveProjectDefault: %v", err)
	}

	// File exists in the CWD.
	if _, err := os.Stat(filepath.Join(root, ".vja.yaml")); err != nil {
		t.Fatalf("expected .vja.yaml: %v", err)
	}

	// Load picks up the override on top of the XDG config.
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Defaults.Project.Name != "work-project" {
		t.Fatalf("project = %q, want work-project", cfg.Defaults.Project.Name)
	}
}

func TestSaveProjectDefault_PreservesOtherFields(t *testing.T) {
	writeXDGConfig(t, xdgBase)
	root := t.TempDir()
	existing := "server:\n  api_url: https://corp.example.com/api/v1\noutput:\n  format: json\n"
	if err := os.WriteFile(filepath.Join(root, ".vja.yaml"), []byte(existing), 0o600); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	chdir(t, root)

	if _, err := SaveProjectDefault(ProjectRef{Name: "work-project"}); err != nil {
		t.Fatalf("SaveProjectDefault: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".vja.yaml"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "https://corp.example.com/api/v1") {
		t.Fatalf("server.api_url not preserved:\n%s", body)
	}
	if !strings.Contains(body, "format: json") {
		t.Fatalf("output.format not preserved:\n%s", body)
	}
	if !strings.Contains(body, "work-project") {
		t.Fatalf("project not written:\n%s", body)
	}

	// Load reflects the merged result.
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.APIURL != "https://corp.example.com/api/v1" {
		t.Fatalf("api_url = %q", cfg.Server.APIURL)
	}
	if cfg.Defaults.Project.Name != "work-project" {
		t.Fatalf("project = %q", cfg.Defaults.Project.Name)
	}
}

func TestSaveProjectDefault_UnsetClearsProject(t *testing.T) {
	writeXDGConfig(t, xdgBase)
	root := t.TempDir()
	existing := "defaults:\n  project: work-project\n"
	if err := os.WriteFile(filepath.Join(root, ".vja.yaml"), []byte(existing), 0o600); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	chdir(t, root)

	if _, err := SaveProjectDefault(ProjectRef{}); err != nil {
		t.Fatalf("SaveProjectDefault unset: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".vja.yaml"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if strings.Contains(string(data), "project") {
		t.Fatalf("expected project to be gone:\n%s", string(data))
	}

	// Load falls back to the XDG default.
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Defaults.Project.Name != "xdg-project" {
		t.Fatalf("project = %q, want xdg-project fallback", cfg.Defaults.Project.Name)
	}
}
