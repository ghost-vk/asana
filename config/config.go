package config

import (
	"github.com/ghost-vk/asana/utils"

	"fmt"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"os"
)

type Conf struct {
	Personal_access_token string
	Workspace             int
	Editor                string
}

// tokenOverride lets `asana config` validate a freshly supplied token against
// the API before it is persisted, so a bad token never lands in the config
// file. ponytail: package global, adequate for a single-shot CLI.
var tokenOverride string

// SetToken makes token the effective Personal Access Token for subsequent
// Load() calls, without writing anything to disk.
func SetToken(token string) { tokenOverride = token }

func Load() Conf {
	home := utils.Home()
	paths := []string{
		home + "/.config/asana-cli/config.yml",
		home + "/.asana.yml",
	}
	var dat []byte
	var err error
	for _, p := range paths {
		dat, err = ioutil.ReadFile(p)
		if err == nil {
			break
		}
	}
	conf := Conf{}
	if err == nil {
		utils.Check(yaml.Unmarshal(dat, &conf))
	} else if tokenOverride == "" {
		fmt.Println("Config file isn't set.\n  ==> $ asana config")
		os.Exit(1)
	}
	if tokenOverride != "" {
		conf.Personal_access_token = tokenOverride
	}
	return conf
}
