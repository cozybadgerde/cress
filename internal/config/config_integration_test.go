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

[nav]
Home = "index.md"
Docs = "docs/index.md"
About = "about.md"
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
	wantOrder := []config.NavItem{
		{Title: "Home", Path: "index.md"},
		{Title: "Docs", Path: "docs/index.md"},
		{Title: "About", Path: "about.md"},
	}
	if len(cfg.Nav) != len(wantOrder) {
		t.Fatalf("nav = %+v, want %d entries", cfg.Nav, len(wantOrder))
	}
	for i, want := range wantOrder {
		if cfg.Nav[i] != want {
			t.Errorf("nav[%d] = %+v, want %+v (order must follow the file)", i, cfg.Nav[i], want)
		}
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
