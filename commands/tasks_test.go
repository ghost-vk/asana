package commands

import "testing"

func TestRenderTaskShowsAssignee(t *testing.T) {
	got := renderTask(0, "123", "default_task", "Inbox", "2026-08-26", "Ada Lovelace", false, "Ship it")
	want := " 0 123 Inbox                [ 2026-08-26 ] [@Ada Lovelace] Ship it"
	if got != want {
		t.Fatalf("renderTask = %q, want %q", got, want)
	}
}

func TestRenderTaskWithoutAssigneeKeepsLegacyShape(t *testing.T) {
	got := renderTask(2, "123", "default_task", "Inbox", "", "", false, "Ship it")
	want := " 2 123 Inbox                Ship it"
	if got != want {
		t.Fatalf("renderTask = %q, want %q", got, want)
	}
}

func TestRenderCachedTaskReadsLegacyAndCurrentShapes(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "v0.5 cache without assignee",
			line: "123\tdefault_task\tInbox\t\tShip it",
			want: " 1 123 Inbox                Ship it",
		},
		{
			name: "v0.6 cache with assignee",
			line: "123\tdefault_task\tInbox\t\tAda Lovelace\tShip it",
			want: " 1 123 Inbox                [@Ada Lovelace] Ship it",
		},
		{
			name: "v0.7 cache, open task",
			line: "123\tdefault_task\tInbox\t\tAda Lovelace\t\tShip it",
			want: " 1 123 Inbox                [@Ada Lovelace] Ship it",
		},
		{
			name: "v0.7 cache, completed milestone",
			line: "123\tmilestone\tInbox\t2026-11-03\t\t1\tShip it",
			want: " 1 123 milestone Inbox                [ 2026-11-03 ] ✓ Ship it",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := renderCachedTask(1, tc.line)
			if !ok {
				t.Fatal("cache line was rejected")
			}
			if got != tc.want {
				t.Fatalf("renderCachedTask = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRenderTaskMarksCompleted(t *testing.T) {
	got := renderTask(0, "123", "default_task", "Done", "", "", true, "Ship it")
	want := " 0 123 Done                 ✓ Ship it"
	if got != want {
		t.Fatalf("renderTask = %q, want %q", got, want)
	}
}

func TestTaskOptions(t *testing.T) {
	opts, err := taskOptions("a@b.c", "", true, "2026-11-03")
	if err != nil {
		t.Fatal(err)
	}
	if opts.Subtype != "milestone" || opts.DueOn != "2026-11-03" || opts.Assignee != "a@b.c" {
		t.Fatalf("taskOptions = %+v", opts)
	}
	if opts, _ := taskOptions("", "approval", false, ""); opts.Subtype != "approval" {
		t.Fatalf("subtype = %q, want approval", opts.Subtype)
	}
	if _, err := taskOptions("", "", false, "11/03"); err == nil {
		t.Fatal("malformed --due must be rejected")
	}
	if _, err := taskOptions("", "epic", false, ""); err == nil {
		t.Fatal("unknown subtype must be rejected")
	}
	if _, err := taskOptions("", "approval", true, ""); err == nil {
		t.Fatal("--milestone with another --subtype must be rejected")
	}
}

func TestTaskQueryParams(t *testing.T) {
	q := taskQuery{project: "7", limit: 0, withCompleted: true, since: "2026-01-01"}
	p := q.params()
	if p.Get("project") != "7" || p.Get("limit") != "0" || p.Get("completed_since") != "2026-01-01" {
		t.Fatalf("params = %v", p)
	}
	if got := (taskQuery{limit: 100}).params().Get("completed_since"); got != "" {
		t.Fatalf("open-only query must leave completed_since to api.Tasks, got %q", got)
	}
}
