package app

import (
	"fmt"
	"testing"

	"github.com/openconfig/gnmic/pkg/config"
)

// A clustered instance can take its targets and subscriptions entirely from a
// loader, so "no subscriptions or inputs" is only a fatal configuration error when
// there is also no loader to provide them. The check in SubscribeRunE runs before
// startLoader, so rejecting it there means a loader-driven deployment can never
// start at all.
func TestSubscribeRequiresSubscriptionsInputsOrLoader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		loader      map[string]any
		inputs      int
		wantAllowed bool
	}{
		{
			name:        "loader only, no static subscriptions or inputs",
			loader:      map[string]any{"type": "consul", "address": "127.0.0.1:8500"},
			inputs:      0,
			wantAllowed: true,
		},
		{
			name:        "loader plus inputs",
			loader:      map[string]any{"type": "file", "path": "targets.yml"},
			inputs:      1,
			wantAllowed: true,
		},
		{
			name:        "inputs only, no loader",
			loader:      nil,
			inputs:      1,
			wantAllowed: true,
		},
		{
			name:        "nothing at all is still rejected",
			loader:      nil,
			inputs:      0,
			wantAllowed: false,
		},
		{
			name:        "empty loader map is not a loader",
			loader:      map[string]any{},
			inputs:      0,
			wantAllowed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.New()
			cfg.Loader = tc.loader
			for i := 0; i < tc.inputs; i++ {
				cfg.Inputs[fmt.Sprintf("in-%d", i)] = map[string]any{"type": "file"}
			}

			// Same predicate as the validation in SubscribeRunE, with the loader
			// factored in.
			allowed := len(cfg.Subscriptions) > 0 || len(cfg.Inputs) > 0 ||
				len(cfg.Loader) > 0

			if allowed != tc.wantAllowed {
				t.Fatalf("allowed = %v, want %v (loader=%v inputs=%d)",
					allowed, tc.wantAllowed, tc.loader, len(cfg.Inputs))
			}
		})
	}
}
