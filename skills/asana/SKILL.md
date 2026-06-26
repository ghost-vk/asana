---
name: asana
description: Drive the `asana` command-line client (github.com/ghost-vk/asana) to manage Asana tasks from the terminal — list/read tasks, create them, complete, set due dates, edit body/comments, manage projects, sections and custom fields, download attachments. Use this whenever the user wants to interact with Asana via the CLI: "show my asana tasks", "create an asana task in project X", "mark task 3 done", "set due date", "comment on a task", "what's in this Asana project", or any time the `asana` command is the right tool. Read it before composing `asana` commands so you get the index/gid addressing and required flags right.
---

# asana CLI

Terminal client for Asana. Commands below; run `asana help` or `asana <cmd> -h` to confirm on an unknown version.

## Prerequisite: config

Needs a Personal Access Token in `~/.asana.yml`. If commands fail with auth/empty output, run `asana config` (interactive: prints the token URL, prompts for token, then workspace). Not installed at all? See the `asana-install` skill.

## The addressing model (most important thing)

Task-targeting commands accept **either an index or a GID**:

- **Index** (`0`, `1`, `2`, …) = position in the _last_ `asana ts` listing. Indices are read from a cache that `asana ts` writes. **So run `asana ts` (or `asana ts -p <project>`) first to populate/refresh indices**, then address tasks by their printed number. Cache lives 5 min.
- **GID** (a long numeric id, ≥10 digits) = used directly, no cache needed. Listings print the GID, so prefer passing the GID when you already have it — it's unambiguous and cache-independent.
- `delete` and `set-field` take a **GID only** (no index).

When an index is omitted, `task`, `due`, `comments` default to index `0` (top task); `done`, `body`, `download` require an explicit arg.

## Commands

| Command    | Aliases | Syntax                                                               | Notes                                                                                                             |
| ---------- | ------- | -------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| config     | c       | `asana config`                                                       | one-time token + workspace setup                                                                                  |
| workspaces | w       | `asana w`                                                            | list workspaces                                                                                                   |
| tasks      | ts      | `asana ts [-p <project>] [-l N] [-n] [-r]`                           | your tasks, or a project's with `-p`. Writes index cache. `-n` skip cache, `-r` refresh, `-l` limit (default 100) |
| task       | t       | `asana t [-v] [-j] [<index\|gid>]`                                   | one task detail. `-v` adds comments+history, `-j` JSON (task+stories+attachments)                                 |
| projects   | ps      | `asana ps [query] [-l N]`                                            | list projects; `query` searches by name server-side                                                               |
| sections   | sec     | `asana sec -p <project> [-n] [-r]`                                   | sections/columns of a project (cached per project)                                                                |
| create     | cr      | `asana cr [-p <project>] [-s <section>] [-b <body>] "<name>"`        | **flags before the name**. Prints new gid                                                                         |
| comment    | cm      | `asana cm <index\|gid>`                                              | opens `$EDITOR`; write, save, close to post                                                                       |
| comments   | cms     | `asana cms <index\|gid>` / `asana cms -g <story_gid>`                | list comments, or read one by story gid                                                                           |
| done       | —       | `asana done <index\|gid>`                                            | complete the task                                                                                                 |
| due        | —       | `asana due <index\|gid> <date>`                                      | date = `YYYY-MM-DD`, `today`, or `tomorrow`                                                                       |
| body       | —       | `asana body <index\|gid> "<text>"`                                   | set notes; empty string clears                                                                                    |
| fields     | cf      | `asana cf -p <project>`                                              | custom fields; enum fields list their options (gid+name)                                                          |
| set-field  | sf      | `asana sf -t <task_gid> -f <field_gid> -V <value>`                   | see value rules below. GID only                                                                                   |
| browse     | b       | `asana b <index\|gid>`                                               | open task in browser                                                                                              |
| download   | dl      | `asana dl <task_index> <att_index>` / `asana dl <att_gid> [-o path]` | attachment indices come from `asana t <index>`                                                                    |
| delete     | rm      | `asana rm <gid>`                                                     | delete by GID only                                                                                                |

## set-field values (`-V`)

- **enum** — option name (case-insensitive, e.g. `Feature`) or its gid. Unknown name fails listing valid options.
- **text** — any string.
- **number** — the number.
- **null** — clears the field.

Get field and option gids from `asana cf -p <project>`.

## Output shapes (for parsing)

- `ts` line: `<idx> <gid> [<type>] <section> [ <due> ] <name>` — type/section/due appear only when set.
- `ps` line: `<idx> <gid> <name>`.
- `sec` / enum options: `<gid> <name>` (cf top-level: `<gid> <name> (<type>)`).
- `cms` line: `<idx> <story_gid>  by <author> (<ts>)` then the comment text on the next line.
- `create` → `created <gid> <name>`; `done` → `DONE! : <name>`.

## Working pattern

1. `asana ps <query>` to find a project gid, or `asana ts` for your tasks.
2. To act inside a project: `asana ts -p <project_gid>` (this caches indices), then target tasks by the printed gid (robust) or index.
3. For project-scoped writes you usually need gids from `ps`/`sec`/`cf` first.
