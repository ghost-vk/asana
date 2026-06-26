package api

import (
	"encoding/json"

	"github.com/thash/asana/utils"
)

func Story(storyGid string) Story_t {
	var out map[string]Story_t
	err := json.Unmarshal(Get("/api/1.0/stories/"+storyGid, nil), &out)
	utils.Check(err)
	return out["data"]
}
