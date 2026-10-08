package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/api"
	"github.com/ghost-vk/asana/utils"
)

// batchOp is one line of `asana batch` input: every change for one task.
// due, assignee and field values take null to clear.
type batchOp struct {
	Task      string                     `json:"task"`
	Name      string                     `json:"name"`
	Due       json.RawMessage            `json:"due"`
	Assignee  json.RawMessage            `json:"assignee"`
	Completed *bool                      `json:"completed"`
	Fields    map[string]json.RawMessage `json:"fields"`
	Project   string                     `json:"project"`
	Section   string                     `json:"section"`
	Copy      bool                       `json:"copy"`
}

// batchStep is a validated op: one PUT for the task fields, then an optional move.
type batchStep struct {
	task    string
	update  map[string]interface{}
	project string
	section string
	copy    bool
}

func Batch(c *cli.Context) {
	src := c.String("file")
	if src == "" {
		log.Fatal("fatal: -f <file> is required (- for stdin)")
	}
	var in io.Reader = os.Stdin
	if src != "-" {
		f, err := os.Open(src)
		utils.Check(err)
		defer f.Close()
		in = f
	}
	raw, err := io.ReadAll(in)
	utils.Check(err)
	ops, err := parseBatch(raw)
	if err != nil {
		log.Fatal("fatal: " + err.Error())
	}

	// Everything is checked before the first write, so a typo in op 40 does
	// not leave 39 tasks changed and the rest untouched.
	fields := map[string]api.CustomFieldDef{}
	lookup := func(gid string) api.CustomFieldDef {
		if f, ok := fields[gid]; ok {
			return f
		}
		fields[gid] = api.CustomField(gid)
		return fields[gid]
	}
	steps := make([]batchStep, len(ops))
	for i, op := range ops {
		if steps[i], err = planOp(op, lookup); err != nil {
			log.Fatalf("fatal: op %d (task %s): %v", i+1, op.Task, err)
		}
	}

	if c.Bool("dry-run") {
		for _, s := range steps {
			fmt.Println("plan " + s.describe())
		}
		return
	}

	failed := 0
	for _, s := range steps {
		if err := runStep(s); err != nil {
			failed++
			fmt.Printf("fail %s: %s\n", s.task, strings.ReplaceAll(err.Error(), "\n", " "))
			continue
		}
		fmt.Println("ok   " + s.describe())
	}
	// Sections, due dates and assignees in the index cache are stale now.
	_ = os.Remove(utils.CacheFile())
	if failed > 0 {
		log.Fatalf("fatal: %d of %d tasks failed", failed, len(steps))
	}
}

// parseBatch reads a JSON array of ops or one op per line (JSON lines).
func parseBatch(raw []byte) ([]batchOp, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("batch input is empty")
	}
	var ops []batchOp
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.DisallowUnknownFields() // a misspelled key must not be skipped silently
	if trimmed[0] == '[' {
		if err := dec.Decode(&ops); err != nil {
			return nil, fmt.Errorf("batch input: %v", err)
		}
		if dec.More() {
			return nil, fmt.Errorf("batch input: unexpected data after the array")
		}
	} else {
		for dec.More() {
			var op batchOp
			if err := dec.Decode(&op); err != nil {
				return nil, fmt.Errorf("batch op %d: %v", len(ops)+1, err)
			}
			ops = append(ops, op)
		}
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("batch input has no ops")
	}
	return ops, nil
}

func planOp(op batchOp, field func(gid string) api.CustomFieldDef) (batchStep, error) {
	s := batchStep{task: op.Task, update: map[string]interface{}{}, project: op.Project, section: op.Section, copy: op.Copy}
	if op.Task == "" {
		return s, fmt.Errorf(`"task" (gid) is required`)
	}
	if op.Copy && op.Project == "" {
		return s, fmt.Errorf(`"copy" needs "project"`)
	}
	if op.Name != "" {
		s.update["name"] = op.Name
	}
	if op.Due != nil {
		due, err := nullableString(op.Due, "due")
		if err != nil {
			return s, err
		}
		if due != nil {
			d := toDate(*due)
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return s, fmt.Errorf("due must be YYYY-MM-DD, today or tomorrow, got %q", *due)
			}
			s.update["due_on"] = d
		} else {
			s.update["due_on"] = nil
		}
	}
	if op.Assignee != nil {
		a, err := nullableString(op.Assignee, "assignee")
		if err != nil {
			return s, err
		}
		if a != nil {
			s.update["assignee"] = *a
		} else {
			s.update["assignee"] = nil
		}
	}
	if op.Completed != nil {
		s.update["completed"] = *op.Completed
	}
	if len(op.Fields) > 0 {
		values := map[string]interface{}{}
		for gid, rawValue := range op.Fields {
			v, err := fieldValue(gid, rawValue, field)
			if err != nil {
				return s, err
			}
			values[gid] = v
		}
		s.update["custom_fields"] = values
	}
	if len(s.update) == 0 && s.project == "" && s.section == "" {
		return s, fmt.Errorf("nothing to change")
	}
	return s, nil
}

// fieldValue turns a JSON value into what Asana takes for a custom field:
// null clears, a number passes through, a string on an enum field is matched
// against option names and gids.
func fieldValue(gid string, raw json.RawMessage, field func(gid string) api.CustomFieldDef) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("field %s: %v", gid, err)
	}
	switch val := v.(type) {
	case nil, float64:
		return val, nil
	case string:
		if f := field(gid); f.Type == "enum" {
			return api.EnumOption(f, val)
		}
		return val, nil
	}
	return nil, fmt.Errorf("field %s: value must be a string, number or null", gid)
}

func nullableString(raw json.RawMessage, key string) (*string, error) {
	var s *string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("%s must be a string or null", key)
	}
	return s, nil
}

func runStep(s batchStep) error {
	// Find the source project before writing anything, so a task that cannot
	// be moved is left untouched rather than half-applied.
	source := ""
	if s.project != "" && !s.copy {
		t, err := api.TryTaskProjects(s.task)
		if err != nil {
			return err
		}
		if source, err = sourceProjectForMove(t); err != nil {
			msg := strings.TrimPrefix(err.Error(), "fatal: ")
			return fmt.Errorf("%s", strings.ReplaceAll(msg, "--copy", `"copy": true`))
		}
	}
	if len(s.update) > 0 {
		if _, err := api.TryUpdateTask(s.task, s.update); err != nil {
			return err
		}
	}
	if err := moveStep(s, source); err != nil {
		if len(s.update) > 0 {
			return fmt.Errorf("fields updated, move failed: %v", err)
		}
		return err
	}
	return nil
}

func moveStep(s batchStep, source string) error {
	switch {
	case s.project != "":
		if err := api.TryAddProject(s.task, s.project, s.section); err != nil {
			return err
		}
		if source != "" && source != s.project {
			return api.TryRemoveProject(s.task, source)
		}
	case s.section != "":
		return api.TryAddToSection(s.section, s.task)
	}
	return nil
}

func (s batchStep) describe() string {
	keys := make([]string, 0, len(s.update))
	for k, v := range s.update {
		if k == "custom_fields" {
			k = fmt.Sprintf("fields(%d)", len(v.(map[string]interface{})))
		} else if v == nil {
			k += "=null"
		} else {
			k += "=" + fmt.Sprint(v)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	switch {
	case s.project != "" && s.copy:
		keys = append(keys, "copy->"+s.project+sectionSuffix(s.section))
	case s.project != "":
		keys = append(keys, "move->"+s.project+sectionSuffix(s.section))
	case s.section != "":
		keys = append(keys, "section->"+s.section)
	}
	return s.task + " " + strings.Join(keys, " ")
}

func sectionSuffix(section string) string {
	if section == "" {
		return ""
	}
	return "/" + section
}
