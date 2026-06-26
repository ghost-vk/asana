package commands

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func Fields(c *cli.Context) {
	project := c.String("project")
	if project == "" {
		log.Fatal("fatal: -p <project_gid> is required")
	}
	for _, f := range api.ProjectFields(project) {
		fmt.Printf("%s %s (%s)\n", f.Gid, f.Name, f.Type)
		for _, opt := range f.EnumOptions {
			fmt.Printf("  %s %s\n", opt.Gid, opt.Name)
		}
	}
}

func SetField(c *cli.Context) {
	task := c.String("task")
	field := c.String("field")
	value := c.String("value")
	if task == "" || field == "" {
		log.Fatal("fatal: -t <task_gid> and -f <field_gid> are required")
	}
	api.SetCustomField(task, field, value)
	fmt.Printf("set field %s on %s\n", field, task)
}
