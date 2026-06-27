package commands

import (
	"testing"

	"github.com/ghost-vk/asana/api"
)

func TestSourceProjectForMove(t *testing.T) {
	if _, err := sourceProjectForMove(api.Task_t{}); err == nil {
		t.Fatal("move without source project should fail")
	}
	if got, err := sourceProjectForMove(api.Task_t{Projects: []api.Base{{Gid: "123"}}}); err != nil || got != "123" {
		t.Fatalf("move source = %q, %v; want 123 nil", got, err)
	}
	if _, err := sourceProjectForMove(api.Task_t{Projects: []api.Base{{Gid: "123"}, {Gid: "456"}}}); err == nil {
		t.Fatal("move with multiple source projects should fail")
	}
}
