package commands

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/thash/asana/api"
	"github.com/thash/asana/utils"
)

const (
	CacheDuration = "5m"
)

func Tasks(c *cli.Context) {
	limit := c.Int("limit")
	if project := c.String("project"); project != "" {
		fromAPI(false, limit, project) // project tasks: skip cache, it's for "my tasks"
		return
	}
	if c.Bool("no-cache") {
		fromAPI(false, limit, "")
		return
	}
	if utils.Older(CacheDuration, utils.CacheFile()) || c.Bool("refresh") {
		fromAPI(true, limit, "")
		return
	}
	txt, err := ioutil.ReadFile(utils.CacheFile())
	if err != nil {
		fromAPI(true, limit, "")
		return
	}
	i := 0
	for _, line := range strings.Split(strings.TrimRight(string(txt), "\n"), "\n") {
		if line == "" {
			continue
		}
		p := strings.SplitN(line, "\t", 5) // gid \t subtype \t section \t due \t name
		if len(p) < 5 {
			continue
		}
		fmt.Println(renderTask(i, p[0], p[1], p[2], p[3], p[4]))
		i++
	}
}

func fromAPI(saveCache bool, limit int, project string) {
	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	if project != "" {
		params.Add("project", project)
	}
	tasks := api.Tasks(params, false)
	if saveCache {
		cache(tasks)
	}
	for i, t := range tasks {
		fmt.Println(renderTask(i, t.Gid, t.ResourceSubtype, t.Section(), t.Due_on, t.Name))
	}
}

// renderTask is the single source of truth for "my tasks" lines, so a fresh
// fetch and a cache read print identically.
func renderTask(i int, gid, subtype, section, due, name string) string {
	typ := ""
	if subtype != "" && subtype != "default_task" {
		typ = subtype + " "
	}
	if due != "" {
		due = "[ " + due + " ] "
	}
	return fmt.Sprintf("%2d %s %s%-20s %s%s", i, gid, typ, section, due, name)
}

func cache(tasks []api.Task_t) {
	f, _ := os.Create(utils.CacheFile())
	defer f.Close()
	for _, t := range tasks {
		// tab-delimited; FindTaskId reads the gid from field 0. Names/sections won't contain tabs.
		f.WriteString(strings.Join([]string{t.Gid, t.ResourceSubtype, t.Section(), t.Due_on, t.Name}, "\t") + "\n")
	}
}
