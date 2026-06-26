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

type ProjectStatus_t struct {
	Text      string `json:"text"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
	Author    Base   `json:"author"`
}

type Project_t struct {
	Gid           string           `json:"gid"`
	Name          string           `json:"name"`
	Notes         string           `json:"notes"`
	Archived      bool             `json:"archived"`
	Completed     bool             `json:"completed"`
	Color         string           `json:"color"`
	CreatedAt     string           `json:"created_at"`
	ModifiedAt    string           `json:"modified_at"`
	DueOn         string           `json:"due_on"`
	StartOn       string           `json:"start_on"`
	PermalinkUrl  string           `json:"permalink_url"`
	Owner         Base             `json:"owner"`
	Team          Base             `json:"team"`
	Workspace     Base             `json:"workspace"`
	CurrentStatus *ProjectStatus_t `json:"current_status"`
}

func Project(gid string) Project_t {
	params := url.Values{}
	params.Add("opt_fields", "name,notes,archived,completed,color,created_at,modified_at,due_on,start_on,permalink_url,owner.name,team.name,workspace.name,current_status.text,current_status.color,current_status.created_at,current_status.author.name")
	var resp struct {
		Data Project_t `json:"data"`
	}
	err := json.Unmarshal(Get("/api/1.0/projects/"+gid, params), &resp)
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
