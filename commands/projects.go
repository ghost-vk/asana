package commands

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func Projects(c *cli.Context) {
	var projects []api.Base
	if q := c.Args().First(); q != "" {
		projects = api.SearchProjects(q)
	} else {
		params := url.Values{}
		params.Add("limit", strconv.Itoa(c.Int("limit")))
		projects = api.Projects(params)
	}
	for i, p := range projects {
		fmt.Printf("%2d %s %s\n", i, p.Gid, p.Name)
	}
}
