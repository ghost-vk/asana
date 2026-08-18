package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNotesField(t *testing.T) {
	if got := NotesField(false); got != "notes" {
		t.Fatalf("NotesField(false) = %q, want \"notes\"", got)
	}
	if got := NotesField(true); got != "html_notes" {
		t.Fatalf("NotesField(true) = %q, want \"html_notes\"", got)
	}
}

func TestCommentFieldName(t *testing.T) {
	if got := CommentField(false); got != "text" {
		t.Fatalf("CommentField(false) = %q, want \"text\"", got)
	}
	if got := CommentField(true); got != "html_text" {
		t.Fatalf("CommentField(true) = %q, want \"html_text\"", got)
	}
}

func TestUpdatePayloadEncodesJSON(t *testing.T) {
	// Newlines, quotes and non-ASCII must survive as valid JSON: html_notes
	// bodies are whole documents, not the short strings strconv.Quote handled.
	payload := updatePayload("html_notes", "<body>a\n\"b\"\tявно</body>")
	var decoded struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("payload is not valid JSON: %v (%s)", err, payload)
	}
	if got, want := decoded.Data["html_notes"], "<body>a\n\"b\"\tявно</body>"; got != want {
		t.Fatalf("round-trip = %q, want %q", got, want)
	}
}

func TestCreateTaskPayload(t *testing.T) {
	cases := []struct {
		name          string
		html          bool
		wantField     string
		unwantedField string
	}{
		{"plain notes", false, "notes", "html_notes"},
		{"rich notes", true, "html_notes", "notes"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := createTaskPayload("name", "111", "<body>x</body>", tc.html)
			var decoded struct {
				Data map[string]interface{} `json:"data"`
			}
			if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
				t.Fatalf("payload is not valid JSON: %v (%s)", err, payload)
			}
			if _, ok := decoded.Data[tc.wantField]; !ok {
				t.Fatalf("payload %s lacks %q", payload, tc.wantField)
			}
			// Asana rejects a request carrying both notes and html_notes.
			if _, ok := decoded.Data[tc.unwantedField]; ok {
				t.Fatalf("payload %s must not carry %q", payload, tc.unwantedField)
			}
		})
	}
}

func TestCreateTaskPayloadOmitsEmptyBody(t *testing.T) {
	payload := createTaskPayload("name", "111", "", false)
	if strings.Contains(payload, "notes") {
		t.Fatalf("payload %s should omit an empty body", payload)
	}
}

func TestCommentPayload(t *testing.T) {
	payload := commentPayload("a\n\"b\"", true)
	var decoded struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("payload is not valid JSON: %v (%s)", err, payload)
	}
	if got, want := decoded.Data["html_text"], "a\n\"b\""; got != want {
		t.Fatalf("html_text = %q, want %q", got, want)
	}
	if _, ok := decoded.Data["text"]; ok {
		t.Fatalf("payload %s must not carry both text and html_text", payload)
	}
}
