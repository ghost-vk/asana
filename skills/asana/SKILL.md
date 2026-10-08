---
name: asana
description: CLI for interacting with Asana. Use when working with Asana tasks, projects, or comments
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
- **Completed tasks are hidden by default.** `asana ts` lists open tasks only; a task that is not in the listing may simply be closed. Use `asana ts -p <project> --completed` (alias `--all`, `-a`) before concluding a task does not exist. Indices from a `--completed` listing include the closed tasks.

When an index is omitted, `task`, `due`, `comments` default to index `0` (top task); `assign`, `done`, `undone`, `body`, `download` require an explicit arg.

## Commands

| Command    | Aliases | Syntax                                                               | Notes                                                                                                             |
| ---------- | ------- | -------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| config     | c       | `asana config`                                                       | one-time token + workspace setup                                                                                  |
| workspaces | w       | `asana w`                                                            | list workspaces                                                                                                   |
| tasks      | ts      | `asana ts [-p <project>] [--completed] [--since YYYY-MM-DD] [-l N] [-n] [-r] [-j]` | your tasks, or a project's with `-p`. Writes index cache. Text output shows `[@assignee]` when set. `--completed`/`--all`/`-a` includes closed tasks (marked `✓`); `--since` limits those to tasks closed since a date. `-p` reads every page unless `-l` is set; your own tasks stop at `-l` (default 100). `-n` skip cache, `-r` refresh. `-j` JSON with full fields |
| task       | t       | `asana t [-v] [-j] [--html] [<index\|gid>]`                          | one task detail. `-v` adds comments+history, `-j` JSON (task+stories+attachments), `--html` prints `html_notes`   |
| projects   | ps      | `asana ps [query] [-l N]`                                            | list projects; `query` searches by name server-side                                                               |
| project    | p       | `asana p <gid> [-j]`                                                 | details for one project: name, URL, team, owner, dates, status, notes. `-j` for full JSON                        |
| sections   | sec     | `asana sec -p <project> [-n] [-r]`                                   | sections/columns of a project (cached per project)                                                                |
| create     | cr      | `asana cr [-p <project>] [-s <section>] [-a <email\|gid>] [--milestone\|--subtype <type>] [--due <date>] [-b <body>\|-f <file>] [--md\|--html] "<name>"` | **flags before the name**. `-a` sets the assignee. `--milestone` = `--subtype milestone`; types: `default_task`, `milestone`, `approval`. `--due` takes `YYYY-MM-DD`, `today`, `tomorrow`. Prints new gid |
| assign     | —       | `asana assign <index\|gid> <email\|gid>`                            | assign or reassign an existing task; accepts the standard index/GID addressing model                              |
| move       | —       | `asana move <index\|gid> -p <project> [-s <section>] [-c]`           | moves a task to another project/section; `-c` copies instead of removing the source project                      |
| comment    | cm      | `asana cm [--md\|--html] [-f <file>] <index\|gid>`                   | opens `$EDITOR`; write, save, close to post. `-f`/stdin skips the editor                                          |
| comments   | cms     | `asana cms <index\|gid>` / `asana cms -g <story_gid>`                | list comments, or read one by story gid                                                                           |
| done       | —       | `asana done <index\|gid>`                                            | complete the task                                                                                                 |
| undone     | reopen  | `asana undone <index\|gid>`                                          | reopen a completed task                                                                                           |
| due        | —       | `asana due <index\|gid> <date>`                                      | date = `YYYY-MM-DD`, `today`, or `tomorrow`                                                                       |
| body       | —       | `asana body [--md\|--html] [-f <file>] <index\|gid> ["<text>"]`      | set notes; `--md`/`--html` write `html_notes` instead. `-f -` reads stdin; `""` clears                             |
| fields     | cf      | `asana cf -p <project>`                                              | custom fields; enum fields list their options (gid+name)                                                          |
| set-field  | sf      | `asana sf -t <task_gid> -f <field_gid> -V <value>`                   | see value rules below. GID only                                                                                   |
| browse     | b       | `asana b <index\|gid>`                                               | open task in browser                                                                                              |
| download   | dl      | `asana dl <task_index> <att_index>` / `asana dl <att_gid> [-o path]` | attachment indices come from `asana t <index>`                                                                    |
| delete     | rm      | `asana rm <gid>`                                                     | delete by GID only                                                                                                |

## set-field values (`-V`)

- **enum** — option name (case-insensitive, e.g. `Feature`) or its gid. Unknown name fails listing valid options.
- **text** — any string.
- **number** — the number.
- **null** — clears the field (`-V null`).

Get field and option gids from `asana cf -p <project>`.

## Output shapes (for parsing)

- `ts` line: `<idx> <gid> [<type>] <section> [ <due> ] [@<assignee>] [✓ ]<name>` — type/section/due/assignee appear only when set; `✓` marks a completed task (only with `--completed`).
- `ts -j`: JSON array of task objects with `gid`, `name`, `completed`, `due_on`, `resource_subtype`, `memberships` (section), `assignee`, `custom_fields`.
- `ps` line: `<idx> <gid> <name>`.
- `p` text: `<gid>  <name>` then indented metadata lines; `p -j`: full `Project_t` JSON.
- `sec` / enum options: `<gid> <name>` (cf top-level: `<gid> <name> (<type>)`).
- `cms` line: `<idx> <story_gid>  by <author> (<ts>)` then the comment text on the next line.
- `create` → `created <gid> <name>`; `assign` → `assigned <gid> to <email|gid>`; `done` → `DONE! : <name>`; `undone` → `REOPENED : <name>`.

## Working pattern

1. `asana ps <query>` to find a project gid, or `asana ts` for your tasks.
2. To act inside a project: `asana ts -p <project_gid>` (this caches indices), then target tasks by the printed gid (robust) or index.
3. For project-scoped writes you usually need gids from `ps`/`sec`/`cf` first.
4. Use `asana move <index|gid> -p <target_project_gid> [-s <target_section_gid>] [-c]` to move or copy a task across projects.


## Rich text (`--md` / `--html`)

`body`, `create` and `comment` write plain text into `notes` by default. Two flags
switch them to Asana's rich-text field (`html_notes`, or `html_text` for comments):

- `--md` — the body is markdown and is converted before sending.
- `--html` — the body is already Asana rich-text HTML and is validated before sending.

Long documents do not belong on the command line: pass `-f <file>`, or `-f -` to read
stdin. Stdin is never read implicitly, so the CLI never hangs when run from a script.

    $ asana body --md <gid> -f spec.md
    $ cat spec.md | asana create --md -f - -p <project> "Spec"
    $ asana cm --md -f review.md <gid>

Read the rich-text body back with `asana t --html <gid>`; `asana t -j` includes
`html_notes` too.

### What Asana accepts

Asana parses `html_notes` as strict XML over a small tag whitelist and rejects
anything else with an opaque `xml_parsing_error`. The CLI validates locally first
and names the offending tag instead.

- Allowed: `body h1 h2 ul ol li strong em u s code pre blockquote a hr`.
- A single root `<body>` is required; `<p>` is **not** supported — paragraphs are
  plain newlines.
- Void tags must self-close: `<hr/>`, never `<hr>`.
- `&`, `<`, `>` must be escaped.

`--md` handles all of this. Constructs Asana cannot render are degraded rather than
dropped: `h3` and deeper collapse to `h2`, images become links, tables become text
rows. `<pre>` keeps newlines and indentation, so ASCII diagrams survive.

Links to other tasks are plain markdown links and stay clickable, e.g. a milestone
body listing its tasks:

    1. [Crop video](https://app.asana.com/0/0/1219065173477806)
    2. [Estimate](https://app.asana.com/0/0/1218737822731836)

With `--html`, `<a data-asana-gid="<gid>"/>` renders as a native task mention.
