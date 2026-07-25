package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cozybadgerde/cress/internal/config"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, config.FileName)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	return path
}

func TestLoad_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	path := writeConfig(t, `
[site]
title = "Site"
base_url = "https://example.com/"
logo = "/logo.svg"
favicon = "/favicon.png"

[nav.main]
Home = "index.md"
Docs = "docs/index.md"
About = "about.md"

[nav.footer]
Imprint = "imprint.md"
Privacy = "privacy.md"
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Site.Title != "Site" {
		t.Errorf("title = %q", cfg.Site.Title)
	}
	if cfg.Site.Theme != config.DefaultTheme {
		t.Errorf("theme = %q, want default %q", cfg.Site.Theme, config.DefaultTheme)
	}
	if cfg.Site.Accent != config.DefaultAccent {
		t.Errorf("accent = %q, want default %q", cfg.Site.Accent, config.DefaultAccent)
	}
	if cfg.Site.Logo != "/logo.svg" || cfg.Site.Favicon != "/favicon.png" {
		t.Errorf("logo/favicon not passed through: %q, %q", cfg.Site.Logo, cfg.Site.Favicon)
	}
	if cfg.Site.BaseURL != "https://example.com" {
		t.Errorf("base_url = %q, want trailing slash trimmed", cfg.Site.BaseURL)
	}
	assertNavOrder(t, "nav.main", cfg.Nav.Main, []config.NavItem{
		{Title: "Home", Path: "index.md"},
		{Title: "Docs", Path: "docs/index.md"},
		{Title: "About", Path: "about.md"},
	})
	assertNavOrder(t, "nav.footer", cfg.Nav.Footer, []config.NavItem{
		{Title: "Imprint", Path: "imprint.md"},
		{Title: "Privacy", Path: "privacy.md"},
	})
}

// assertNavOrder checks one navigation group entry by entry, order included.
func assertNavOrder(t *testing.T, group string, got, want []config.NavItem) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("[%s] = %+v, want %d entries", group, got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%s][%d] = %+v, want %+v (order must follow the file)", group, i, got[i], want[i])
		}
	}
}

func TestLoad_navGroupsAreOptional_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	path := writeConfig(t, "[site]\ntitle = \"Site\"\n\n[nav.main]\nHome = \"index.md\"\n")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Nav.Main) != 1 {
		t.Errorf("nav.main = %+v, want 1 entry", cfg.Nav.Main)
	}
	if cfg.Nav.Footer != nil {
		t.Errorf("nav.footer = %+v, want nil when the table is absent", cfg.Nav.Footer)
	}
}

func TestLoad_unknownNavGroup_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// A mistyped group must fail rather than silently rendering nothing.
	path := writeConfig(t, "[site]\ntitle = \"Site\"\n\n[nav.sidebar]\nHome = \"index.md\"\n")

	if _, err := config.Load(path); err == nil {
		t.Fatal("expected an error for an unknown nav group")
	}
}

func TestLoad_unknownKey_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	path := writeConfig(t, "[site]\ntitel = \"typo\"\n")

	if _, err := config.Load(path); err == nil {
		t.Fatal("expected an error for an unknown key")
	}
}

func TestLoad_invalidAccent_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	path := writeConfig(t, "[site]\naccent = \"not-a-color\"\n")

	if _, err := config.Load(path); err == nil {
		t.Fatal("expected an error for an invalid accent color")
	}
}

func TestLoad_missing_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	_, err := config.Load(filepath.Join(t.TempDir(), "nope.toml"))
	if !errors.Is(err, config.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
