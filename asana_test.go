package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	expects := []string{"config", "workspaces", "tasks", "projects", "project", "sections",
		"create", "move", "task", "comment", "comments", "done", "due", "body",
		"browse", "fields", "set-field", "delete", "download"}
	cmds := defs()
	if len(cmds) != len(expects) {
		t.Error("commands mismatch")
	}
	for _, cmd := range cmds {
		if !include(cmd.Name, expects) {
			t.Error("commands mismatch")
		}
	}
}

func include(target string, list []string) bool {
	for _, item := range list {
		if target == item {
			return true
		}
	}
	return false
}
