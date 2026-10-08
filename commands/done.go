package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func Done(c *cli.Context) {
	task := api.SetCompleted(api.FindTaskId(c.Args().First(), false), true)
	fmt.Println("DONE! : " + task.Name)
}

func Undone(c *cli.Context) {
	task := api.SetCompleted(api.FindTaskId(c.Args().First(), false), false)
	fmt.Println("REOPENED : " + task.Name)
}
