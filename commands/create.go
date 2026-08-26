package commands

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func CreateTask(c *cli.Context) {
	name := c.Args().First()
	if name == "" {
		log.Fatal("fatal: task name is required")
	}
	var payload string
	var html bool
	if body, _ := readBody(c, c.String("body"), c.IsSet("body")); body != "" {
		payload, html = prepareBody(body, resolveFormat(c))
	}
	t := api.CreateTaskWithAssignee(name, c.String("project"), c.String("section"), payload, html, c.String("assignee"))
	fmt.Printf("created %s %s\n", t.Gid, t.Name)
}
