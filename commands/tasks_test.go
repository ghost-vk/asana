package commands

import "testing"

func TestRenderTaskShowsAssignee(t *testing.T) {
	got := renderTask(0, "123", "default_task", "Inbox", "2026-08-26", "Ada Lovelace", "Ship it")
	want := " 0 123 Inbox                [ 2026-08-26 ] [@Ada Lovelace] Ship it"
	if got != want {
		t.Fatalf("renderTask = %q, want %q", got, want)
	}
}

func TestRenderTaskWithoutAssigneeKeepsLegacyShape(t *testing.T) {
	got := renderTask(2, "123", "default_task", "Inbox", "", "", "Ship it")
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
