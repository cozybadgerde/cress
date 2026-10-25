package command_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/command"
)

func TestCommands(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      string
		wantError bool
	}{
		{name: "root", want: "site"},
		{name: "root help", args: []string{"--help"}, want: "theme"},
		{name: "version", args: []string{"--version"}, want: "cress version test-version"},
		{name: "site", args: []string{"site"}, want: "build"},
		{name: "site help", args: []string{"site", "--help"}, want: "build"},
		{name: "theme", args: []string{"theme"}, want: "Theme development"},
		{name: "theme help", args: []string{"theme", "--help"}, want: "Theme development"},
		{name: "build help", args: []string{"site", "build", "--help"}, want: "[directory]"},
		{name: "preview help", args: []string{"site", "preview", "--help"}, want: "--drafts"},
		{name: "preview too many directories", args: []string{"site", "preview", "one", "two"}, wantError: true},
		{name: "preview invalid port", args: []string{"site", "preview", "--port", "-1"}, wantError: true},
		{name: "unknown root command", args: []string{"unknown"}, wantError: true},
		{name: "unknown site command", args: []string{"site", "unknown"}, wantError: true},
		{name: "unknown theme command", args: []string{"theme", "unknown"}, wantError: true},
		{name: "site init help", args: []string{"site", "init", "--help"}, want: "--theme"},
		{name: "invalid init theme", args: []string{"site", "init", "--theme", "unknown"}, wantError: true},
		{name: "unknown option", args: []string{"--unknown"}, wantError: true},
		{name: "too many directories", args: []string{"site", "build", "one", "two"}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := command.NewCommand(command.Deps{
				Version: "test-version", Output: &stdout, Errors: &stderr,
			})
			err := cmd.Run(context.Background(), append([]string{"cress"}, test.args...))
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, want error %v; output %s", err, test.wantError, stdout.String())
			}
			if test.want != "" && !strings.Contains(stdout.String(), test.want) {
				t.Errorf("output = %q, want %q", stdout.String(), test.want)
			}
		})
	}
}
