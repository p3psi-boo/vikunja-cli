---
name: managing-vikunja-tasks
description: "Manages tasks, projects, and labels in Vikunja via the vja CLI. Use when asked to create, list, edit, complete, defer, clone, or delete tasks, manage projects, or work with labels."
---

# Managing Vikunja Tasks with vja

`vja` is a stateless CLI for Vikunja located at `vja` in the project root.

## Key Principles

- **Always use `--json` (`-j`)** for machine-readable output. Parse JSON results to extract IDs and data.
- **Always use `--quiet` (`-q`)** to suppress informational messages when chaining commands.
- **Never use interactive login.** Provide `--username`, `--password`, and `--totp` flags, or rely on pre-configured `VJA_API_TOKEN` / token file.
- **Projects and labels accept names or IDs.** Prefer IDs when known; use exact title strings otherwise.
- **Check exit codes:** 0 = success, 1 = API error, 2 = usage error, 3 = auth error, 4 = not found.

## Command Reference

### List Tasks

```bash
vja ls --json                          # all open tasks
vja ls -p "Inbox" -n 10 --json        # 10 tasks from project "Inbox"
vja ls -l backend -f --json           # favorite tasks with label "backend"
vja ls -a --json                       # include completed tasks
vja ls -d tomorrow --json             # tasks due by tomorrow
vja ls -s "due_date asc" --json       # sorted by due date
vja ls --filter "priority >= 3" --json # raw filter
```

Flags: `--all` (`-a`), `--project` (`-p`), `--label` (`-l`, repeatable), `--priority`, `--due` (`-d`), `--favorite` (`-f`), `--filter` (repeatable), `--sort` (`-s`), `--limit` (`-n`).

### Show Task Details

```bash
vja show 42 --json          # single task
vja show 42 43 44 --json    # multiple tasks
```

Returns full task details: ID, title, description, done, due date, project, priority, labels, timestamps.

### Create Task

```bash
vja add "Task title" -p "Inbox" --json
vja add "Ship docs" -p 1 -d tomorrow -l backend --prio 3 -r due -f --json
```

Flags: `--project` (`-p`, required or set `defaults.project`), `--due` (`-d`), `--label` (`-l`, repeatable), `--priority`/`--prio`, `--note` (`-n`), `--reminder` (`-r`, date or `due`), `--favorite` (`-f`).

### Edit Task

```bash
vja edit 42 --title "New title" --json
vja edit 42 -d "next monday" --note-append "Updated info" --json
vja edit 42 43 --prio 5 -l urgent --json   # batch edit
```

Flags: `--title` (`-t`), `--note` (`-n`), `--note-append`, `--project` (`-p`), `--due` (`-d`), `--priority`/`--prio`, `--label` (`-l`, toggles on/off), `--favorite` (`-f`), `--done`, `--reminder` (`-r`). Supports multiple task IDs.

### Complete / Toggle Done

```bash
vja done 42 --json          # toggle done state
vja done 42 43 44 --json    # batch toggle
```

### Defer Task

```bash
vja defer 42 2d --json      # defer by 2 days
vja defer 42 43 1w --json   # defer multiple tasks by 1 week
```

Duration format: `Nd` (days), `Nw` (weeks), `Nh` (hours).

### Clone Task

```bash
vja clone 42 --json                    # clone with same title
vja clone 42 "Cloned task title" --json # clone with new title
```

### Delete Task

```bash
vja rm 42 --json
vja rm 42 43 --json         # batch delete
```

### Projects

```bash
vja project ls --json                       # list all projects
vja project show 1 --json                   # project details
vja project add "Operations" --parent "Work" --json  # create project
```

### Labels

```bash
vja label ls --json          # list all labels
vja label add "backend" --json  # create label
```

### Auth (if needed)

```bash
vja login --username alice --password '***' --totp 123456
vja user --json              # verify current user
vja logout
```

## Date Expressions

Due dates and reminders accept flexible date expressions:

- Absolute: `2026-03-01`, `2026-03-01T15:00:00`
- Relative: `tomorrow`, `today`, `next monday`
- Durations (defer only): `2d`, `1w`, `3h`

## Workflow Examples

### Create a task and capture its ID

```bash
TASK_JSON=$(vja add "Write tests" -p "Inbox" -d tomorrow -l dev --json)
TASK_ID=$(echo "$TASK_JSON" | jq '.id')
```

### List tasks, find one, then edit it

```bash
vja ls -p "Work" --json | jq '.[] | select(.title | test("deploy"))'
vja edit 42 --note-append "Deployed to staging" --done --json
```

### Batch operations

```bash
# Complete all tasks matching criteria
IDS=$(vja ls -p "Sprint 1" --json | jq -r '.[].id')
for id in $IDS; do vja done "$id" -q; done
```

## Error Handling

- On exit code 3 (auth error): check token configuration, run `vja user --json` to verify auth.
- On exit code 4 (not found): verify the task/project ID exists with `show`.
- With `--json`, errors are written to stderr as `{"error":"...","code":N}`.
