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
	if err != nil {
		fmt.Println("Config file isn't set.\n  ==> $ asana config")
		os.Exit(1)
	}
	conf := Conf{}
	err = yaml.Unmarshal(dat, &conf)
	utils.Check(err)
	return conf
}
