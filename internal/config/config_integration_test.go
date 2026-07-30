package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	// An unset accent stays unset, so the theme's own per-scheme accent applies.
	if cfg.Site.Accent != "" || cfg.Site.AccentDark != "" {
		t.Errorf("accent/accent_dark = %q/%q, want both empty when unset", cfg.Site.Accent, cfg.Site.AccentDark)
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

func TestLoad_accents_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	for _, tc := range []struct {
		name            string
		body            string
		wantErr         bool
		accent, accent2 string
	}{
		{name: "invalid accent", body: "accent = \"not-a-color\"\n", wantErr: true},
		{name: "invalid accent_dark", body: "accent_dark = \"not-a-color\"\n", wantErr: true},
		{name: "both set", body: "accent = \"#4f7a4a\"\naccent_dark = \"#9ccb8f\"\n", accent: "#4f7a4a", accent2: "#9ccb8f"},
		// Only a light accent: it applies to both schemes, which is what the
		// theme's cascade does when no dark override is emitted.
		{name: "accent only", body: "accent = \"#4f7a4a\"\n", accent: "#4f7a4a"},
		// Coherent on its own: the theme's accent in light, this one in dark.
		{name: "accent_dark only", body: "accent_dark = \"#9ccb8f\"\n", accent2: "#9ccb8f"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Load(writeConfig(t, "[site]\n"+tc.body))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Load(%q) = nil error, want one", tc.body)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load(%q): %v", tc.body, err)
			}
			if cfg.Site.Accent != tc.accent || cfg.Site.AccentDark != tc.accent2 {
				t.Errorf("accent/accent_dark = %q/%q, want %q/%q", cfg.Site.Accent, cfg.Site.AccentDark, tc.accent, tc.accent2)
			}
		})
	}
}

// Unlike accent_dark, which is coherent on its own because the theme supplies
// the light-scheme value it pairs with, logo_dark has nothing to fall back on:
// a logo is site identity, so no theme default can stand in for the light one.
func TestLoad_logos_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	for _, tc := range []struct {
		name       string
		body       string
		wantErr    bool
		logo, dark string
	}{
		{name: "neither set", body: ""},
		{name: "logo only", body: "logo = \"/logo.svg\"\n", logo: "/logo.svg"},
		{
			name: "both set",
			body: "logo = \"/logo.svg\"\nlogo_dark = \"/logo-dark.svg\"\n",
			logo: "/logo.svg", dark: "/logo-dark.svg",
		},
		{name: "logo_dark without logo", body: "logo_dark = \"/logo-dark.svg\"\n", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Load(writeConfig(t, "[site]\n"+tc.body))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Load(%q) = nil error, want one", tc.body)
				}
				if !strings.Contains(err.Error(), "logo_dark") {
					t.Errorf("error %q does not name the offending key", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load(%q): %v", tc.body, err)
			}
			if cfg.Site.Logo != tc.logo || cfg.Site.LogoDark != tc.dark {
				t.Errorf("logo/logo_dark = %q/%q, want %q/%q", cfg.Site.Logo, cfg.Site.LogoDark, tc.logo, tc.dark)
			}
		})
	}
}

// base_url decides where in-site links are rooted, so a value that parses as
// something other than the author meant is worse than a rejected one: the site
// builds, publishes, and 404s everywhere.
func TestLoad_baseURL_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	for _, tc := range []struct {
		name     string
		body     string
		wantErr  bool
		baseURL  string
		basePath string
	}{
		{name: "unset", body: ""},
		{name: "domain root", body: "base_url = \"https://example.com\"\n", baseURL: "https://example.com"},
		{name: "domain root with slash", body: "base_url = \"https://example.com/\"\n", baseURL: "https://example.com"},
		{
			name: "project page", body: "base_url = \"https://user.github.io/cress\"\n",
			baseURL: "https://user.github.io/cress", basePath: "/cress",
		},
		{
			name: "project page with slash", body: "base_url = \"https://user.github.io/cress/\"\n",
			baseURL: "https://user.github.io/cress", basePath: "/cress",
		},
		{
			name: "nested path", body: "base_url = \"https://example.com/a/b\"\n",
			baseURL: "https://example.com/a/b", basePath: "/a/b",
		},
		// A host with no scheme parses as a bare path, which would root every link
		// under a directory named after the domain.
		{name: "no scheme", body: "base_url = \"example.com/cress\"\n", wantErr: true},
		{name: "path only", body: "base_url = \"/cress\"\n", wantErr: true},
		{name: "scheme only", body: "base_url = \"https://\"\n", wantErr: true},
		{name: "unparseable", body: "base_url = \"://nope\"\n", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Load(writeConfig(t, "[site]\n"+tc.body))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Load(%q) = nil error, want one", tc.body)
				}
				if !strings.Contains(err.Error(), "base_url") {
					t.Errorf("error should name base_url: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load(%q): %v", tc.body, err)
			}
			if cfg.Site.BaseURL != tc.baseURL {
				t.Errorf("BaseURL = %q, want %q", cfg.Site.BaseURL, tc.baseURL)
			}
			if cfg.Site.BasePath != tc.basePath {
				t.Errorf("BasePath = %q, want %q", cfg.Site.BasePath, tc.basePath)
			}
		})
	}
}

// base_path is derived from base_url, so setting it directly is a typo rather
// than a second way to say the same thing.
func TestLoad_basePathIsNotConfigurable_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	_, err := config.Load(writeConfig(t, "[site]\nbase_path = \"/cress\"\n"))
	if err == nil {
		t.Fatal("Load = nil error, want base_path rejected as an unknown key")
	}
	if !strings.Contains(err.Error(), "base_path") {
		t.Errorf("error should name the offending key: %v", err)
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
