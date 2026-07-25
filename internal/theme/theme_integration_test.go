package theme_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/theme"
)

func TestResolveBuiltin_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// An empty site root has no themes/ directory, so the default name resolves
	// to the embedded theme.
	thm, err := theme.Resolve(t.TempDir(), "themes", config.DefaultTheme)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if thm.Name() != config.DefaultTheme {
		t.Errorf("Name() = %q, want %q", thm.Name(), config.DefaultTheme)
	}
	if thm.StaticFS() == nil {
		t.Error("built-in theme should ship static assets")
	}

	var buf bytes.Buffer
	data := struct {
		Site config.Site
		Nav  struct{ Main, Footer []any }
		Page struct{ Title, URL, HTML string }
	}{Site: config.Site{Title: "T"}}
	if err := thm.Render(&buf, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(buf.String(), "<!DOCTYPE html>") {
		t.Errorf("built-in theme output missing doctype:\n%s", buf.String())
	}
}

func TestResolveDiskOverride_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	dir := filepath.Join(root, "themes", "mine", "templates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "page.html"), []byte(`CUSTOM:{{ .Site.Title }}`), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	thm, err := theme.Resolve(root, "themes", "mine")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	var buf bytes.Buffer
	data := struct{ Site config.Site }{Site: config.Site{Title: "Hi"}}
	if err := thm.Render(&buf, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if buf.String() != "CUSTOM:Hi" {
		t.Errorf("Render = %q, want the on-disk theme", buf.String())
	}
}

func TestResolveUnknown_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if _, err := theme.Resolve(t.TempDir(), "themes", "does-not-exist"); err == nil {
		t.Fatal("expected an error for an unknown theme with no directory")
	}
}
