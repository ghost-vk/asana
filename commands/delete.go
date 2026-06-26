package commands

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func DeleteTask(c *cli.Context) {
	id := c.Args().First()
	if id == "" {
		log.Fatal("fatal: task gid is required")
	}
	api.DeleteTask(id)
	fmt.Printf("deleted %s\n", id)
}
