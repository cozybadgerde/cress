package generator_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/generator"
)

func TestBuildConfiguredSite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{
		"cress.toml": `
[site]
title = "<Business>"
description = 'Shared "description"'
copyright = "<Copyright>"
language = "de"
host = "https://publication.example.org"
path = "/company/"
[site.image]
src = "https://images.example.org/team.jpg"
alt = "<Team>"
[[navigation.main]]
label = "First"
url = "/first/"
[[navigation.main]]
label = "<Second>"
url = "/second/?a=1&b=2#team"
[[navigation.footer]]
label = "External"
url = "https://example.net/"
[theme.custom]
private = "not raw configuration for templates"
`,
		"content/index.md":         "+++\ntitle = 'Home'\n+++\nHello",
		"content/landing.md":       "+++\ntitle = 'Landing'\ntype = 'landing'\n+++\nLanding intro",
		"content/hidden-list.md":   "+++\ndraft = true\ntype = 'list'\n+++\nHidden",
		"content/clear.md":         "+++\ntitle = 'Clear'\ntype = 'landing'\ndescription = ''\n[image]\nsrc = ''\n+++\nPlain",
		"content/section/index.md": "+++\ntitle = 'Section'\ndescription = 'Section only'\n+++\nSection",
		"content/section/child.md": "+++\ntitle = 'Child'\n+++\nChild",
	})
	result, err := generator.Build(context.Background(), site)
	if err != nil {
		t.Fatal(err)
	}
	if result.Pages != 6 || len(result.Warnings) != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	readPage := func(name string) string {
		t.Helper()
		bytes, err := os.ReadFile(filepath.Join(site, "public", name, "index.html"))
		if err != nil {
			t.Fatal(err)
		}
		return string(bytes)
	}
	home := readPage("")
	for _, expected := range []string{
		`<html lang="de">`, `<a class="identity" href="/company/">`,
		`<span class="site-title">&lt;Business&gt;</span>`,
		`content="Shared &#34;description&#34;"`, `href="/company/first/"`,
		`href="/company/second/?a=1&amp;b=2#team"`, `>&lt;Second&gt;</a>`,
		`href="https://example.net/"`, `<p>&lt;Copyright&gt;</p>`,
	} {
		if !strings.Contains(home, expected) {
			t.Errorf("home does not contain %q:\n%s", expected, home)
		}
	}
	if strings.Index(home, ">First</a>") > strings.Index(home, ">&lt;Second&gt;</a>") {
		t.Fatal("navigation was reordered")
	}
	if strings.Contains(home, `<a href="https://publication.example.org`) ||
		!strings.Contains(home, `<link rel="canonical" href="https://publication.example.org/company/">`) {
		t.Fatal("canonical identity or hostname-independent browsing is incorrect")
	}
	landing := readPage("landing")
	if !strings.Contains(landing, `class="hero`) || strings.Contains(home, `class="hero`) {
		t.Fatal("page types did not select distinct templates")
	}
	if !strings.Contains(landing, `src="https://images.example.org/team.jpg" alt="&lt;Team&gt;"`) ||
		strings.Contains(home, "team.jpg") {
		t.Fatal("site image must default only on landing pages")
	}
	clear := readPage("clear")
	if strings.Contains(clear, "<img") || strings.Contains(clear, `name="description"`) {
		t.Fatal("explicitly cleared metadata was inherited")
	}
	child := readPage("section/child")
	if !strings.Contains(child, "Shared &#34;description&#34;") || strings.Contains(child, "Section only") {
		t.Fatal("child inherited directory metadata instead of site metadata")
	}
	if _, err := os.Lstat(filepath.Join(site, "public", "company")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("URL prefix became an output subdirectory: %v", err)
	}
}

func TestInvalidConfigurationAndPageTypesPreserveOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	tests := []struct{ name, configuration, fields string }{
		{"invalid config", "[site]\npath = '../outside/'", "title = 'Home'"},
		{"invalid language", "[site]\nlanguage = 'en_US'", "title = 'Home'"},
		{"malformed config", "[site", "title = 'Home'"},
		{"invalid list", "", "title = 'List'\ntype = 'list'\n[list]\norder = 'invalid'"},
		{"invalid draft", "", "draft = true\ntype = 'invalid'"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			site := createSite(t, map[string]string{
				"cress.toml":       test.configuration,
				"content/index.md": "+++\n" + test.fields + "\n+++\nBody",
				"public/keep.txt":  "previous output",
			})
			if _, err := generator.Build(context.Background(), site); err == nil {
				t.Fatal("invalid input accepted")
			}
			marker, err := os.ReadFile(filepath.Join(site, "public", "keep.txt"))
			if err != nil || string(marker) != "previous output" {
				t.Fatalf("previous output changed: %q, %v", marker, err)
			}
		})
	}
}

func TestBuildWarningsIncludeConfigurationAndPages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{
		"cress.toml":       "[site]\nunknown = 'warning'",
		"content/index.md": "+++\ntitle = 'Home'\nunknown = 'warning'\n+++\nBody",
	})
	result, err := generator.Build(context.Background(), site)
	if err != nil || result.Pages != 2 || len(result.Warnings) != 3 {
		t.Fatalf("result = %+v, error = %v", result, err)
	}
	if !strings.Contains(result.Warnings[0], "cress.toml:") ||
		!strings.Contains(result.Warnings[1], "content/index.md:") {
		t.Fatalf("warning sources missing: %v", result.Warnings)
	}
}
