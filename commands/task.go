package commands

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
)

func Task(c *cli.Context) {
	taskId := api.FindTaskId(c.Args().First(), true)
	t, stories := api.Task(taskId, c.Bool("verbose"))
	attachments := api.Attachments(taskId)

	// html_notes is absent from the default task record, so fetch it only when
	// the output will actually show it.
	richBody := c.Bool("html")
	if richBody || c.Bool("json") {
		t.HtmlNotes = api.TaskHtmlNotes(taskId)
	}

	if c.Bool("json") {
		output := map[string]interface{}{
			"task": t,
		}
		if stories != nil {
			output["stories"] = stories
		}
		if len(attachments) > 0 {
			output["attachments"] = attachments
		}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Printf("Error marshalling JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	fmt.Printf("[ %s ] %s\n", t.Due_on, t.Name)
	showAssignee(t.Assignee)

	showTags(t.Tags)
	showCustomFields(t.CustomFields)
	showAttachments(attachments)

	body := t.Notes
	if richBody {
		body = t.HtmlNotes
	}
	fmt.Printf("\n%s\n", body)

	if stories != nil {
		fmt.Printf("\n----------------------------------------\n")
		for _, s := range stories {
			fmt.Printf("%s\n", s)
		}
	}
}

func showAssignee(assignee api.Base) {
	if assignee.Name == "" && assignee.Gid == "" {
		fmt.Println("  Assignee: unassigned")
		return
	}
	if assignee.Name == "" {
		fmt.Printf("  Assignee: %s\n", assignee.Gid)
		return
	}
	if assignee.Gid == "" {
		fmt.Printf("  Assignee: %s\n", assignee.Name)
		return
	}
	fmt.Printf("  Assignee: %s (%s)\n", assignee.Name, assignee.Gid)
}

func showTags(tags []api.Base) {
	if len(tags) > 0 {
		fmt.Print("  Tags: ")
		for i, tag := range tags {
			print(tag.Name)
			if len(tags) != 1 && i != (len(tags)-1) {
				print(", ")
			}
		}
		println("")
	}
}

func showCustomFields(fields []api.CustomField_t) {
	if len(fields) > 0 {
		fmt.Println("\n  Custom Fields:")
		for _, field := range fields {
			if field.DisplayValue != "" {
				fmt.Printf("    %s: %s\n", field.Name, field.DisplayValue)
			}
		}
	}
}

func showAttachments(attachments []api.Attachment_t) {
	if len(attachments) > 0 {
		fmt.Printf("\n  Attachments (%d):\n", len(attachments))
		for i, att := range attachments {
			fmt.Printf("    [%d] %s (GID: %s)\n", i, att.Name, att.Gid)
		}
	}
}
