package api

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/ghost-vk/asana/config"
	"github.com/ghost-vk/asana/utils"
)

func Projects(params url.Values) []Base {
	params.Add("workspace", strconv.Itoa(config.Load().Workspace))
	params.Add("opt_fields", "name")
	if params.Get("limit") == "" {
		params.Set("limit", "100") // ponytail: Asana rejects unpaginated lists as "too large"
	}
	var resp struct {
		Data []Base `json:"data"`
	}
	err := json.Unmarshal(Get("/api/1.0/projects", params), &resp)
	utils.Check(err)
	return resp.Data
}

func SearchProjects(query string) []Base {
	ws := strconv.Itoa(config.Load().Workspace)
	params := url.Values{}
	params.Add("resource_type", "project")
	params.Add("query", query)
	params.Add("opt_fields", "name")
	params.Add("count", "100") // typeahead max
	var resp struct {
		Data []Base `json:"data"`
	}
	err := json.Unmarshal(Get("/api/1.0/workspaces/"+ws+"/typeahead", params), &resp)
	utils.Check(err)
	return resp.Data
}
