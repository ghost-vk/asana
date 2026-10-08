package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	expects := []string{"config", "workspaces", "tasks", "projects", "project", "sections",
		"create", "assign", "move", "task", "comment", "comments", "done", "undone", "due", "body",
		"browse", "fields", "set-field", "batch", "delete", "download"}
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

func TestCreateAcceptsAssigneeFlag(t *testing.T) {
	for _, cmd := range defs() {
		if cmd.Name != "create" {
			continue
		}
		for _, flag := range cmd.Flags {
			if include("assignee", flag.Names()) && include("a", flag.Names()) {
				return
			}
		}
		t.Fatal("create command lacks -a/--assignee")
	}
	t.Fatal("create command not found")
}

func include(target string, list []string) bool {
	for _, item := range list {
		if target == item {
			return true
		}
	}
	return false
}
