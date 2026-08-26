package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
	"github.com/ghost-vk/asana/utils"
)

// Assign sets a task's assignee by email or user GID.
func Assign(c *cli.Context) {
	assignee := c.Args().Get(1)
	if assignee == "" {
		log.Fatal("fatal: assignee email or user gid is required")
	}
	taskId := api.FindTaskId(c.Args().First(), false)
	api.AssignTask(taskId, assignee)
	// The cache now contains the old assignee; force the next listing to refresh.
	_ = os.Remove(utils.CacheFile())
	fmt.Printf("assigned %s to %s\n", taskId, assignee)
}
