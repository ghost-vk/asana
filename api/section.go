package api

import (
	"encoding/json"
	"net/url"

	"github.com/thash/asana/utils"
)

func Sections(project string) []Base {
	params := url.Values{}
	params.Add("opt_fields", "name")
	var resp struct {
		Data []Base `json:"data"`
	}
	err := json.Unmarshal(Get("/api/1.0/projects/"+project+"/sections", params), &resp)
	utils.Check(err)
	return resp.Data
}
