[![Build Status](https://travis-ci.org/thash/asana.svg?branch=master)](https://travis-ci.org/thash/asana)

OverView
=========================================

[Asana](https://asana.com/) command line client implemented in Go.


Install
=========================================

Requirements: go

### Mac OS X

    $ brew tap thash/asana
    $ brew install asana


### Others

    $ go get github.com/thash/asana


Usage
=========================================

    $ asana help

    NAME:
       asana - asana cui client ( https://github.com/thash/asana )

    USAGE:
       asana [global options] command [command options] [arguments...]

    VERSION:
       x.x.x

    COMMANDS:
       config, c            Asana configuration. Your settings will be saved in ~/.asana.yml
       workspaces, w        get workspaces
       tasks, ts            get tasks
       projects, ps         get projects
       sections, sec        get sections/columns of a project
       create, cr           create a task
       task, t              get a task
       comment, cm          Post comment
       comments, cms        list or read comments of a task
       done                 Complete task
       due                  set due date
       body                 set task body (notes)
       fields, cf           list custom fields of a project
       set-field, sf        set a custom field value on a task
       delete, rm           delete a task by gid
       browse, b            open a task in the web browser
       download, dl         download attachment from a task
       help, h              Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       --help, -h           show help
       --version, -v        print the version


### Configure


    $ asana config
    visit: http://app.asana.com/-/account_api
      Settings > Apps > Manage Developer Apps > Personal Access Tokens
      + Create New Personal Access Token

    paste your Personal Access Token: _ <Copy Token from URL above and paste it.>

![](https://raw.githubusercontent.com/ghost-vk/asana/images/token.webp)

When you paste valid token, your workspaces will be displayed.

    2 workspaces found.
    [0]    4444444444444 My Project
    [1]     999999999999 Work

    Choose one out of them: _

Select one workspace. Configurations are saved in `~/.asana.yml`.

    $ cat ~/.asana.yml

    personal_access_token: 0/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    workspace: 4444444444444


### Tasks

`asana tasks` or `asana ts` list your tasks.

    $ asana ts

    0 [ 2014-08-13 ] Write README.
    1 [ 2014-08-18 ] Buy gift for coworkers.
    2 [ 2014-08-29 ] Read "Unweaving the Rainbow".
    3 [            ] haircut

`asana task <index>` or `asana t <index>` shows the task in detail. When you run it without index, top of the tasks list will be used.

`-v` option adds comments and modification histories to the output.

    $ asana t -v 0

    [ 2014-08-13 ] Write README.
    --------
    Write README.md for Asana Cli project.

    ----------------------------------------

    assigned to you (2014-07-07T05:31:18.278Z)
    --------
    changed the name to "Write README." (2014-07-18T08:52:57.020Z)
    --------
    changed the due date to August 8 (2014-08-04T10:33:07.168Z)
    --------
    How about progress?
    by Lain Iwakura (2014-08-10T04:17:57.741Z)
    --------
    moved from Piyo to Hoge (2014-08-11T02:02:53.051Z)
    --------
    No progress.
    by Hash (2014-08-11T01:21:38.014Z)
    --------
    moved from Hoge to Piyo (2014-08-11T02:02:53.051Z)
    --------
    changed the due date to August 13 (2014-08-11T10:30:39.785Z)

`-p <project_gid>` lists tasks of a project instead of your own. Output shows the task gid, non-default type, section and due date (when set).

    $ asana ts -p 1202689990538470

    0 1214634990237303 Raw Signals 📋        Доработки по порталу
    1 1211297085634051 [S] Engineering Sprint Обновление баннера
    2 1207951833057398 milestone [R] Released Релиз v2


### Projects

`asana projects` or `asana ps` list projects of your workspace (gid + name).

    $ asana ps

    0 1202689990538470 ID-1916 [ДИТ] Редизайн ММЦ
    1 1202773036545383 Design System

Pass a query to search by name (server-side, whole workspace):

    $ asana ps ID-1916


### Sections

`asana sections -p <project_gid>` (alias `sec`) list the sections/columns of a project (gid + name). Results are cached per project for 5m; use `-n` to skip the cache or `-r` to refresh it.

    $ asana sec -p 1202689990538470

    0 1202689990657994 Landed Signals
    1 1202689990660101 Inbox 📬

Use the printed section gid as `-s` when creating a task.


### Create a task

`asana create` or `asana cr` create a task. Flags must come **before** the name.

    $ asana create "buy coffee"                                  # in your workspace
    $ asana create -p <project_gid> "task in a project"
    $ asana create -p <project_gid> -s <section_gid> "task in a column"
    $ asana create -p <project_gid> -b "task description here" "task name"

`-p` adds the task to a project, `-s` puts it into a section/column, `-b` sets the body (notes).


### Task body (notes)

`asana body <index> <text>` sets the body (the `notes` field) of a task picked by its index in the tasks list. Quote multi-word text.

    $ asana body 0 "Updated description, multiple words."

Newlines and quotes inside the text are preserved. Pass an empty string to clear the body.


### Custom fields

`asana fields -p <project_gid>` (alias `cf`) list the custom fields attached to a project, with their type and — for `enum` fields — the available options (gid + name).

    $ asana cf -p 1202689990538470

    1199105780031549 Type (enum)
      1198862357412458 Feature
      1199105780031551 Bug
    1199542488281141 Priority (enum)
      ...

`asana set-field` (alias `sf`) set a custom field value on a task:

    $ asana sf -t <task_gid> -f <field_gid> -V <value>

`-V` accepts:

- `enum` — the option **name** (case-insensitive, e.g. `Feature`) or its gid. An unknown name fails with the list of valid options.
- `text` — any string.
- `number` — the number.
- `null` — clears the field.

      $ asana sf -t 1214222140735157 -f 1199105780031549 -V Feature   # enum by name
      $ asana sf -t 1214222140735157 -f 1167156493787491 -V 8         # number
      $ asana sf -t 1214222140735157 -f 1199105780031549 -V null      # clear

Field and option gids come from `asana fields -p <project>`.


### Complete, set due on a task

To complete task, use `asana complete <index>` or `asana done <index>`.

    $ asana done 12

To change(or newly set) due date, use `asana due <index> <due_date>`.

    $ asana due 5 2014-08-21

Or, `today` or `tomorrow`.

    $ asana due 5 today


### Set task body

`asana body <index> <text>` sets the body (notes) of an existing task.

    $ asana body 2 "Updated description with details"


### Comment

`asana comment <index>` or `asana cm <index>` enable you to post new comment for the task.

    $ asana cm 2

This command opens editor. Write comment, save and close.

![](https://raw.githubusercontent.com/thash/asana/images/cmt.png)

You can change editor by updating `$EDITOR` environment variable.

`asana comments <index>` (alias `cms`) lists all comments on a task.

    $ asana cms 0

    0 1234567890123456  by Alice (2024-01-01T10:00:00.000Z)
    Great progress!

Pass `-g <story_gid>` to read a single comment by its gid.

    $ asana cms -g 1234567890123456


### Open a task in the browser

`asana browse <index>` or `asana b <index>` will open task in browser.

    $ asana browse 1
    // => open browser


### Custom fields

`asana fields -p <project_gid>` (alias `cf`) lists all custom fields for a project with their types and, for enum fields, available options.

    $ asana cf -p 1202689990538470

    1198862357412455 Type (enum)
      1198862357412458 Feature
      1199105780031551 Bug
    1199542488281138 Priority (enum)
      1199542488281141 High
      1199542488281142 Medium

`asana set-field` (alias `sf`) sets a custom field value on a task by GID.

    $ asana sf -t <task_gid> -f <field_gid> -V <value>

    # enum: pass the option gid
    $ asana sf -t 1216060636060282 -f 1198862357412455 -V 1198862357412458

    # text
    $ asana sf -t 1216060636060282 -f 1198862900000001 -V "in review"

    # number
    $ asana sf -t 1216060636060282 -f 1198862900000002 -V 42

    # clear (any type)
    $ asana sf -t 1216060636060282 -f 1198862357412455 -V null


TODO
=========================================

See [Issues](https://github.com/thash/asana/issues)
