package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func Comments(c *cli.Context) {
	if g := c.String("gid"); g != "" {
		fmt.Println(api.Story(g))
		return
	}
	taskId := api.FindTaskId(c.Args().First(), true)
	_, stories := api.Task(taskId, true)
	i := 0
	for _, s := range stories {
		if s.Type != "comment" { // ponytail: comments only; system events hidden
			continue
		}
		fmt.Printf("%2d %s  by %s (%s)\n%s\n", i, s.Gid, s.Created_by.Name, s.Created_at, s.Text)
		i++
	}
}
