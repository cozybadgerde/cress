package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/build"
	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/theme"
	"github.com/cozybadgerde/cress/internal/version"
)

// writeFile writes one file under dir, creating parents.
func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// The starter theme is boilerplate a person copies, so it has to satisfy the
// contract it demonstrates. Rendering it is what catches drift: html/template
// fails on a field a struct does not have, so renaming one in PageData breaks
// this test rather than quietly emitting an empty string in every scaffolded
// theme.
func TestCLIThemeInitThenBuild_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	writeFile(t, filepath.Join(site, "cress.toml"), `
[site]
title = "Test Site"
description = "A site for the starter theme."
theme = "mine"
footer = "A small corner."
copyright = "(c) Somebody"
accent = "#4f7a4a"

[nav.main]
Home = "index.md"
`)
	writeFile(t, filepath.Join(site, "content", "index.md"), "# Hello\n\nSome prose.\n")
	writeFile(t, filepath.Join(site, "content", "front.md"), "---\ntitle: Front\nlayout: landing\n---\n\n# Front\n")

	if err := run(t, "theme", "init", "mine", "--source", site); err != nil {
		t.Fatalf("cress theme init: %v", err)
	}
	if err := run(t, "build", "--source", site); err != nil {
		t.Fatalf("cress build: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(site, "public", "index.html")) // #nosec G304 -- test path
	if err != nil {
		t.Fatalf("read built page: %v", err)
	}
	got := string(out)

	// One assertion per part of the contract the starter demonstrates, so a
	// failure names the piece that stopped working rather than the whole page.
	for _, want := range []struct{ what, marker string }{
		{"the site title", "Test Site"},
		{"the document language", `<html lang="en">`},
		{"cress-rendered head metadata", `<meta name="description"`},
		{"the theme's own stylesheet link", `href="/style.css"`},
		{"the configured accent", "--accent: #4f7a4a"},
		{"the main navigation", `<nav>`},
		{"the current page marked", `aria-current="page"`},
		{"the rendered Markdown body", "<h1>Hello</h1>"},
		{"the footer line", "A small corner."},
		{"the copyright", "(c) Somebody"},
	} {
		if !strings.Contains(got, want.marker) {
			t.Errorf("built page is missing %s (%q):\n%s", want.what, want.marker, got)
		}
	}

	// The starter ships a second layout so that the scaffolded site, which asks
	// for `landing`, builds without a warning. A page reaching it must render
	// through it, and an ordinary page must not.
	front, err := os.ReadFile(filepath.Join(site, "public", "front.html")) // #nosec G304 -- test path
	if err != nil {
		t.Fatalf("read landing page: %v", err)
	}
	if !strings.Contains(string(front), "<section>") {
		t.Errorf("front.html did not render through the landing layout:\n%s", front)
	}
	if strings.Contains(got, "<section>") {
		t.Error("index.html rendered through the landing layout, want the entry template")
	}
	// Both layouts call the same partials, which is what keeps them one site.
	if !strings.Contains(string(front), "<nav>") {
		t.Errorf("the landing layout is missing the shared header:\n%s", front)
	}
}

// The scaffolded site names `layout: landing`, so scaffolding a theme and
// switching to it has to build clean. A starter that warns on its first build
// is a starter that teaches its user to ignore warnings.
func TestCLIThemeInitMatchesScaffoldedSite_integration(t *testing.T) {
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

	cfg, err := os.ReadFile(filepath.Join(site, "cress.toml")) // #nosec G304 -- test path
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	swapped := strings.Replace(string(cfg), `theme = "cress"`, `theme = "mine"`, 1)
	if swapped == string(cfg) {
		t.Fatal("scaffolded cress.toml no longer names the default theme; update this test")
	}
	writeFile(t, filepath.Join(site, "cress.toml"), swapped)

	// An untagged local build has no version, which would skip the contract
	// check and leave the starter's theme.toml out of the clean-build promise
	// entirely. Running as the version the starter declares puts it back in: a
	// typo in that declaration, or a key cress does not know, then warns here.
	thm, err := theme.Resolve(site, config.ThemesDir, "mine")
	if err != nil {
		t.Fatalf("resolve scaffolded theme: %v", err)
	}
	if thm.Meta() == nil || thm.Meta().Cress == "" {
		t.Fatalf("the scaffolded theme declares no contract; update this test")
	}
	previous := version.Version
	t.Cleanup(func() { version.Version = previous })
	version.Version = thm.Meta().Cress

	// Through the build package rather than the CLI, because a warning is
	// printed rather than returned and this test is about what it says.
	res, err := build.Build(build.Options{Root: site})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("building the scaffolded site with the scaffolded theme warned: %v", res.Warnings)
	}
}

func TestCLIThemeInitRequiresName_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if err := run(t, "theme", "init", "--source", t.TempDir()); err == nil {
		t.Fatal("expected an error when no theme name is given")
	}
}

// The name is joined onto themes/, so a path has to be refused before it
// becomes one.
func TestCLIThemeInitRejectsPathName_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	for _, name := range []string{"../escape", "a/b", ".."} {
		t.Run(name, func(t *testing.T) {
			if err := run(t, "theme", "init", name, "--source", site); err == nil {
				t.Errorf("expected an error for theme name %q", name)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(site, "themes")); !os.IsNotExist(err) {
		t.Errorf("a refused name must not create themes/: %v", err)
	}
}

func TestCLIThemeInitRefusesNonEmpty_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	writeFile(t, filepath.Join(site, "themes", "mine", "templates", "page.html"), "MINE")

	if err := run(t, "theme", "init", "mine", "--source", site); err == nil {
		t.Fatal("expected an error for a theme directory that already has files")
	}
	if err := run(t, "theme", "init", "mine", "--source", site, "--allow-existing"); err != nil {
		t.Fatalf("cress theme init --allow-existing: %v", err)
	}

	kept, err := os.ReadFile(filepath.Join(site, "themes", "mine", "templates", "page.html")) // #nosec G304 -- test path
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(kept) != "MINE" {
		t.Errorf("page.html = %q, want the existing file untouched", kept)
	}
}

func TestCLIThemeUnknownSubcommand_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if err := run(t, "theme", "frobnicate"); err == nil {
		t.Fatal("expected an error for an unknown theme subcommand")
	}
}
