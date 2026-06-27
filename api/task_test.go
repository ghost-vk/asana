package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"testing"
)

func TestTasksPaginateByLimit(t *testing.T) {
	restore := fetchTasksPage
	defer func() { fetchTasksPage = restore }()

	calls := []string{}
	fetchTasksPage = func(params url.Values) []byte {
		call := params.Get("limit") + ":" + params.Get("offset")
		calls = append(calls, call)

		switch params.Get("offset") {
		case "":
			if params.Get("limit") != "100" {
				t.Fatalf("first request should use page size 100, got %q", params.Get("limit"))
			}
			tasks := make([]Task_t, 0, 100)
			for i := 1; i <= 100; i++ {
				tasks = append(tasks, Task_t{
					Gid:    fmt.Sprintf("task-%02d", i),
					Due_on: fmt.Sprintf("2026-01-%02d", (i%30)+1),
				})
			}
			return taskListResponse(tasks, "cursor")
		case "cursor":
			if params.Get("limit") != "20" {
				t.Fatalf("second request should use remaining page size 20, got %q", params.Get("limit"))
			}
			tasks := make([]Task_t, 0, 20)
			for i := 101; i <= 120; i++ {
				tasks = append(tasks, Task_t{
					Gid:    fmt.Sprintf("task-%02d", i),
					Due_on: "",
				})
			}
			return taskListResponse(tasks, "")
		default:
			t.Fatalf("unexpected offset %q", params.Get("offset"))
			return []byte("{}")
		}
	}

	params := url.Values{}
	params.Set("limit", "120")
	tasks := Tasks(params, false, true)
	if got, want := len(tasks), 120; got != want {
		t.Fatalf("tasks len = %d, want %d", got, want)
	}
	if got := calls; !reflect.DeepEqual(got, []string{"100:", "20:cursor"}) {
		t.Fatalf("request sequence = %v, want %v", got, []string{"100:", "20:cursor"})
	}
	for i := 1; i < 100; i++ {
		if tasks[i-1].Due_on > tasks[i].Due_on {
			t.Fatalf("due-date sorting broken at %d: %q > %q", i, tasks[i-1].Due_on, tasks[i].Due_on)
		}
	}
	for i := 100; i < 120; i++ {
		if tasks[i].Due_on != "" {
			t.Fatalf("undued tasks must be moved to the end, got due %q at %d", tasks[i].Due_on, i)
		}
	}
}

func TestMoveProjectPayloads(t *testing.T) {
	if got, want := addProjectPayload("123", "456"), `{"data":{"project":"123","section":"456"}}`; got != want {
		t.Fatalf("addProjectPayload = %s, want %s", got, want)
	}
	if got, want := addProjectPayload("123", ""), `{"data":{"project":"123"}}`; got != want {
		t.Fatalf("addProjectPayload without section = %s, want %s", got, want)
	}
	if got, want := removeProjectPayload("123"), `{"data":{"project":"123"}}`; got != want {
		t.Fatalf("removeProjectPayload = %s, want %s", got, want)
	}
}

func taskListResponse(data []Task_t, nextOffset string) []byte {
	response := struct {
		Data     []Task_t `json:"data"`
		NextPage *struct {
			Offset string `json:"offset"`
		} `json:"next_page"`
	}{
		Data: data,
	}
	if nextOffset != "" {
		response.NextPage = &struct {
			Offset string `json:"offset"`
		}{
			Offset: nextOffset,
		}
	}
	b, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	return b
}
