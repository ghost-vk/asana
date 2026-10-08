package commands

import (
	"strings"
	"testing"

	"github.com/ghost-vk/asana/api"
)

func TestParseBatchAcceptsArrayAndJSONLines(t *testing.T) {
	arr, err := parseBatch([]byte(`[{"task":"1","due":"2026-11-03"},{"task":"2","completed":true}]`))
	if err != nil || len(arr) != 2 {
		t.Fatalf("array: ops=%d err=%v", len(arr), err)
	}
	lines, err := parseBatch([]byte("{\"task\":\"1\",\"due\":null}\n\n{\"task\":\"2\",\"section\":\"9\"}\n"))
	if err != nil || len(lines) != 2 || lines[1].Section != "9" {
		t.Fatalf("json lines: ops=%+v err=%v", lines, err)
	}
}

func TestParseBatchRejectsBadInput(t *testing.T) {
	for _, in := range []string{"", "[]", `{"task":"1","dew":"2026-11-03"}`, `{"task":`} {
		if _, err := parseBatch([]byte(in)); err == nil {
			t.Fatalf("parseBatch(%q) should fail", in)
		}
	}
}

func fakeFields(gid string) api.CustomFieldDef {
	if gid == "enum" {
		return api.CustomFieldDef{Gid: gid, Name: "Stage", Type: "enum", EnumOptions: []api.Base{{Gid: "opt-1", Name: "Ready"}}}
	}
	return api.CustomFieldDef{Gid: gid, Name: "Note", Type: "text"}
}

func TestPlanOpBuildsOneUpdate(t *testing.T) {
	ops, err := parseBatch([]byte(`{"task":"1","name":"N","due":null,"assignee":"a@b.c","completed":false,` +
		`"fields":{"enum":"ready","text":"hi","num":7,"gone":null},"project":"P","section":"S"}`))
	if err != nil {
		t.Fatal(err)
	}
	s, err := planOp(ops[0], fakeFields)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := s.update["due_on"]; !ok || v != nil {
		t.Fatalf("due null must clear due_on, got %#v", s.update["due_on"])
	}
	if s.update["assignee"] != "a@b.c" || s.update["completed"] != false || s.update["name"] != "N" {
		t.Fatalf("update = %#v", s.update)
	}
	cf := s.update["custom_fields"].(map[string]interface{})
	if cf["enum"] != "opt-1" || cf["text"] != "hi" || cf["num"] != float64(7) || cf["gone"] != nil {
		t.Fatalf("custom_fields = %#v", cf)
	}
	if s.project != "P" || s.section != "S" {
		t.Fatalf("move = %s/%s", s.project, s.section)
	}
	if got := s.describe(); !strings.HasSuffix(got, "move->P/S") || !strings.Contains(got, "due_on=null") {
		t.Fatalf("describe = %q", got)
	}
}

func TestPlanOpRejects(t *testing.T) {
	cases := map[string]string{
		"no task":       `{"due":"2026-11-03"}`,
		"bad due":       `{"task":"1","due":"11/03"}`,
		"unknown enum":  `{"task":"1","fields":{"enum":"Shipped"}}`,
		"nothing to do": `{"task":"1"}`,
		"copy alone":    `{"task":"1","copy":true}`,
		"bool field":    `{"task":"1","fields":{"text":true}}`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			ops, err := parseBatch([]byte(in))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := planOp(ops[0], fakeFields); err == nil {
				t.Fatalf("planOp(%s) should fail", in)
			}
		})
	}
}
