package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/p3psi-boo/vikunja-cli/output"
	"github.com/spf13/cobra"
)

// skillContent holds the SKILL.md body injected from the main package (see
// SetSkillContent). It is embedded there because `go:embed` cannot reach
// above the embedding package's directory.
var skillContent string

// SetSkillContent records the SKILL.md body embedded by the main package.
func SetSkillContent(content string) {
	skillContent = content
}

const (
	skillDirGlobal  = ".agents/skills/vja"
	skillDirProject = ".agents/skills/vja"
	skillFileName   = "SKILL.md"
)

var (
	skillScope string
	skillForce bool
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Install the vja agent skill (SKILL.md)",
	Long: `Install the vja agent skill so coding agents know how to drive vja.

The skill is written to:
  --scope global   ~/.agents/skills/vja/SKILL.md      (default)
  --scope project  ./.agents/skills/vja/SKILL.md      (relative to CWD)

Existing files are left untouched unless --force is given.`,
	Args: cobra.NoArgs,
	RunE: runSkillInstall,
}

func init() {
	skillCmd.Flags().StringVar(&skillScope, "scope", "global", "Install target: global or project")
	skillCmd.Flags().BoolVar(&skillForce, "force", false, "Overwrite an existing SKILL.md")
	rootCmd.AddCommand(skillCmd)
}

func runSkillInstall(cmd *cobra.Command, _ []string) error {
	if skillContent == "" {
		return fmt.Errorf("no skill content embedded in this build")
	}

	scope := skillScope
	switch scope {
	case "global", "project":
	default:
		return fmt.Errorf("invalid --scope %q (expected global or project)", scope)
	}

	targetDir, err := resolveSkillDir(scope)
	if err != nil {
		return err
	}
	targetPath := filepath.Join(targetDir, skillFileName)

	if exists, err := fileExists(targetPath); err != nil {
		return fmt.Errorf("check existing skill %q: %w", targetPath, err)
	} else if exists && !skillForce {
		return fmt.Errorf("skill already exists at %q (use --force to overwrite)", targetPath)
	}

	dryRun := flagDryRun
	if dryRun {
		verb := "install"
		if exists, _ := fileExists(targetPath); exists {
			verb = "overwrite"
		}
		return output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "[dry-run] would %s skill to %s\n", verb, targetPath)
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create skill directory %q: %w", targetDir, err)
	}

	if err := os.WriteFile(targetPath, []byte(skillContent), 0o644); err != nil {
		return fmt.Errorf("write skill %q: %w", targetPath, err)
	}

	return output.PrintInfo(cmd.OutOrStdout(), flagQuiet, "Installed skill to %s\n", targetPath)
}

// resolveSkillDir returns the directory the skill should be written to for the
// given scope. Global targets the user's home directory; project targets the
// current working directory.
func resolveSkillDir(scope string) (string, error) {
	switch scope {
	case "global":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, skillDirGlobal), nil
	case "project":
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve working directory: %w", err)
		}
		return filepath.Join(cwd, skillDirProject), nil
	default:
		return "", fmt.Errorf("invalid scope %q", scope)
	}
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
