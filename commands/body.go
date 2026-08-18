package commands

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func Body(c *cli.Context) {
	taskId := api.FindTaskId(c.Args().Get(0), false)
	// An explicit empty argument still clears the body, as it always has.
	text, given := readBody(c, c.Args().Get(1), c.Args().Len() >= 2)
	if !given {
		log.Fatal(`fatal: body is required; pass it as an argument or via --file (- for stdin), or "" to clear`)
	}
	payload, html := prepareBody(text, resolveFormat(c))
	api.Update(taskId, api.NotesField(html), payload)
	fmt.Printf("updated %s on %s\n", describeFormat(html), taskId)
}
