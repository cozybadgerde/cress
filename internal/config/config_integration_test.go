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

// mustLoad loads a config that is expected to be valid, failing the test if it
// is not.
//
// The tests below come in pairs, one for what a key accepts and one for what it
// refuses, rather than a single table with a wantErr column. A row in such a
// table carries the columns for both jobs and uses half of them, and the body
// has to branch before it can assert anything. These two helpers are what the
// pairs cost: each table then does one job and reads straight down.
func mustLoad(t *testing.T, body string) *config.Config {
	t.Helper()
	cfg, err := config.Load(writeConfig(t, body))
	if err != nil {
		t.Fatalf("Load(%q): %v", body, err)
	}
	return cfg
}

// assertRejected checks that a config is refused, and that the error quotes
// each of names. What the message says matters as much as the refusal: these
// are values an author typed on purpose, so the error has to point at the key
// and at their own text rather than announce that something is wrong.
func assertRejected(t *testing.T, body string, names ...string) {
	t.Helper()
	_, err := config.Load(writeConfig(t, body))
	if err == nil {
		t.Fatalf("Load(%q) = nil error, want a refusal", body)
	}
	for _, name := range names {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not name %s", err, name)
		}
	}
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

// Absence is the current behaviour rather than a missing setting: a site that
// never heard of the table renders exactly as it did before the table existed.
func TestLoad_markdown_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tests := []struct {
		name string
		body string
		want bool
	}{
		{"no table at all", "[site]\ntitle = \"Site\"\n", false},
		{"an empty table", "[site]\ntitle = \"Site\"\n\n[markdown]\n", false},
		{"turned off", "[site]\ntitle = \"Site\"\n\n[markdown]\nhighlight = false\n", false},
		{"turned on", "[site]\ntitle = \"Site\"\n\n[markdown]\nhighlight = true\n", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := mustLoad(t, tc.body).Markdown.Highlight; got != tc.want {
				t.Errorf("markdown.highlight = %v, want %v", got, tc.want)
			}
		})
	}
}

// The table's keys are as fixed as every other group's, so a mistyped one is
// reported rather than dropped: silently ignored, it would read as a setting
// that had no effect.
func TestLoad_rejectsBadMarkdownKeys_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	assertRejected(t, "[site]\ntitle = \"Site\"\n\n[markdown]\nhighlite = true\n", "highlite")
	assertRejected(t, "[site]\ntitle = \"Site\"\n\n[markdown]\nhighlight = \"yes\"\n", "highlight")
}

func TestLoad_accents_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	for _, tc := range []struct {
		name            string
		body            string
		accent, accent2 string
	}{
		{name: "both set", body: "accent = \"#4f7a4a\"\naccent_dark = \"#9ccb8f\"\n", accent: "#4f7a4a", accent2: "#9ccb8f"},
		// Only a light accent: it applies to both schemes, which is what the
		// theme's cascade does when no dark override is emitted.
		{name: "accent only", body: "accent = \"#4f7a4a\"\n", accent: "#4f7a4a"},
		// Coherent on its own: the theme's accent in light, this one in dark.
		{name: "accent_dark only", body: "accent_dark = \"#9ccb8f\"\n", accent2: "#9ccb8f"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := mustLoad(t, "[site]\n"+tc.body)
			if cfg.Site.Accent != tc.accent || cfg.Site.AccentDark != tc.accent2 {
				t.Errorf("accent/accent_dark = %q/%q, want %q/%q", cfg.Site.Accent, cfg.Site.AccentDark, tc.accent, tc.accent2)
			}
		})
	}
}

func TestLoad_rejectsInvalidAccents_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	for _, tc := range []struct{ name, body string }{
		{name: "invalid accent", body: "accent = \"not-a-color\"\n"},
		{name: "invalid accent_dark", body: "accent_dark = \"not-a-color\"\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertRejected(t, "[site]\n"+tc.body)
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
		logo, dark string
	}{
		{name: "neither set", body: ""},
		{name: "logo only", body: "logo = \"/logo.svg\"\n", logo: "/logo.svg"},
		{
			name: "both set",
			body: "logo = \"/logo.svg\"\nlogo_dark = \"/logo-dark.svg\"\n",
			logo: "/logo.svg", dark: "/logo-dark.svg",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := mustLoad(t, "[site]\n"+tc.body)
			if cfg.Site.Logo != tc.logo || cfg.Site.LogoDark != tc.dark {
				t.Errorf("logo/logo_dark = %q/%q, want %q/%q", cfg.Site.Logo, cfg.Site.LogoDark, tc.logo, tc.dark)
			}
		})
	}
}

func TestLoad_rejectsALoneDarkLogo_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	assertRejected(t, "[site]\nlogo_dark = \"/logo-dark.svg\"\n", "logo_dark")
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
		baseURL  string
		basePath string
	}{
		{name: "unset", body: ""},
		// Blank in any spelling means the same as absent: no site root named, which
		// is what a site served from a domain root wants. The theme key already
		// treats blank as unset, and a site reachable at several domains has no one
		// canonical host to name here anyway.
		{name: "empty", body: "base_url = \"\"\n"},
		{name: "whitespace only", body: "base_url = \"   \"\n"},
		{name: "a bare slash", body: "base_url = \"/\"\n"},
		{name: "nothing but slashes", body: "base_url = \"//\"\n"},
		{name: "domain root", body: "base_url = \"https://example.com\"\n", baseURL: "https://example.com"},
		{name: "domain root with slash", body: "base_url = \"https://example.com/\"\n", baseURL: "https://example.com"},
		{
			name: "surrounding whitespace is trimmed", body: "base_url = \"  https://example.com/cress  \"\n",
			baseURL: "https://example.com/cress", basePath: "/cress",
		},
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := mustLoad(t, "[site]\n"+tc.body)
			if cfg.Site.BaseURL != tc.baseURL {
				t.Errorf("BaseURL = %q, want %q", cfg.Site.BaseURL, tc.baseURL)
			}
			if cfg.Site.BasePath != tc.basePath {
				t.Errorf("BasePath = %q, want %q", cfg.Site.BasePath, tc.basePath)
			}
		})
	}
}

// A rejected base_url has to quote the author's own text back at them, because
// the values that get refused are the ones that look right.
func TestLoad_rejectsInvalidBaseURL_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	for _, tc := range []struct{ name, value string }{
		// A host with no scheme parses as a bare path, which would root every link
		// under a directory named after the domain.
		{name: "no scheme", value: "example.com/cress"},
		{name: "path only", value: "/cress"},
		// Canonicalizing before validating once made this one report a value
		// nobody wrote ("https:").
		{name: "scheme only", value: "https://"},
		{name: "unparseable", value: "://nope"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertRejected(t, "[site]\nbase_url = \""+tc.value+"\"\n", "base_url", `"`+tc.value+`"`)
		})
	}
}

// Every document declares a language, so the key defaults rather than staying
// empty. Blank means unset, matching how theme and base_url already read it.
func TestLoad_language_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	for _, tc := range []struct {
		name, body, want string
	}{
		{name: "unset defaults", body: "", want: config.DefaultLanguage},
		{name: "empty defaults", body: "language = \"\"\n", want: config.DefaultLanguage},
		{name: "whitespace defaults", body: "language = \"  \"\n", want: config.DefaultLanguage},
		{name: "a set tag is kept", body: "language = \"de\"\n", want: "de"},
		{name: "a region tag is kept", body: "language = \"en-GB\"\n", want: "en-GB"},
		{name: "surrounding whitespace is trimmed", body: "language = \" de \"\n", want: "de"},
		// No registry check: a typo reaches the output rather than failing a build
		// over a value cress cannot authoritatively judge.
		{name: "an unknown tag is passed through", body: "language = \"klingon\"\n", want: "klingon"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Load(writeConfig(t, "[site]\n"+tc.body))
			if err != nil {
				t.Fatalf("Load(%q): %v", tc.body, err)
			}
			if cfg.Site.Language != tc.want {
				t.Errorf("Language = %q, want %q", cfg.Site.Language, tc.want)
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

// A theme name selects which templates render every page, and it is joined onto
// themes/ to find them. filepath.Join resolves ".." rather than refusing it, so
// anything but a single directory name has to be rejected here: otherwise a
// config could point the build at a directory outside the site and publish that
// directory's static/ assets along with it.
func TestLoad_theme_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	for _, tc := range []struct{ name, body, want string }{
		// Unset normalizes to the built-in theme, so the check has to run after
		// normalize and still pass the name it substituted.
		{name: "unset", body: "", want: config.DefaultTheme},
		{name: "a plain name", body: "theme = \"mine\"\n", want: "mine"},
		{name: "a name with punctuation", body: "theme = \"my-theme_2\"\n", want: "my-theme_2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := mustLoad(t, "[site]\n"+tc.body).Site.Theme; got != tc.want {
				t.Errorf("theme = %q, want %q", got, tc.want)
			}
		})
	}
}

// The theme name is the one config value that chooses which code runs, since it
// is joined onto themes/ to find the templates every page renders through. So
// anything that is not a single directory name is refused rather than cleaned
// up: filepath.Join resolves ".." instead of rejecting it, and a theme outside
// the site would render every page and copy its static/ into the output.
func TestLoad_rejectsAThemeThatIsAPath_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	for _, tc := range []struct{ name, theme string }{
		{name: "parent traversal", theme: "../../evil"},
		{name: "single parent", theme: ".."},
		{name: "current directory", theme: "."},
		{name: "a nested path", theme: "mine/nested"},
		{name: "an absolute path", theme: "/etc"},
		{name: "a trailing separator", theme: "mine/"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertRejected(t, "[site]\ntheme = \""+tc.theme+"\"\n", "invalid theme")
		})
	}
}
