package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
	"github.com/ghost-vk/asana/utils"
)

const (
	CacheDuration = "5m"
)

func Tasks(c *cli.Context) {
	q, err := queryFromFlags(c)
	if err != nil {
		log.Fatal("fatal: " + err.Error())
	}
	if c.Bool("json") {
		tasks := api.Tasks(q.params(), q.withCompleted, true)
		b, _ := json.MarshalIndent(tasks, "", "  ")
		fmt.Println(string(b))
		return
	}
	// The cache holds open tasks only, so a completed listing always goes to the API.
	if q.project != "" || q.withCompleted {
		fromAPI(true, q)
		return
	}
	if c.Bool("no-cache") {
		fromAPI(false, q)
		return
	}
	if utils.Older(CacheDuration, utils.CacheFile()) || c.Bool("refresh") {
		fromAPI(true, q)
		return
	}
	txt, err := ioutil.ReadFile(utils.CacheFile())
	if err != nil {
		fromAPI(true, q)
		return
	}
	i := 0
	for _, line := range strings.Split(strings.TrimRight(string(txt), "\n"), "\n") {
		if line == "" {
			continue
		}
		rendered, ok := renderCachedTask(i, line)
		if !ok {
			continue
		}
		fmt.Println(rendered)
		i++
	}
}

func renderCachedTask(i int, line string) (string, bool) {
	p := strings.SplitN(line, "\t", 7) // gid \t subtype \t section \t due \t assignee \t completed \t name
	switch len(p) {
	case 5: // v0.5 and earlier: no assignee column
		return renderTask(i, p[0], p[1], p[2], p[3], "", false, p[4]), true
	case 6: // v0.6: no completed column
		return renderTask(i, p[0], p[1], p[2], p[3], p[4], false, p[5]), true
	case 7:
		return renderTask(i, p[0], p[1], p[2], p[3], p[4], p[5] == "1", p[6]), true
	}
	return "", false
}

// taskQuery is what `asana ts` asks Asana for.
type taskQuery struct {
	project       string
	limit         int    // 0 fetches every page
	withCompleted bool   // include completed tasks
	since         string // completed_since; empty with withCompleted means all of them
}

func queryFromFlags(c *cli.Context) (taskQuery, error) {
	q := taskQuery{
		project:       c.String("project"),
		limit:         c.Int("limit"),
		withCompleted: c.Bool("completed") || c.IsSet("since"),
		since:         c.String("since"),
	}
	// A project listing is read in full unless -l caps it: stopping at the
	// first page hides tasks, and a hidden task reads as a missing one.
	if q.project != "" && !c.IsSet("limit") {
		q.limit = 0
	}
	if q.since != "" {
		if _, err := time.Parse("2006-01-02", q.since); err != nil {
			return q, fmt.Errorf("--since must be YYYY-MM-DD, got %q", q.since)
		}
	}
	return q, nil
}

func (q taskQuery) params() url.Values {
	params := url.Values{}
	params.Add("limit", strconv.Itoa(q.limit))
	if q.project != "" {
		params.Add("project", q.project)
	}
	if q.since != "" {
		params.Set("completed_since", q.since)
	}
	return params
}

func fromAPI(saveCache bool, q taskQuery) {
	tasks := api.Tasks(q.params(), q.withCompleted, false)
	if saveCache {
		cache(tasks)
	}
	for i, t := range tasks {
		fmt.Println(renderTask(i, t.Gid, t.ResourceSubtype, t.Section(), t.Due_on, t.Assignee.Name, t.Completed, t.Name))
	}
}

// renderTask is the single source of truth for "my tasks" lines, so a fresh
// fetch and a cache read print identically.
func renderTask(i int, gid, subtype, section, due, assignee string, completed bool, name string) string {
	typ := ""
	if subtype != "" && subtype != "default_task" {
		typ = subtype + " "
	}
	if due != "" {
		due = "[ " + due + " ] "
	}
	if assignee != "" {
		assignee = "[@" + assignee + "] "
	}
	if completed {
		name = "✓ " + name
	}
	return fmt.Sprintf("%2d %s %s%-20s %s%s%s", i, gid, typ, section, due, assignee, name)
}

func cache(tasks []api.Task_t) {
	f, _ := os.Create(utils.CacheFile())
	defer f.Close()
	for _, t := range tasks {
		// tab-delimited; FindTaskId reads the gid from field 0. Names/sections won't contain tabs.
		completed := ""
		if t.Completed {
			completed = "1"
		}
		f.WriteString(strings.Join([]string{t.Gid, t.ResourceSubtype, t.Section(), t.Due_on, t.Assignee.Name, completed, t.Name}, "\t") + "\n")
	}
}
