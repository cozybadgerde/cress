package theme_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/theme"
	"github.com/cozybadgerde/cress/internal/version"
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
	data := theme.PageData{Site: config.Site{Title: "T"}}
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
	data := theme.PageData{Site: config.Site{Title: "Hi"}}
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

func TestResolveReadsMeta_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := writeTheme(t, "mine", map[string]string{
		theme.MetaFile: "name = \"Mine\"\nlicense = \"MIT\"\ncress = \"1.1\"\n",
	})

	thm, err := theme.Resolve(root, "themes", "mine")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	meta := thm.Meta()
	if meta == nil {
		t.Fatal("Meta() = nil, want the theme's own metadata")
	}
	if meta.Name != "Mine" || meta.License != "MIT" || meta.Cress != "1.1" {
		t.Errorf("Meta() = %+v, want the file's fields", meta)
	}
	if warnings := meta.Warnings("1.1.0"); len(warnings) != 0 {
		t.Errorf("Warnings = %v, want none", warnings)
	}
}

// A theme with no metadata file loads exactly as it did before the file
// existed, which is the property that makes theme.toml optional rather than a
// new requirement.
func TestResolveWithoutMeta_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	thm, err := theme.Resolve(writeTheme(t, "mine", nil), "themes", "mine")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if thm.Meta() != nil {
		t.Errorf("Meta() = %+v, want nil", thm.Meta())
	}
}

func TestResolveMalformedMeta_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := writeTheme(t, "mine", map[string]string{theme.MetaFile: "name = \nnot toml"})

	_, err := theme.Resolve(root, "themes", "mine")
	if err == nil {
		t.Fatal("expected an error for a theme.toml that does not parse")
	}
	if !strings.Contains(err.Error(), theme.MetaFile) {
		t.Errorf("error %q does not name %s", err, theme.MetaFile)
	}
}

// The built-in theme describes itself, so the file a theme author is told to
// write is one the theme they read first actually has.
func TestBuiltinCarriesMeta_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	thm, err := theme.Resolve(t.TempDir(), "themes", config.DefaultTheme)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	meta := thm.Meta()
	if meta == nil {
		t.Fatalf("built-in theme ships no %s", theme.MetaFile)
	}
	if meta.Name == "" || meta.License == "" {
		t.Errorf("built-in %s = %+v, want at least a name and a license", theme.MetaFile, meta)
	}
	// It ships inside the binary, so the contract it was written against is
	// always the one rendering it. Declaring one could not catch a mismatch and
	// could only ever be wrong, which is why it declares none and why the theme
	// cress ships must never be the theme that warns.
	if meta.Cress != "" {
		t.Errorf("built-in theme declares cress = %q; it renders on whatever binary it ships in", meta.Cress)
	}
	if warnings := meta.Warnings(version.Version); len(warnings) != 0 {
		t.Errorf("built-in theme warns against cress %s: %v", version.Version, warnings)
	}
}

// The theme cress ships is the one people read before writing their own, so it
// has to pass the check the command applies to theirs.
func TestValidateBuiltin_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	report, err := theme.Validate(theme.ValidateOptions{
		SiteRoot:     t.TempDir(),
		ThemesDir:    config.ThemesDir,
		Name:         config.DefaultTheme,
		CressVersion: version.Version,
	})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !report.OK() {
		t.Errorf("the built-in theme does not satisfy its own contract: %v", report.Findings)
	}
}

func TestValidateOnDisk_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := writeTheme(t, "mine", map[string]string{
		"templates/wide.html":          "<img src=\"/logo.svg\">\n{{ .Page.Titel }}\n",
		"templates/partials/head.html": `{{ define "head" }}{{ .Site.Title }}{{ end }}`,
		theme.MetaFile:                 "name = \"Mine\"\ncress = \"1.1\"\n",
	})

	report, err := theme.Validate(theme.ValidateOptions{
		SiteRoot:     root,
		ThemesDir:    config.ThemesDir,
		Name:         "mine",
		CressVersion: "1.1.0",
	})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if report.Name != "mine" {
		t.Errorf("Name = %q, want %q", report.Name, "mine")
	}
	if !strings.Contains(report.Path, filepath.Join(config.ThemesDir, "mine")) {
		t.Errorf("Path = %q, want the theme directory", report.Path)
	}

	want := []string{"is not rooted under", "is not a field of PageView"}
	if len(report.Findings) != len(want) {
		t.Fatalf("got %d findings, want %d: %v", len(report.Findings), len(want), report.Findings)
	}
	for i, substr := range want {
		if !strings.Contains(report.Findings[i].Message, substr) {
			t.Errorf("finding %d = %q, want it to contain %q", i, report.Findings[i].Message, substr)
		}
		if report.Findings[i].File != filepath.ToSlash("templates/wide.html") {
			t.Errorf("finding %d is in %q, want templates/wide.html", i, report.Findings[i].File)
		}
	}
}

func TestValidateUnknownTheme_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// A theme that is not there is an error rather than a report: there is
	// nothing to validate, which is a different answer from "nothing is wrong".
	_, err := theme.Validate(theme.ValidateOptions{
		SiteRoot:  t.TempDir(),
		ThemesDir: config.ThemesDir,
		Name:      "does-not-exist",
	})
	if err == nil {
		t.Fatal("expected an error for a theme with no directory")
	}
}

// writeTheme lays out a minimal theme under root/themes/name, plus whatever
// extra files the caller names (relative to the theme directory), and returns
// the site root.
func writeTheme(t *testing.T, name string, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	dir := filepath.Join(root, "themes", name)
	if err := os.MkdirAll(filepath.Join(dir, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "templates", "page.html"), []byte(`{{ .Page.Title }}`), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	for rel, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return root
}
