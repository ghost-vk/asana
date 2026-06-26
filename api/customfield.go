package api

import (
	"encoding/json"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/thash/asana/utils"
)

type CustomFieldDef struct {
	Gid         string `json:"gid"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	EnumOptions []Base `json:"enum_options"`
}

func ProjectFields(project string) []CustomFieldDef {
	params := url.Values{}
	params.Add("opt_fields", "custom_field.gid,custom_field.name,custom_field.type,custom_field.enum_options.gid,custom_field.enum_options.name")
	var resp struct {
		Data []struct {
			CustomField CustomFieldDef `json:"custom_field"`
		} `json:"data"`
	}
	err := json.Unmarshal(Get("/api/1.0/projects/"+project+"/custom_field_settings", params), &resp)
	utils.Check(err)
	out := make([]CustomFieldDef, len(resp.Data))
	for i, d := range resp.Data {
		out[i] = d.CustomField
	}
	return out
}

func CustomField(fieldGid string) CustomFieldDef {
	params := url.Values{}
	params.Add("opt_fields", "name,type,enum_options.gid,enum_options.name")
	var resp struct {
		Data CustomFieldDef `json:"data"`
	}
	err := json.Unmarshal(Get("/api/1.0/custom_fields/"+fieldGid, params), &resp)
	utils.Check(err)
	return resp.Data
}

func SetCustomField(taskGid, fieldGid, value string) {
	if value != "null" {
		// ponytail: one GET to learn the field type, which also lets us accept enum
		// option NAMES, not just gids. multi_enum not handled — out of scope.
		if f := CustomField(fieldGid); f.Type == "enum" {
			value = resolveEnum(f, value)
		}
	}
	body := `{"data":{"custom_fields":{"` + fieldGid + `":` + jsonVal(value) + `}}}`
	Put("/tasks/"+taskGid, body)
}

// resolveEnum maps an enum option gid or case-insensitive name to its option gid.
func resolveEnum(f CustomFieldDef, v string) string {
	for _, o := range f.EnumOptions {
		if o.Gid == v || strings.EqualFold(o.Name, v) {
			return o.Gid
		}
	}
	log.Fatalf("fatal: %q is not an option of %q (run: asana fields -p <project>)", v, f.Name)
	return ""
}

// jsonVal: quote everything; Asana coerces "7"→number, enum needs the gid quoted. "null" clears.
// ponytail: a text field literally set to the word "null" isn't reachable — acceptable.
func jsonVal(v string) string {
	if v == "null" {
		return "null"
	}
	return strconv.Quote(v)
}
