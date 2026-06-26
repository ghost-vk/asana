package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/ghost-vk/asana/config"
	"github.com/ghost-vk/asana/utils"
)

type CustomField_t struct {
	Gid          string `json:"gid"`
	Name         string `json:"name"`
	DisplayValue string `json:"display_value"`
	Type         string `json:"type"`
}

type Attachment_t struct {
	Gid          string `json:"gid"`
	Name         string `json:"name"`
	CreatedAt    string `json:"created_at"`
	DownloadUrl  string `json:"download_url"`
	ViewUrl      string `json:"view_url"`
	PermanentUrl string `json:"permanent_url"`
	Host         string `json:"host"`
}

type Membership_t struct {
	Project Base `json:"project"`
	Section Base `json:"section"`
}

type Task_t struct {
	Gid             string          `json:"gid"`
	ResourceSubtype string          `json:"resource_subtype"`
	Memberships     []Membership_t  `json:"memberships"`
	Created_at      string          `json:"created_at"`
	Modified_at     string          `json:"modified_at"`
	Name            string          `json:"name"`
	Notes           string          `json:"notes"`
	Assignee        Base            `json:"assignee"`
	Completed       bool            `json:"completed"`
	Assignee_status string          `json:"assignee_status"`
	Completed_at    string          `json:"completed_at"`
	Due_on          string          `json:"due_on"`
	Tags            []Base          `json:"tags"`
	CustomFields    []CustomField_t `json:"custom_fields"`
	Workspace       Base            `json:"workspace"`
	Parent          Base            `json:"parent"`
	Projects        []Base          `json:"projects"`
	Folloers        []Base          `json:"followers"`
}

type Story_t struct {
	Gid        string
	Text       string
	Type       string
	Created_at string
	Created_by Base
}

func (t Task_t) Section() string {
	if len(t.Memberships) > 0 {
		return t.Memberships[0].Section.Name // ponytail: first membership; match by project gid if a task spans projects
	}
	return ""
}

type ByDue []Task_t

func (a ByDue) Len() int           { return len(a) }
func (a ByDue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDue) Less(i, j int) bool { return a[i].Due_on < a[j].Due_on }

func Tasks(params url.Values, withCompleted bool, detailed bool) []Task_t {
	if params.Get("project") == "" {
		params.Add("workspace", strconv.Itoa(config.Load().Workspace))
		params.Add("assignee", "me")
	}
	optFields := "name,completed,due_on,resource_subtype,memberships.section.name"
	if detailed {
		optFields += ",assignee.name,custom_fields.name,custom_fields.display_value"
	}
	params.Add("opt_fields", optFields)
	if !withCompleted {
		params.Set("completed_since", "now") // фильтруем на стороне Asana, иначе limit съедают старые завершённые
	}
	if params.Get("limit") == "" {
		params.Set("limit", "100") // ponytail: Asana rejects unpaginated /tasks as "too large"
	}
	var tasks struct {
		Data []Task_t `json:"data"`
	}
	err := json.Unmarshal(Get("/api/1.0/tasks", params), &tasks)
	utils.Check(err)
	var tasks_without_due, tasks_with_due []Task_t
	for _, t := range tasks.Data {
		if !withCompleted && t.Completed {
			continue
		}
		if t.Due_on == "" {
			tasks_without_due = append(tasks_without_due, t)
		} else {
			tasks_with_due = append(tasks_with_due, t)
		}
	}
	sort.Sort(ByDue(tasks_with_due))
	return append(tasks_with_due, tasks_without_due...)
}

func Task(taskId string, verbose bool) (Task_t, []Story_t) {
	var (
		err     error
		t       map[string]Task_t
		ss      map[string][]Story_t
		stories []Story_t
	)
	task_chan, stories_chan := make(chan []byte), make(chan []byte)
	go func() {
		task_chan <- Get("/api/1.0/tasks/"+taskId, nil)
	}()

	stories = nil
	if verbose {
		go func() {
			stories_chan <- Get("/api/1.0/tasks/"+taskId+"/stories", nil)
		}()
		err = json.Unmarshal(<-stories_chan, &ss)
		utils.Check(err)
		stories = ss["data"]
	}

	err = json.Unmarshal(<-task_chan, &t)
	utils.Check(err)
	return t["data"], stories
}

func Attachments(taskId string) []Attachment_t {
	var attachments map[string][]Attachment_t
	err := json.Unmarshal(Get("/api/1.0/tasks/"+taskId+"/attachments", nil), &attachments)
	utils.Check(err)
	return attachments["data"]
}

func Attachment(attachmentId string) Attachment_t {
	var attachment map[string]Attachment_t
	err := json.Unmarshal(Get("/api/1.0/attachments/"+attachmentId, nil), &attachment)
	utils.Check(err)
	return attachment["data"]
}

func FindTaskId(index string, autoFirst bool) string {
	if index == "" {
		if autoFirst == false {
			log.Fatal("fatal: Task index is required.")
		} else {
			index = "0"
		}
	}
	// ponytail: GID is 16+ digits; skip cache lookup
	if len(index) >= 10 {
		if _, err := strconv.ParseUint(index, 10, 64); err == nil {
			return index
		}
	}

	var id string
	txt, err := ioutil.ReadFile(utils.CacheFile())

	if err != nil { // cache file not exist
		ind, parseErr := strconv.Atoi(index)
		utils.Check(parseErr)
		task := Tasks(url.Values{}, false, false)[ind]
		id = task.Gid
	} else {
		i := 0
		for _, line := range strings.Split(strings.TrimRight(string(txt), "\n"), "\n") {
			if line == "" {
				continue
			}
			if index == strconv.Itoa(i) {
				id = strings.SplitN(line, "\t", 2)[0] // gid is field 0
			}
			i++
		}
	}
	return id
}

func (s Story_t) String() string {
	if s.Type == "comment" {
		return fmt.Sprintf("> %s\nby %s (%s)", s.Text, s.Created_by.Name, s.Created_at)
	} else {
		return fmt.Sprintf("* %s (%s)", s.Text, s.Created_at)
	}
}

type Commented_t struct {
	Text string `json:"text"` // Define only required field.
}

func CommentTo(taskId string, comment string) string {

	respBody := Post("/tasks/"+taskId+"/stories", `{"data":{"text":"`+comment+`"}}`)

	var output map[string]Commented_t
	err := json.Unmarshal(respBody, &output)
	utils.Check(err)

	return output["data"].Text
}

func CreateTask(name, project, section, notes string) Task_t {
	data := `{"data":{"name":` + strconv.Quote(name)
	if notes != "" {
		data += `,"notes":` + strconv.Quote(notes)
	}
	if project != "" {
		data += `,"projects":["` + project + `"]`
	} else {
		data += `,"workspace":"` + strconv.Itoa(config.Load().Workspace) + `"`
	}
	data += `}}`

	var output map[string]Task_t
	err := json.Unmarshal(Post("/tasks", data), &output)
	utils.Check(err)
	t := output["data"]

	if section != "" {
		Post("/sections/"+section+"/addTask", `{"data":{"task":"`+t.Gid+`"}}`)
	}
	return t
}

func DeleteTask(taskId string) {
	Delete("/tasks/" + taskId)
}

func Update(taskId string, key string, value string) Task_t {
	respBody := Put("/tasks/"+taskId, `{"data":{"`+key+`":`+strconv.Quote(value)+`}}`)

	var output map[string]Task_t
	err := json.Unmarshal(respBody, &output)
	utils.Check(err)

	return output["data"]
}
