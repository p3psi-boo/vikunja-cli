# vikunja-cli

`vikunja-cli` (`vja`) is a stateless Go CLI for Vikunja.

It is designed for both interactive usage and script/agent workflows.

## Features

- Stateless: no local cache, only optional token file.
- XDG config/token lookup, plus per-project `.vja.yaml` overrides.
- Task, project, and label commands.
- Human-readable text output and machine-readable JSON output.
- Colorized output with urgency highlighting (overdue / due soon / done),
  auto-disabled on pipes and when `NO_COLOR` is set.
- Top-level shortcuts: `vja ls`, `vja show`, `vja add`, `vja rm`, ...
- `--dry-run` previews for write commands; confirmation prompts for deletes.
- Shell completion for bash/zsh/fish.

## Requirements

- Go 1.22+
- A reachable Vikunja API endpoint

## Build

```bash
go build -o vja .
```

Version defaults to `dev`. `vja version` prints the version, commit, and
build date, which can be injected at build time:

```bash
go build -ldflags "-X main.version=v0.1.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o vja .
```

## Quick Start

1) Create config file at `~/.config/vja/config.toml`.

2) Login or provide `VJA_API_TOKEN`.

3) Run `./vja ls`.

## Configuration

Config file lookup order:

1. `$VJA_CONFIG_DIR/config.toml`
2. `$XDG_CONFIG_HOME/vja/config.toml`
3. `$HOME/.config/vja/config.toml`

Example:

```toml
[server]
api_url = "https://vikunja.example.com/api/v1"
frontend_url = "https://vikunja.example.com"
# Optional static token. If set, token file login is skipped.
# api_token = "tk_..."

[defaults]
# Optional. ID or exact project title used by `vja task add`.
# project = "Inbox"

[output]
# Optional: "text" (default) or "json"
format = "text"
```

## Project-local Configuration

A project-local `.vja.yaml` file can override selected fields per working
directory. It is discovered by walking up from the current working directory to
the filesystem root (like `.git`), so it applies from any subdirectory of the
project.

Precedence (highest first):

1. Command-line flags
2. Environment variables (`VJA_API_URL`, `VJA_API_TOKEN`)
3. Project-local `.vja.yaml`
4. XDG `config.toml`

The project file is YAML and supports `defaults.project`, `server.*`, and
`output.format`. Only non-empty fields are overlaid; everything else falls back
to the XDG config. The XDG config is still required (for `server.api_url` at
minimum) — `.vja.yaml` only layers on top, it never replaces global login state.

`.vja.yaml` example:

```yaml
# Pin the default project for this repository, so `vja add "..."` lands there.
defaults:
  project: my-work-project   # project title (string) or id (integer)

# Optional overrides:
server:
  api_url: https://vikunja.corp.example.com/api/v1
output:
  format: json
```

The easiest way to set `defaults.project` for the current working directory is
`vja project use <project>` (ID or title). Titles are validated against the
server before being stored. Use `vja project use --unset` to clear it again.

Token file lookup order:

1. `$VJA_CONFIG_DIR/token.json`
2. `$XDG_STATE_HOME/vja/token.json`
3. `$HOME/.local/state/vja/token.json`

Token file format:

```json
{"token":"<jwt>"}
```

## Environment Overrides

- `VJA_API_URL` overrides `server.api_url`
- `VJA_API_TOKEN` overrides `server.api_token`
- `VJA_CONFIG_DIR` overrides config and token base directory

## Common Commands

```bash
# auth
./vja login
./vja login --username alice --password '***' --totp 123456
./vja logout
./vja user

# tasks
./vja ls
./vja task list -p 1 -l backend -f -n 20
./vja show 42
./vja add "Ship docs" -p 1 -d tomorrow -l backend
./vja edit 42 --title "Ship README" --note-append "include examples"
./vja done 42        # mark done (errors if already done; use `undone`)
./vja undone 42      # reopen a task
./vja toggle 42      # flip done state (alias: `check`)
./vja defer 42 2d    # push due date forward
./vja defer 42 2d --set-due   # set a due date when none exists
./vja clone 42 "Ship README v2"
./vja rm 42          # prompts in a TTY; -y skips the prompt
./vja open 42

# projects
./vja project ls
./vja project show 1
./vja project add "Operations" --parent "Work"
./vja project open 1

# labels
./vja label ls
./vja label add "backend"
```

## Install the Agent Skill

`vja skill` writes the bundled `SKILL.md` (the `managing-vikunja-tasks` skill)
so that coding agents know how to drive `vja`. The skill is embedded in the
binary at build time.

```bash
./vja skill                          # ~/.agents/skills/vja/SKILL.md  (default)
./vja skill --scope project          # ./.agents/skills/vja/SKILL.md  (current dir)
./vja skill --force                  # overwrite an existing file
./vja skill --dry-run                # preview the target path
```

`--scope global` (default) targets `~/.agents/skills/vja/SKILL.md`;
`--scope project` targets `./.agents/skills/vja/SKILL.md`. An existing file is
left untouched unless `--force` is given.


## Output Modes

- `--json` (`-j`): print JSON to stdout.
- `--quiet` (`-q`): suppress non-data informational messages.

Examples:

```bash
./vja ls --json
./vja add "task from script" -p 1 --json
./vja rm 42 --quiet
```

## Color

Output is colorized by default in an interactive terminal:

- Overdue dates and high priorities are highlighted.
- Completed tasks are muted.
- Favorites are marked with `★` and done tasks with `✓`.

Control it with the global `--color` flag or the `NO_COLOR` environment variable:

```bash
./vja --color=never ls     # disable color
./vja --color=always ls    # force color (e.g. into a pager)
NO_COLOR=1 ./vja ls        # also disables color
```

Color is automatically turned off when stdout is piped or redirected, so
`vja ls | grep` stays clean.

## Dry Run

Pass `--dry-run` to any write command (`add`, `edit`, `clone`, `delete`,
`defer`, `done`/`undone`/`toggle`, `project add`) to preview the action
without making changes:

```bash
./vja edit 42 --title "New" --dry-run
./vja rm 41 42 43 --dry-run
```

## Exit Codes

- `0`: success
- `1`: general error / API error
- `2`: usage error
- `3`: authentication required/failed
- `4`: resource not found

## Shell Completion

```bash
./vja completion bash
./vja completion zsh
./vja completion fish
```

`--project` and `--label` support dynamic completion (from API) on relevant commands.
