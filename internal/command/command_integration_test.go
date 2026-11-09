package command_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

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
			if !strings.Contains(stdout.String(), "Built 2 page(s)") ||
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

func TestSiteInitCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, mode := range []string{"current directory", "empty directory", "new directory", "flags after directory"} {
		t.Run(mode, func(t *testing.T) {
			site := t.TempDir()
			args := []string{"cress", "site", "init"}
			name := "poppy"
			switch mode {
			case "current directory":
				t.Chdir(site)
			case "empty directory":
				args = append(args, site)
			case "new directory":
				site = filepath.Join(site, "new site")
				name = "anemone"
				args = append(args, "--theme", name, site)
			case "flags after directory":
				name = "clover"
				args = append(args, site, "--theme", name)
			}
			var stdout, stderr bytes.Buffer
			cmd := command.NewCommand(command.Deps{Version: "dev", Output: &stdout, Errors: &stderr})
			if err := cmd.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			config, err := os.ReadFile(filepath.Join(site, "cress.toml"))
			if err != nil || !strings.Contains(string(config), `name = "`+name+`"`) {
				t.Fatalf("selected theme: %s, %v", config, err)
			}
			if !strings.Contains(stdout.String(), "Created site in") {
				t.Fatalf("missing initialization message: %s", stdout.String())
			}
			for _, path := range []string{"public", ".cress", ".git"} {
				if _, err := os.Lstat(filepath.Join(site, path)); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("initialization created %s or inspection failed: %v", path, err)
				}
			}
			stdout.Reset()
			cmd = command.NewCommand(command.Deps{Version: "dev", Output: &stdout, Errors: &stderr})
			if err := cmd.Run(context.Background(), args); err == nil || stdout.Len() != 0 {
				t.Fatalf("existing target accepted or reported success: %v, %s", err, stdout.String())
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

func TestPreviewSignalProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process integration test")
	}
	if site := os.Getenv("CRESS_TEST_PREVIEW_SITE"); site != "" {
		cmd := command.NewCommand(command.Deps{Version: "dev", Output: os.Stdout, Errors: os.Stderr})
		if err := cmd.Run(context.Background(), []string{"cress", "site", "preview", "--port", "0", site}); err != nil {
			t.Fatal(err)
		}
		return
	}
	site := t.TempDir()
	if err := os.Mkdir(filepath.Join(site, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(site, "content/index.md"), []byte("+++\ntitle = 'Home'\n+++"), 0o644); err != nil {
		t.Fatal(err)
	}
	process := exec.Command(os.Args[0], "-test.run=^TestPreviewSignalProcess$")
	process.Env = append(os.Environ(), "CRESS_TEST_PREVIEW_SITE="+site)
	var stderr bytes.Buffer
	process.Stderr = &stderr
	stdout, err := process.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := process.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		if !waited {
			if err := process.Process.Kill(); err != nil {
				t.Error(err)
			}
			if err := process.Wait(); err == nil {
				t.Error("expected killed process to fail")
			}
		}
	}()
	ready := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "Preview http://127.0.0.1:") {
				ready <- true
				return
			}
		}
		ready <- false
	}()
	select {
	case ok := <-ready:
		if !ok {
			t.Fatal("preview exited before readiness")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("preview never became ready")
	}
	if err := process.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- process.Wait() }()
	select {
	case err := <-done:
		waited = true
		if err != nil {
			t.Fatalf("SIGTERM did not shut down cleanly: %v\n%s", err, stderr.String())
		}
	case <-time.After(10 * time.Second):
		if err := process.Process.Kill(); err != nil {
			t.Error(err)
		}
		if err := <-done; err == nil {
			t.Error("expected killed preview to fail")
		}
		waited = true
		t.Fatal("SIGTERM did not stop preview")
	}
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
