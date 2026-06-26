package commands

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/thash/asana/api"
)

func CreateTask(c *cli.Context) {
	name := c.Args().First()
	if name == "" {
		log.Fatal("fatal: task name is required")
	}
	t := api.CreateTask(name, c.String("project"), c.String("section"), c.String("body"))
	fmt.Printf("created %s %s\n", t.Gid, t.Name)
}
