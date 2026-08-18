package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
	"github.com/ghost-vk/asana/config"
	"github.com/ghost-vk/asana/utils"
)

func Comment(c *cli.Context) {
	taskId := api.FindTaskId(c.Args().First(), false)
	format := resolveFormat(c)

	// --file/stdin skips the editor entirely, which also skips the '#' comment
	// stripping below - markdown headings would not survive it.
	if body, given := readBody(c, "", false); given && body != "" {
		postComment(taskId, body, format)
		return
	}

	task, stories := api.Task(taskId, true)

	tmpFile := os.TempDir() + "/asana_comment.txt"
	f, err := os.Create(tmpFile)
	utils.Check(err)
	defer f.Close()

	err = template(f, task, stories)
	utils.Check(err)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = config.Load().Editor
	}
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, tmpFile)
	cmd.Stdin, cmd.Stdout = os.Stdin, os.Stdout
	err = cmd.Run()

	txt, err := ioutil.ReadFile(tmpFile)

	utils.Check(err)

	body := trim(string(txt), format)
	if body == "" {
		fmt.Println("Aborting comment due to empty content.")
		return
	}
	fmt.Println("Commented on Task: \"" + task.Name + "\"\n")
	postComment(taskId, body, format)
}

func postComment(taskId, body string, format bodyFormat) {
	payload, html := prepareBody(body, format)
	fmt.Println(api.CommentTo(taskId, payload, html))
}

func template(f *os.File, task api.Task_t, stories []api.Story_t) error {
	var err error
	_, err = f.WriteString("\n\n\n")
	_, err = f.WriteString("# =================================== \n")
	_, err = f.WriteString("# " + task.Name + "\n#\n")
	_, err = f.WriteString(commentOut(task.Notes) + "\n#\n")
	_, err = f.WriteString("\n# ----------------------------------- \n")
	for _, s := range stories {
		_, err = f.WriteString(commentOut(fmt.Sprintf("%s", s)) + "\n")
	}
	return err
}

func commentOut(txt string) string {
	return strings.Replace("# "+txt, "\n", "\n# ", -1)
}

// templateStart marks the first line of the read-only block the editor template
// appends. In markdown mode only that block is cut, so that '#' headings the
// user typed above it survive; plain mode keeps the historical behaviour of
// dropping every '#' line.
var templateStart = regexp.MustCompile(`(?m)^# ={3,}\s*$`)

func trim(txt string, format bodyFormat) string {
	var result string
	if format == formatPlain {
		result = regexp.MustCompile("#.*\n").ReplaceAllString(txt, "") // Remove comments
	} else if loc := templateStart.FindStringIndex(txt); loc != nil {
		result = txt[:loc[0]]
	} else {
		result = txt
	}
	return strings.Trim(result, "\n")
}
