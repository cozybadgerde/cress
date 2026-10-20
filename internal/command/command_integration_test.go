package command_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/command"
)

func TestBuildCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, explicit := range []bool{true, false} {
		name := "current directory"
		if explicit {
			name = "explicit directory"
		}
		t.Run(name, func(t *testing.T) {
			site := t.TempDir()
			if err := os.Mkdir(filepath.Join(site, "content"), 0o755); err != nil {
				t.Fatal(err)
			}
			source := "+++\ntitle = 'Home'\nunknown = true\n+++\nHello"
			if err := os.WriteFile(filepath.Join(site, "content", "index.md"), []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			args := []string{"cress", "site", "build"}
			if explicit {
				args = append(args, site)
			} else {
				t.Chdir(site)
			}
			var stdout, stderr bytes.Buffer
			cmd := command.NewCommand(command.Deps{Version: "dev", Output: &stdout, Errors: &stderr})
			if err := cmd.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(stdout.String(), "Built 1 page(s)") ||
				!strings.Contains(stderr.String(), `warning: content/index.md: unknown frontmatter field "unknown"`) {
				t.Fatalf("stdout = %q; stderr = %q", stdout.String(), stderr.String())
			}
			stdout.Reset()
			cmd = command.NewCommand(command.Deps{Version: "dev", Output: &stdout, Errors: &stderr})
			if err := cmd.Run(context.Background(), args); err != nil {
				t.Fatalf("owned output could not be rebuilt: %v", err)
			}
			stdout.Reset()
			if err := os.WriteFile(filepath.Join(site, "content/index.md"), []byte("missing title"), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd = command.NewCommand(command.Deps{Version: "dev", Output: &stdout, Errors: &stderr})
			if err := cmd.Run(context.Background(), args); err == nil {
				t.Fatal("invalid rebuild accepted")
			}
			if strings.Contains(stdout.String(), "Built") {
				t.Fatal("failed build reported success")
			}
		})
	}
}

func TestThemeInitCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	var stdout, stderr bytes.Buffer
	cmd := command.NewCommand(command.Deps{Version: "1.0.0", Output: &stdout, Errors: &stderr})
	args := []string{"cress", "theme", "init", "demo", site}
	if err := cmd.Run(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "Created themes/demo/") {
		t.Fatalf("missing initialization message: %s", stdout.String())
	}
	manifest := filepath.Join(site, "themes/demo/theme.json")
	previous, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	cmd = command.NewCommand(command.Deps{Version: "1.0.0", Output: &stdout, Errors: &stderr})
	if err := cmd.Run(context.Background(), args); err == nil {
		t.Fatal("existing theme target accepted")
	}
	current, err := os.ReadFile(manifest)
	if err != nil || !bytes.Equal(previous, current) || stdout.Len() != 0 {
		t.Fatalf("failed initialization modified files or claimed success: %s, %v", stdout.String(), err)
	}
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestWarningFailurePreservesBuildError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	if err := os.Mkdir(filepath.Join(site, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	source := "+++\ntitle = 1\nunknown = true\n+++"
	if err := os.WriteFile(filepath.Join(site, "content", "index.md"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	cmd := command.NewCommand(command.Deps{Version: "dev", Output: &stdout, Errors: brokenWriter{}})
	err := cmd.Run(context.Background(), []string{"cress", "site", "build", site})
	if !errors.Is(err, io.ErrClosedPipe) || !strings.Contains(err.Error(), "title must be") {
		t.Fatalf("error = %v, want both validation and diagnostic-output failures", err)
	}
}
