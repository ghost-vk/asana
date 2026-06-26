package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
	"github.com/ghost-vk/asana/utils"
)

func Sections(c *cli.Context) {
	project := c.String("project")
	if project == "" {
		log.Fatal("fatal: -p <project_gid> is required")
	}
	file := utils.CacheFileFor("sections-" + project)
	if !c.Bool("no-cache") && !c.Bool("refresh") && !utils.Older(CacheDuration, file) {
		if printSectionsCache(file) {
			return
		}
	}
	sections := api.Sections(project)
	cacheSections(file, sections)
	for i, s := range sections {
		fmt.Printf("%2d %s %s\n", i, s.Gid, s.Name)
	}
}

func cacheSections(file string, sections []api.Base) {
	f, err := os.Create(file)
	if err != nil {
		return // ponytail: cache is best-effort, lookup still printed from API
	}
	defer f.Close()
	for _, s := range sections {
		f.WriteString(s.Gid + ":" + s.Name + "\n")
	}
}

func printSectionsCache(file string) bool {
	txt, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	i := 0
	for _, line := range strings.Split(strings.TrimRight(string(txt), "\n"), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2) // gid never contains ':', rest is name
		fmt.Printf("%2d %s %s\n", i, parts[0], parts[1])
		i++
	}
	return true
}
