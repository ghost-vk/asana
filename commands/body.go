package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func Body(c *cli.Context) {
	taskId := api.FindTaskId(c.Args().Get(0), false)
	text := c.Args().Get(1)
	api.Update(taskId, "notes", text)
	fmt.Printf("updated body on %s\n", taskId)
}
