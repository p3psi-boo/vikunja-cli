# vikunja-cli

`vikunja-cli` (`vja`) is a stateless Go CLI for Vikunja.

It is designed for both interactive usage and script/agent workflows.

## Features

- Stateless: no local cache, only optional token file.
- XDG config/token lookup.
- Task, project, and label commands.
- Human-readable text output and machine-readable JSON output.
- Shell completion for bash/zsh/fish.

## Requirements

- Go 1.22+
- A reachable Vikunja API endpoint

## Build

```bash
go build -o vja .
```

Version defaults to `dev`.

Set build-time version:

```bash
go build -ldflags "-X main.version=v0.1.0" -o vja .
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
./vja done 42
./vja defer 42 2d
./vja clone 42 "Ship README v2"
./vja rm 42
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

## Output Modes

- `--json` (`-j`): print JSON to stdout.
- `--quiet` (`-q`): suppress non-data informational messages.

Examples:

```bash
./vja ls --json
./vja add "task from script" -p 1 --json
./vja rm 42 --quiet
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
