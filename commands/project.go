package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func Project(c *cli.Context) {
	gid := c.Args().First()
	if gid == "" {
		fmt.Println("usage: project <gid>")
		return
	}
	p := api.Project(gid)

	if c.Bool("json") {
		b, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		fmt.Println(string(b))
		return
	}

	fmt.Printf("%s  %s\n", p.Gid, p.Name)
	if p.PermalinkUrl != "" {
		fmt.Printf("  url:       %s\n", p.PermalinkUrl)
	}
	if p.Workspace.Name != "" {
		fmt.Printf("  workspace: %s\n", p.Workspace.Name)
	}
	if p.Team.Name != "" {
		fmt.Printf("  team:      %s\n", p.Team.Name)
	}
	if p.Owner.Name != "" {
		fmt.Printf("  owner:     %s\n", p.Owner.Name)
	}
	if p.Color != "" {
		fmt.Printf("  color:     %s\n", p.Color)
	}
	var flags []string
	if p.Archived {
		flags = append(flags, "archived")
	}
	if p.Completed {
		flags = append(flags, "completed")
	}
	if len(flags) > 0 {
		fmt.Printf("  status:    %s\n", strings.Join(flags, ", "))
	}
	if p.CreatedAt != "" {
		fmt.Printf("  created:   %s\n", p.CreatedAt)
	}
	if p.ModifiedAt != "" {
		fmt.Printf("  modified:  %s\n", p.ModifiedAt)
	}
	if p.StartOn != "" {
		fmt.Printf("  start:     %s\n", p.StartOn)
	}
	if p.DueOn != "" {
		fmt.Printf("  due:       %s\n", p.DueOn)
	}
	if p.CurrentStatus != nil {
		fmt.Printf("  [%s] %s (by %s, %s)\n", p.CurrentStatus.Color, p.CurrentStatus.Text, p.CurrentStatus.Author.Name, p.CurrentStatus.CreatedAt)
	}
	if p.Notes != "" {
		fmt.Printf("\n%s\n", p.Notes)
	}
}
