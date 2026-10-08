package main

import (
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/commands"
)

// version is injected by goreleaser via ldflags; stays "dev" for `go build`.
var version = "dev"

// resolveVersion falls back to the module version embedded by `go install`,
// which ldflags don't cover, so `asana version` isn't stuck on "dev".
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	app := cli.NewApp()
	app.Name = "asana"
	app.Version = resolveVersion()
	app.Usage = "asana cui client ( https://github.com/ghost-vk/asana )"

	app.Commands = defs()
	app.Run(os.Args)
}

func defs() []*cli.Command {
	return []*cli.Command{
		{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Asana configuration. Your settings will be saved in ~/.asana.yml",
			Action: func(c *cli.Context) error {
				commands.Config(c)
				return nil
			},
		},
		{
			Name:    "workspaces",
			Aliases: []string{"w"},
			Usage:   "get workspaces",
			Action: func(c *cli.Context) error {
				commands.Workspaces(c)
				return nil
			},
		},
		{
			Name:    "tasks",
			Aliases: []string{"ts"},
			Usage:   "get tasks",
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "no-cache", Aliases: []string{"n"}, Usage: "without cache"},
				&cli.BoolFlag{Name: "refresh", Aliases: []string{"r"}, Usage: "update cache"},
				&cli.IntFlag{Name: "limit", Aliases: []string{"l"}, Value: 100, Usage: "max tasks to fetch; with -p every task unless set, 0 = no limit"},
				&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "tasks of a project (gid)"},
				&cli.BoolFlag{Name: "json", Aliases: []string{"j"}, Usage: "output as JSON (detailed fields)"},
				&cli.BoolFlag{Name: "completed", Aliases: []string{"all", "a"}, Usage: "include completed tasks (marked ✓)"},
				&cli.StringFlag{Name: "since", Usage: "with completed tasks, only those completed since YYYY-MM-DD (implies --completed)"},
			},
			Action: func(c *cli.Context) error {
				commands.Tasks(c)
				return nil
			},
		},
		{
			Name:    "projects",
			Aliases: []string{"ps"},
			Usage:   "get projects",
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "limit", Aliases: []string{"l"}, Value: 100, Usage: "max projects to fetch"},
			},
			Action: func(c *cli.Context) error {
				commands.Projects(c)
				return nil
			},
		},
		{
			Name:    "project",
			Aliases: []string{"p"},
			Usage:   "get project details",
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "json", Aliases: []string{"j"}, Usage: "output as JSON"},
			},
			Action: func(c *cli.Context) error {
				commands.Project(c)
				return nil
			},
		},
		{
			Name:    "sections",
			Aliases: []string{"sec"},
			Usage:   "get sections/columns of a project: sections -p <project>",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project gid"},
				&cli.BoolFlag{Name: "no-cache", Aliases: []string{"n"}, Usage: "without cache"},
				&cli.BoolFlag{Name: "refresh", Aliases: []string{"r"}, Usage: "update cache"},
			},
			Action: func(c *cli.Context) error {
				commands.Sections(c)
				return nil
			},
		},
		{
			Name:    "create",
			Aliases: []string{"cr"},
			Usage:   "create a task: create [-p project] [-s section] [-a email|gid] [--milestone|--subtype type] [--due date] [-b body|-f file] [--md|--html] <name> (flags before name)",
			Flags: append([]cli.Flag{
				&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project gid"},
				&cli.StringFlag{Name: "section", Aliases: []string{"s"}, Usage: "section/column gid"},
				&cli.StringFlag{Name: "assignee", Aliases: []string{"a"}, Usage: "assignee email or user gid"},
				&cli.BoolFlag{Name: "milestone", Usage: "create a milestone (same as --subtype milestone)"},
				&cli.StringFlag{Name: "subtype", Usage: "task type: default_task, milestone or approval"},
				&cli.StringFlag{Name: "due", Usage: "due date: YYYY-MM-DD, today or tomorrow"},
				&cli.StringFlag{Name: "body", Aliases: []string{"b"}, Usage: "task body"},
			}, commands.RichTextFlags("read the task body from a file (- for stdin)")...),
			Action: func(c *cli.Context) error {
				commands.CreateTask(c)
				return nil
			},
		},
		{
			Name:  "assign",
			Usage: "assign a task: assign <index|gid> <email|gid>",
			Action: func(c *cli.Context) error {
				commands.Assign(c)
				return nil
			},
		},
		{
			Name:  "move",
			Usage: "move or copy a task between projects",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "target project gid"},
				&cli.StringFlag{Name: "section", Aliases: []string{"s"}, Usage: "target section gid"},
				&cli.BoolFlag{Name: "copy", Aliases: []string{"c"}, Usage: "copy only; do not remove the source project"},
			},
			Action: func(c *cli.Context) error {
				commands.Move(c)
				return nil
			},
		},
		{
			Name:    "task",
			Aliases: []string{"t"},
			Usage:   "get a task",
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "verbose output"},
				&cli.BoolFlag{Name: "json", Aliases: []string{"j"}, Usage: "output as JSON"},
				&cli.BoolFlag{Name: "html", Usage: "print the rich-text body (html_notes) instead of notes"},
			},
			Action: func(c *cli.Context) error {
				commands.Task(c)
				return nil
			},
		},
		{
			Name:    "comment",
			Aliases: []string{"cm"},
			Usage:   "post a comment: comment [--md|--html] [-f file] <index>",
			Flags:   commands.RichTextFlags("read the comment from a file (- for stdin) instead of $EDITOR"),
			Action: func(c *cli.Context) error {
				commands.Comment(c)
				return nil
			},
		},
		{
			Name:    "comments",
			Aliases: []string{"cms"},
			Usage:   "list comments of a task (comments <index>) or read one (comments -g <story_gid>)",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "gid", Aliases: []string{"g"}, Usage: "story gid to fetch a single comment"},
			},
			Action: func(c *cli.Context) error {
				commands.Comments(c)
				return nil
			},
		},
		{
			Name:  "done",
			Usage: "Complete task",
			Action: func(c *cli.Context) error {
				commands.Done(c)
				return nil
			},
		},
		{
			Name:    "undone",
			Aliases: []string{"reopen"},
			Usage:   "reopen a completed task",
			Action: func(c *cli.Context) error {
				commands.Undone(c)
				return nil
			},
		},
		{
			Name:  "due",
			Usage: "set due date",
			Action: func(c *cli.Context) error {
				commands.DueOn(c)
				return nil
			},
		},
		{
			Name:  "body",
			Usage: "set task body: body [--md|--html] [-f file] <index> [text]",
			Flags: commands.RichTextFlags("read the body from a file (- for stdin)"),
			Action: func(c *cli.Context) error {
				commands.Body(c)
				return nil
			},
		},
		{
			Name:    "browse",
			Aliases: []string{"b"},
			Usage:   "open a task in the web browser",
			Action: func(c *cli.Context) error {
				commands.Browse(c)
				return nil
			},
		},
		{
			Name:    "fields",
			Aliases: []string{"cf"},
			Usage:   "list custom fields of a project",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "project", Aliases: []string{"p"}, Usage: "project gid"},
			},
			Action: func(c *cli.Context) error {
				commands.Fields(c)
				return nil
			},
		},
		{
			Name:    "set-field",
			Aliases: []string{"sf"},
			Usage:   "set a custom field on a task: -t <task> -f <field> -V <value|option_name|option_gid|null>",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "task", Aliases: []string{"t"}, Usage: "task gid"},
				&cli.StringFlag{Name: "field", Aliases: []string{"f"}, Usage: "custom field gid"},
				&cli.StringFlag{Name: "value", Aliases: []string{"V"}, Usage: "enum: option name or gid; text: string; number: value; null to clear"},
			},
			Action: func(c *cli.Context) error {
				commands.SetField(c)
				return nil
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"rm"},
			Usage:   "delete a task by gid",
			Action: func(c *cli.Context) error {
				commands.DeleteTask(c)
				return nil
			},
		},
		{
			Name:    "download",
			Aliases: []string{"dl"},
			Usage:   "download attachment from a task",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "output file path"},
			},
			Action: func(c *cli.Context) error {
				commands.Download(c)
				return nil
			},
		},
	}
}
