package config

import (
	"testing"
)

// SetToken must make Load() return the injected token regardless of any
// on-disk config, so `asana config` can validate a token before persisting it.
func TestSetTokenOverridesLoad(t *testing.T) {
	defer SetToken("")

	SetToken("2/override/token:secret")
	if got := Load().Personal_access_token; got != "2/override/token:secret" {
		t.Errorf("Load() token = %q, want override token", got)
	}
}
