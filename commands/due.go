package commands

import (
	"fmt"
	"regexp"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

const (
	DateRegexp = `^\d{4}-\d{2}-\d{2}$`
)

func DueOn(c *cli.Context) {
	taskId := api.FindTaskId(c.Args().First(), true)
	task := api.Update(taskId, "due_on", toDate(c.Args().Get(1)))
	fmt.Println("set due on [ " + task.Due_on + " ] :" + task.Name)
}

func toDate(str string) string {
	switch {
	case regexp.MustCompile(DateRegexp).MatchString(str):
		return str
	case str == "today":
		return time.Now().Format("2006-01-02")
	case str == "tomorrow":
		d, _ := time.ParseDuration("24h")
		return time.Now().Add(d).Format("2006-01-02")
	default:
		// Asana API should return err.
		return str
	}
}
