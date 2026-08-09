package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/config"
)

// runCapturing runs the CLI and returns what it printed alongside the error, so
// a test can assert on the report rather than only on the exit status.
func runCapturing(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newRootCommand()
	cmd.Writer = &out
	cmd.ErrWriter = &out
	err := cmd.Run(context.Background(), append([]string{"cress"}, args...))
	return out.String(), err
}

// Both themes cress ships are the ones a theme author reads first, so they have
// to pass the check the command applies to theirs. Failing here means cress is
// shipping an example of something it tells people not to do.
func TestCLIValidateBundledThemes_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	if err := run(t, "init", site); err != nil {
		t.Fatalf("cress init: %v", err)
	}
	if err := run(t, "theme", "init", "mine", "--source", site); err != nil {
		t.Fatalf("cress theme init: %v", err)
	}

	for _, name := range []string{config.DefaultTheme, "mine"} {
		t.Run(name, func(t *testing.T) {
			out, err := runCapturing(t, "theme", "validate", name, "--source", site)
			if err != nil {
				t.Fatalf("cress theme validate %s failed:\n%s", name, out)
			}
			if !strings.Contains(out, "no problems found") {
				t.Errorf("output = %q, want it to report a clean theme", out)
			}
		})
	}
}

// With no name, the theme the site already renders with is the one meant.
func TestCLIValidateInfersTheThemeFromConfig_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	if err := run(t, "init", site); err != nil {
		t.Fatalf("cress init: %v", err)
	}
	if err := run(t, "theme", "init", "mine", "--source", site); err != nil {
		t.Fatalf("cress theme init: %v", err)
	}
	swapTheme(t, site, "mine")

	out, err := runCapturing(t, "theme", "validate", "--source", site)
	if err != nil {
		t.Fatalf("cress theme validate failed:\n%s", out)
	}
	if !strings.Contains(out, filepath.Join("themes", "mine")) {
		t.Errorf("output = %q, want it to name the theme from cress.toml", out)
	}
}

func TestCLIValidateReportsAndFails_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	if err := run(t, "init", site); err != nil {
		t.Fatalf("cress init: %v", err)
	}
	if err := run(t, "theme", "init", "mine", "--source", site); err != nil {
		t.Fatalf("cress theme init: %v", err)
	}
	writeFile(t, filepath.Join(site, "themes", "mine", "templates", "page.html"),
		"<link rel=\"stylesheet\" href=\"/style.css\">\n{{ .Page.Titel }}\n")

	out, err := runCapturing(t, "theme", "validate", "mine", "--source", site)
	if err == nil {
		t.Fatal("a theme with problems must exit non-zero")
	}
	// The report is the message, so nothing is added to it on the way out.
	if !errors.Is(err, errFailed) {
		t.Errorf("error = %v, want the silent failure that leaves the report to speak", err)
	}

	for _, want := range []string{
		"2 problem(s)",
		"templates/page.html:1",
		"is not rooted under",
		"templates/page.html:2",
		"did you mean Title?",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output is missing %q:\n%s", want, out)
		}
	}
}

// Quiet is for a pipeline that only wants the verdict.
func TestCLIValidateQuiet_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	if err := run(t, "init", site); err != nil {
		t.Fatalf("cress init: %v", err)
	}
	if err := run(t, "theme", "init", "mine", "--source", site); err != nil {
		t.Fatalf("cress theme init: %v", err)
	}
	writeFile(t, filepath.Join(site, "themes", "mine", "templates", "page.html"), "{{ .Page.Titel }}\n")

	out, err := runCapturing(t, "theme", "validate", "mine", "--source", site, "--quiet")
	if err == nil {
		t.Fatal("a theme with problems must exit non-zero, quiet or not")
	}
	if out != "" {
		t.Errorf("output = %q, want nothing", out)
	}
}

func TestCLIValidateRejectsBadInput_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	if err := run(t, "init", site); err != nil {
		t.Fatalf("cress init: %v", err)
	}

	t.Run("a theme with no directory", func(t *testing.T) {
		if _, err := runCapturing(t, "theme", "validate", "nope", "--source", site); err == nil {
			t.Fatal("expected an error for a theme that is not there")
		}
	})

	t.Run("a name that is a path", func(t *testing.T) {
		_, err := runCapturing(t, "theme", "validate", "../elsewhere", "--source", site)
		if err == nil {
			t.Fatal("expected an error for a theme name that is a path")
		}
		if !strings.Contains(err.Error(), "invalid theme name") {
			t.Errorf("error = %v, want it to name the rule", err)
		}
	})

	// With no config there is no theme to infer, which is worth saying rather
	// than validating the default theme somebody did not ask about.
	t.Run("no config and no name", func(t *testing.T) {
		_, err := runCapturing(t, "theme", "validate", "--source", t.TempDir())
		if err == nil {
			t.Fatal("expected an error with no config to infer a theme from")
		}
		if !strings.Contains(err.Error(), config.FileName) {
			t.Errorf("error = %v, want it to name the missing config", err)
		}
	})
}

// swapTheme points the scaffolded site at a different theme.
func swapTheme(t *testing.T, site, name string) {
	t.Helper()
	path := filepath.Join(site, config.FileName)
	cfg, err := os.ReadFile(path) // #nosec G304 -- test path
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	swapped := strings.Replace(string(cfg), `theme = "cress"`, `theme = "`+name+`"`, 1)
	if swapped == string(cfg) {
		t.Fatal("scaffolded cress.toml no longer names the default theme; update this test")
	}
	writeFile(t, path, swapped)
}
