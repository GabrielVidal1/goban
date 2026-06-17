package cli

import (
	"reflect"
	"testing"
)

func TestParseTopLevel(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantDir  string
		wantPort string
		wantRest []string
	}{
		{
			name:     "subcommand flags pass through untouched",
			args:     []string{"ticket", "create", "homelab", "Done", "My title", "--priority", "high", "--tags", "a,b"},
			wantRest: []string{"ticket", "create", "homelab", "Done", "My title", "--priority", "high", "--tags", "a,b"},
		},
		{
			name:     "list flag after positional passes through",
			args:     []string{"tickets", "list", "homelab", "--column", "Done", "--json"},
			wantRest: []string{"tickets", "list", "homelab", "--column", "Done", "--json"},
		},
		{
			name:     "global flag before command is consumed",
			args:     []string{"--dir", "/tmp/k", "tickets", "list", "homelab"},
			wantDir:  "/tmp/k",
			wantRest: []string{"tickets", "list", "homelab"},
		},
		{
			name:     "global flag after command is consumed (serve --port)",
			args:     []string{"serve", "--port", "9090"},
			wantPort: "9090",
			wantRest: []string{"serve"},
		},
		{
			name:     "inline =value form for global flag",
			args:     []string{"--dir=/tmp/k", "serve"},
			wantDir:  "/tmp/k",
			wantRest: []string{"serve"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, rest := ParseTopLevel(tc.args)
			if tc.wantDir != "" && g.Dir != tc.wantDir {
				t.Errorf("Dir = %q, want %q", g.Dir, tc.wantDir)
			}
			if tc.wantPort != "" && g.Port != tc.wantPort {
				t.Errorf("Port = %q, want %q", g.Port, tc.wantPort)
			}
			if !reflect.DeepEqual(rest, tc.wantRest) {
				t.Errorf("rest = %#v, want %#v", rest, tc.wantRest)
			}
		})
	}
}
