package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
	"github.com/ghost-vk/asana/config"
	"github.com/ghost-vk/asana/utils"
)

func Config(c *cli.Context) {
	path := utils.Home() + "/.asana.yml"

	// Token priority: CLI argument > existing config > interactive prompt.
	token := c.Args().First()
	if token == "" {
		if _, err := ioutil.ReadFile(path); err == nil {
			token = config.Load().Personal_access_token
		}
	}
	if token == "" {
		println("visit: http://app.asana.com/-/account_api")
		println("  Settings > Apps > Manage Developer Apps > Personal Access Tokens")
		println("  + Create New Personal Access Token")
		print("\npaste your Personal Access Token: ")
		fmt.Scanf("%s", &token)
	}
	if token == "" {
		log.Fatal("fatal: no Personal Access Token provided")
	}

	// Validate the token against the API before writing anything: Me() fatals
	// on a bad token (e.g. 401) so an invalid credential never gets persisted.
	config.SetToken(token)
	ws := api.Me().Workspaces
	index := 0

	if len(ws) > 1 {
		fmt.Println("\n" + strconv.Itoa(len(ws)) + " workspaces found.")
		for i, w := range ws {
			fmt.Printf("[%d] %s %s\n", i, w.Gid, w.Name)
		}
		index = utils.EndlessSelect(len(ws)-1, index)
	}

	f, _ := os.Create(path)
	defer f.Close()
	f.WriteString("personal_access_token: " + token + "\n")
	f.WriteString("workspace: " + ws[index].Gid + "\n")
}
