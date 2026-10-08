package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestCreateTaskSendsAssignee(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got, want := req.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := req.URL.Path, "/api/1.0/tasks"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		var payload struct {
			Data map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("request body is invalid JSON: %v", err)
		}
		if got, want := payload.Data["assignee"], "person@example.com"; got != want {
			t.Fatalf("assignee = %#v, want %#v", got, want)
		}
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader(`{"data":{"gid":"123","name":"Ship it"}}`)),
			Header:     make(http.Header),
		}, nil
	})

	task := CreateTaskWithAssignee("Ship it", "456", "", "", false, "person@example.com")
	if got, want := task.Gid, "123"; got != want {
		t.Fatalf("created task gid = %q, want %q", got, want)
	}
}

func TestAssignTaskUpdatesAssignee(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got, want := req.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := req.URL.Path, "/api/1.0/tasks/1234567890"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		var payload struct {
			Data map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("request body is invalid JSON: %v", err)
		}
		if got, want := payload.Data["assignee"], "9876543210"; got != want {
			t.Fatalf("assignee = %#v, want %#v", got, want)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"data":{"gid":"1234567890","name":"Ship it","assignee":{"gid":"9876543210","name":"Ada"}}}`,
			)),
			Header: make(http.Header),
		}, nil
	})

	task := AssignTask("1234567890", "9876543210")
	if got, want := task.Assignee.Name, "Ada"; got != want {
		t.Fatalf("assigned user = %q, want %q", got, want)
	}
}

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

func TestTasksWithCompletedReadsEveryPage(t *testing.T) {
	restore := fetchTasksPage
	defer func() { fetchTasksPage = restore }()

	pages := map[string]struct {
		size int
		next string
	}{"": {100, "p2"}, "p2": {100, "p3"}, "p3": {50, ""}}
	var completedSince []string
	fetchTasksPage = func(params url.Values) []byte {
		completedSince = append(completedSince, params.Get("completed_since"))
		if params.Get("limit") != "100" {
			t.Fatalf("unbounded listing should request full pages, got limit %q", params.Get("limit"))
		}
		page, ok := pages[params.Get("offset")]
		if !ok {
			t.Fatalf("unexpected offset %q", params.Get("offset"))
		}
		tasks := make([]Task_t, page.size)
		for i := range tasks {
			tasks[i] = Task_t{Gid: fmt.Sprintf("%s-%d", params.Get("offset"), i), Completed: i%2 == 0}
		}
		return taskListResponse(tasks, page.next)
	}

	params := url.Values{}
	params.Set("limit", "0")
	params.Set("project", "7")
	tasks := Tasks(params, true, false)
	if got, want := len(tasks), 250; got != want {
		t.Fatalf("tasks len = %d, want %d", got, want)
	}
	for _, cs := range completedSince {
		if cs != "" {
			t.Fatalf("completed listing must not send completed_since=%q", cs)
		}
	}
}

func TestTasksOpenOnlyFiltersServerSide(t *testing.T) {
	restore := fetchTasksPage
	defer func() { fetchTasksPage = restore }()

	fetchTasksPage = func(params url.Values) []byte {
		if got := params.Get("completed_since"); got != "now" {
			t.Fatalf("completed_since = %q, want now", got)
		}
		return taskListResponse([]Task_t{{Gid: "1"}, {Gid: "2", Completed: true}}, "")
	}
	params := url.Values{}
	params.Set("project", "7")
	if got := Tasks(params, false, false); len(got) != 1 || got[0].Gid != "1" {
		t.Fatalf("tasks = %+v, want only the open one", got)
	}
}

func TestSetCompletedSendsBoolean(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	for _, completed := range []bool{true, false} {
		http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPut || req.URL.Path != "/api/1.0/tasks/1234567890" {
				t.Fatalf("request = %s %s", req.Method, req.URL.Path)
			}
			body, _ := io.ReadAll(req.Body)
			var payload struct {
				Data map[string]interface{} `json:"data"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatal(err)
			}
			if got := payload.Data["completed"]; got != completed {
				t.Fatalf("completed = %#v, want bool %v", got, completed)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"data":{"gid":"1234567890","name":"Ship it"}}`)),
				Header:     make(http.Header),
			}, nil
		})
		if got := SetCompleted("1234567890", completed); got.Name != "Ship it" {
			t.Fatalf("task name = %q", got.Name)
		}
	}
}

func TestCreateTaskPayloadSubtypeAndDue(t *testing.T) {
	payload := createTaskPayload("M1", "7", "", false, TaskOptions{Subtype: "milestone", DueOn: "2026-11-03"})
	var decoded struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Data["resource_subtype"] != "milestone" || decoded.Data["due_on"] != "2026-11-03" {
		t.Fatalf("payload = %s", payload)
	}
	plain := createTaskPayload("T", "7", "", false, TaskOptions{})
	if strings.Contains(plain, "resource_subtype") || strings.Contains(plain, "due_on") {
		t.Fatalf("payload %s should omit unset subtype and due", plain)
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
