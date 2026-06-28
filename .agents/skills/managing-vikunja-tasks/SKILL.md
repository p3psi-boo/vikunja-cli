---
name: vja
description: "Manages tasks, projects, and labels in Vikunja via the vja CLI. Use when asked to create, list, view, edit, complete, reopen, toggle, defer, clone, or delete tasks, manage projects, work with labels, or otherwise operate on a Vikunja instance from the shell."
---

# Managing Vikunja Tasks with vja

`vja` is a stateless Go CLI for Vikunja. The binary lives at `vja` in the project
root (`just build` or `go build -o vja .`). Every task subcommand has a
**top-level shortcut**, so `vja ls` ≡ `vja task list` and `vja add` ≡ `vja task add`.

## Critical Operating Rules

- **Always pass `--json` (`-j`)** for machine-readable output. Parse JSON to read IDs and data.
- **Always pass `--quiet` (`-q`)** to suppress human info lines when chaining commands.
- **Never use interactive login.** Provide `--username`/`--password`/`--totp`, or rely on a preconfigured `VJA_API_TOKEN` / token file. Interactive prompts will hang a non-TTY agent.
- **Projects and labels accept an ID or a title string.** Prefer IDs once known; titles are matched case-insensitively (exact → prefix → substring), and an ambiguous substring match errors and lists candidates.
- **`done` ≠ `toggle`.** See below — getting this wrong is the most common mistake.
- **Check for `.vja.yaml` before assuming `-p` is required.** A project-local file may pin `defaults.project`, so `vja add "..."` lands in that project with no `--project` flag. See *Configuration & Project-Local Defaults* below.
- **Exit codes:** `0` success · `1` API error · `2` usage error · `3` auth error · `4` not found.
  In `--json` mode, errors are emitted on stderr as `{"error":"...","code":N}`.

## Task Done State — read this carefully

Three distinct commands; do not interchange them:

```bash
vja done 42            # Mark complete. ERRORS if already done (tells you to use `undone`).
vja undone 42          # Reopen. ERRORS if not done (tells you to use `done`).
vja toggle 42          # Flip current state. Alias: vja check 42
```

All three accept multiple IDs (`vja done 42 43 44`) and `--json`/`--dry-run`.
If you do not know the current state, use `toggle` (idempotent). If you want to
force a specific state and fail loudly on mismatch, use `done`/`undone`.
`vja edit 42 --done=true` also sets state directly.

## List Tasks

```bash
vja ls --json                              # open tasks (incomplete only, by default)
vja ls -p "Inbox" -n 10 --json             # in a project, limited to 10
vja ls -l backend -l urgent -f --json      # with labels, favorites only (labels repeatable, AND)
vja ls -a --json                           # include completed tasks
vja ls -d tomorrow --json                  # due by a date
vja ls -s "-priority,due_date" --json      # sort expression
vja ls --filter "due_date < now" --json    # raw Vikunja filter (repeatable, AND)
vja ls --absolute                          # show absolute dates alongside relative
vja ls --no-summary                        # hide the "N tasks (X overdue...)" tally line
```

Flags: `--all` (`-a`), `--project` (`-p`), `--label` (`-l`, repeatable),
`--priority`, `--due` (`-d`), `--favorite` (`-f`), `--filter` (repeatable),
`--sort` (`-s`), `--limit` (`-n`), `--absolute`, `--no-summary`.

## Show Task Details

```bash
vja show 42 --json          # single → JSON object
vja show 42 43 44 --json    # multiple → JSON array
```

Returns full task details: id, title, description, done, due date, project,
priority, labels, favorite, created/updated timestamps.

## Create Task

```bash
vja add "Task title" -p "Inbox" --json
vja add "Ship docs" -p 1 -d tomorrow -l backend --prio 3 -f --json
vja add "Review PR" -p Work -d "next monday" -r          # -r with no value ⇒ reminder at due date
```

Flags: `--project` (`-p`, **required** unless `defaults.project` is set),
`--due` (`-d`), `--label` (`-l`, repeatable), `--priority`/`--prio` (int),
`--note` (`-n`), `--reminder` (`-r`; bare `-r` means "remind at the due date"),
`--favorite` (`-f`).

## Edit Task

```bash
vja edit 42 --title "New title" --json
vja edit 42 -d "next monday" --note-append "Updated info" --json
vja edit 42 43 --prio 5 -l urgent --json   # batch edit (IDs repeat)
```

Flags: `--title` (`-t`), `--note` (`-n`), `--note-append`, `--project` (`-p`),
`--due` (`-d`), `--priority`/`--prio`, `--label` (`-l`, **toggles** the label on/off),
`--favorite` (`-f`), `--done` (bool), `--reminder` (`-r`). Supports multiple IDs.

Note: on `edit`, `--label` **toggles** membership (adds if absent, removes if present),
unlike `add` which only adds.

## Defer Task

```bash
vja defer 42 2d --json          # push due date (+ reminder) forward by 2 days
vja defer 42 43 1w --json       # defer multiple tasks by 1 week
vja defer 42 3h --set-due       # task has no due date: bootstrap it to now + duration
```

Behavior: past dates are shifted from *now*; future dates have the duration added.
A reminder is shifted alongside. Tasks with no due date and no reminder **error**
unless `--set-due` is given (then due = now + duration). Duration units: `w`, `d`,
`h`, `m`, `s`, and they **combine** (e.g. `1w2d`, `2h30m`).

## Clone Task

```bash
vja clone 42 --json                     # clone with same title
vja clone 42 "Cloned task title" --json # clone with a new title
```

Copies description, due date, priority, labels, reminders, favorite, and done state.

## Delete Task

```bash
vja rm 42 --json                # JSON mode never prompts
vja rm 42 -y                    # skip the TTY confirmation prompt
vja rm 42 43 --json             # batch
```

In a TTY without `--json`/`--quiet`/`-y`, deletion **prompts** for confirmation.
Use `-y` (or `--json`/`-q`) to stay non-interactive.

## Open in Browser

```bash
vja open 42            # open task 42 in the frontend
vja open 42 43         # open several
vja open               # open the Vikunja frontend root
```

Requires `server.frontend_url` in config. (`vja project open <id>` also exists.)

## Projects

```bash
vja project ls --json                              # list (alias: project list)
vja project show 1 --json                          # details
vja project add "Operations" --parent "Work" --json
vja project open 1
vja project use "Work"                             # pin default project for this repo (writes .vja.yaml)
vja project use 1                                  # pin by id (no API call)
vja project use --unset                            # clear the repo's default project
```

`project use` is the easy way to set `defaults.project` for the current working
directory (see *Configuration & Project-Local Defaults*). It accepts a project
ID or title; titles are validated against the server (must match one project)
before being stored. `--unset` removes the pin. Honors `--dry-run`/`-q`.

## Labels

```bash
vja label ls --json              # list
vja label add "backend" --json   # create
```

## Auth (only if no token is configured)

```bash
vja login --api-url https://vikunja.example.com/api/v1 --username alice --password '***' --totp 123456
vja user --json                  # verify current user
vja logout
```

## Date & Duration Expressions

Due dates and reminders accept flexible expressions (parsed via `olebedev/when`):

- Absolute: `2026-03-01`, `2026-03-01T15:00:00`, `2026-03-01 15:04`.
- Natural language: `tomorrow`, `today`, `next monday`, `in 3 days`, `friday`.
- `--reminder`/`-r`: a date expression, or the bare word `due` (remind at the due date).
- Defer durations (defer only): `1w`, `2d`, `3h`, `30m`, combinable like `1d2h`.

## Global Flags (apply to every command)

`--json`/`-j`, `--quiet`/`-q`, `--verbose`/`-v`, `--color auto|always|never`
(honors `NO_COLOR`; auto-disables when piped), `--dry-run` (preview writes without
changing anything — works on `add`, `edit`, `clone`, `defer`, `done`/`undone`/`toggle`,
`rm`, `project add`, `project use`), `--version`.

## Shell Completion

`vja` ships cobra-generated completion for bash, zsh, and fish:

```bash
vja completion zsh > "${fpath[1]}/_vja"   # or bash/fish; see `vja completion --help`
```

Once installed, dynamic completions suggest: project IDs+titles for `--project`/`-p`
and the `project use <project>` argument, label IDs+titles for `--label`/`-l`. These
hit the live API on demand, so they require a configured token.

## Configuration & Project-Local Defaults

`vja` is stateless — it reads config at startup, no local cache. The base config is
a TOML file found at `$VJA_CONFIG_DIR/config.toml`, `$XDG_CONFIG_HOME/vja/config.toml`,
or `~/.config/vja/config.toml`. It must define at least `server.api_url`.

On top of that, a **project-local `.vja.yaml`** can layer per-repo overrides. It is
discovered by walking up from the current working directory to the root (like `.git`),
so it applies from any subdirectory of the repo:

```yaml
# .vja.yaml — pins the default project for `vja add` in this repo
defaults:
  project: my-work-project   # project title (string) or id (integer)
# Optional overrides (only non-empty fields overlay the global config):
server:
  api_url: https://vikunja.corp.example.com/api/v1
output:
  format: json
```

Precedence, highest first: **flags > env vars (`VJA_API_URL`, `VJA_API_TOKEN`,
`VJA_CONFIG_DIR`) > `.vja.yaml` > global `config.toml`**. The global config is still
required for login state — `.vja.yaml` only overlays, it never replaces it.

**Agent guidance:** before creating a task, glance for a `.vja.yaml` in the project
tree. If it sets `defaults.project`, omit `-p` (or you may double-resolve). If it
sets `output.format: json`, output is already JSON even without `-j` — but pass `-j`
explicitly anyway to be safe across repos.

To pin the default project for the current repo, use `vja project use <project>`
(ID or title; titles are validated against the server). Clear it with
`vja project use --unset`.

## Workflow Recipes

### Create a task and capture its ID

```bash
TASK_ID=$(vja add "Write tests" -p "Inbox" -d tomorrow -l dev --json | jq '.id')
```

### Find a task by title, then edit it

```bash
vja ls -p "Work" --json | jq '.[] | select(.title | test("deploy"; "i"))'
vja edit 42 --note-append "Deployed to staging" --done --json
```

### Complete all tasks in a project

```bash
IDS=$(vja ls -p "Sprint 1" --json | jq -r '[.[] | select(.done==false) | .id] | @sh')
# safer than blind `done`: toggle is idempotent if state is already as desired
for id in $IDS; do vja done "$id" -q; done
```

## Error Handling

- **Exit 3 (auth):** token missing/expired. Verify with `vja user --json`; re-run
  `vja login` with flags, or set `VJA_API_TOKEN`.
- **Exit 4 (not found):** confirm the task/project/label ID exists with `show`/`ls`.
- **`done` on an already-done task:** intentional error — switch to `undone` or `toggle`.
- **Ambiguous project/label title:** the error lists candidate IDs; pass an exact ID.
- **With `--json`:** errors go to stderr as `{"error":"...","code":N}`; stdout stays clean.
