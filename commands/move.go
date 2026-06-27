package commands

import (
	"errors"
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func Move(c *cli.Context) {
	targetProject := c.String("project")
	if targetProject == "" {
		log.Fatal("fatal: -p <target_project_gid> is required")
	}

	taskId := api.FindTaskId(c.Args().First(), false)
	copyOnly := c.Bool("copy")
	sourceProject := ""
	if !copyOnly {
		task := api.TaskProjects(taskId)
		var err error
		sourceProject, err = sourceProjectForMove(task)
		if err != nil {
			log.Fatal(err)
		}
	}

	api.AddProject(taskId, targetProject, c.String("section"))
	if sourceProject != "" && sourceProject != targetProject {
		api.RemoveProject(taskId, sourceProject)
	}

	if copyOnly {
		fmt.Printf("copied %s\n", taskId)
		return
	}
	fmt.Printf("moved %s\n", taskId)
}

func sourceProjectForMove(task api.Task_t) (string, error) {
	switch len(task.Projects) {
	case 0:
		return "", errors.New("fatal: task has no source project; use --copy to add it to a project")
	case 1:
		return task.Projects[0].Gid, nil
	default:
		return "", errors.New("fatal: task is in multiple projects; use --copy")
	}
}
