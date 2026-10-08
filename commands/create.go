package commands

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

var taskSubtypes = map[string]bool{"default_task": true, "milestone": true, "approval": true}

func CreateTask(c *cli.Context) {
	name := c.Args().First()
	if name == "" {
		log.Fatal("fatal: task name is required")
	}
	opts, err := taskOptions(c.String("assignee"), c.String("subtype"), c.Bool("milestone"), c.String("due"))
	if err != nil {
		log.Fatal("fatal: " + err.Error())
	}
	var payload string
	var html bool
	if body, _ := readBody(c, c.String("body"), c.IsSet("body")); body != "" {
		payload, html = prepareBody(body, resolveFormat(c))
	}
	t := api.CreateTaskWithOptions(name, c.String("project"), c.String("section"), payload, html, opts)
	fmt.Printf("created %s %s\n", t.Gid, t.Name)
}

func taskOptions(assignee, subtype string, milestone bool, due string) (api.TaskOptions, error) {
	if milestone {
		if subtype != "" && subtype != "milestone" {
			return api.TaskOptions{}, fmt.Errorf("--milestone conflicts with --subtype %s", subtype)
		}
		subtype = "milestone"
	}
	if subtype != "" && !taskSubtypes[subtype] {
		return api.TaskOptions{}, fmt.Errorf("--subtype must be default_task, milestone or approval, got %q", subtype)
	}
	if due != "" {
		due = toDate(due)
	}
	return api.TaskOptions{Assignee: assignee, Subtype: subtype, DueOn: due}, nil
}
