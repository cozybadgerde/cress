package generator_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/theme"
)

func TestBuildInitializedTheme(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{
		"cress.toml":         "[site]\nhost = 'https://example.org'\npath = '/company/'\n[theme]\nname = 'demo'\n[brand]\ncolor = '#123aBc'\nfavicon = '/favicon.svg'\n[brand.logo]\nsrc = '/logo.svg'\nalt = '<Logo>'",
		"content/index.md":   "+++\ntitle = 'Home'\n+++\nBody",
		"content/landing.md": "+++\ntitle = 'Landing'\ntype = 'landing'\n+++",
		"content/list.md":    "+++\ntitle = 'List'\ntype = 'list'\n[list]\norder = 'title'\n+++",
	})
	if err := theme.Init(context.Background(), site, "demo"); err != nil {
		t.Fatal(err)
	}
	if err := theme.Init(context.Background(), site, "demo"); err == nil {
		t.Fatal("existing target overwritten")
	}
	result, err := generator.BuildVersion(context.Background(), site, "dev")
	if err != nil || len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0], "baseline") {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	data, err := os.ReadFile(filepath.Join(site, "public", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`href="/company/themes/demo/site.css"`, `name="theme-color" content="#123aBc"`,
		`rel="icon" href="/company/favicon.svg"`, `src="/company/logo.svg" alt="&lt;Logo&gt;"`,
	} {
		if !strings.Contains(string(data), want) {
			t.Errorf("missing %q in %s", want, data)
		}
	}
	for _, private := range []string{
		"themes/demo/theme.json", "themes/demo/templates/page.gohtml",
		"themes/demo/templates/404.gohtml", "themes/demo/locales/en.toml",
	} {
		if _, err := os.Lstat(filepath.Join(site, "public", private)); !os.IsNotExist(err) {
			t.Errorf("private file published: %s %v", private, err)
		}
	}
	css, err := os.ReadFile(filepath.Join(site, "public", "themes", "demo", "site.css"))
	if err != nil || len(css) == 0 {
		t.Fatalf("public CSS missing: %v", err)
	}
}

func TestThemeErrorsPreserveOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, kind := range []string{
		"missing", "manifest", "incompatible", "name", "template", "locale",
		"missing-404-labels", "empty-404-labels", "asset", "symlink", "parent-symlink",
	} {
		t.Run(kind, func(t *testing.T) {
			site := createSite(t, map[string]string{
				"cress.toml":       "[site]\nhost = 'https://example.org'\nlanguage = 'en-US'\n[theme]\nname = 'demo'",
				"content/index.md": "+++\ntitle = 'Home'\n+++",
				"public/keep.txt":  "previous",
			})
			if kind != "missing" {
				if err := theme.Init(context.Background(), site, "demo"); err != nil {
					t.Fatal(err)
				}
				write := func(name, value string) {
					t.Helper()
					if err := os.WriteFile(filepath.Join(site, "themes", "demo", name), []byte(value), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				switch kind {
				case "manifest":
					write("theme.json", `{"name":null}`)
				case "incompatible":
					write("theme.json", `{"name":"demo","version":"1.0.0","cress":{"minimum":"1.1.0","major":1}}`)
				case "name":
					write("theme.json", `{"name":"other","version":"1.0.0","cress":{"minimum":"1.0.0","major":1}}`)
				case "template":
					write("templates/list.gohtml", `{{missing}}`)
				case "locale":
					write("templates/page.gohtml", `{{label "more"}}`)
					write("locales/en-US.toml", "not_found_title = 'Page not found'\nnot_found_message = 'Missing'\n")
				case "missing-404-labels":
					write("locales/en-US.toml", "more = 'No 404 labels'\n")
				case "empty-404-labels":
					write("locales/en-US.toml", "not_found_title = ' '\nnot_found_message = ''\n")
				case "asset":
					write("templates/page.gohtml", `{{asset "../theme.json"}}`)
				case "symlink":
					if err := os.Symlink("theme.json", filepath.Join(site, "themes", "demo", "private-link")); err != nil {
						t.Fatal(err)
					}
				case "parent-symlink":
					if err := os.Rename(filepath.Join(site, "themes"), filepath.Join(site, "real-themes")); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink("real-themes", filepath.Join(site, "themes")); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err := generator.BuildVersion(context.Background(), site, "1.0.0"); err == nil {
				t.Fatal("invalid local theme accepted")
			}
			previous, err := os.ReadFile(filepath.Join(site, "public", "keep.txt"))
			if err != nil || string(previous) != "previous" {
				t.Fatalf("previous output changed: %s %v", previous, err)
			}
		})
	}
}

func TestExactThemeLabelsAndAssetCollisions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, collision := range []bool{false, true} {
		site := createSite(t, map[string]string{
			"cress.toml":       "[site]\nhost = 'https://example.org'\n[theme]\nname = 'demo'",
			"content/index.md": "+++\ntitle = 'Home'\n+++",
		})
		if err := theme.Init(context.Background(), site, "demo"); err != nil {
			t.Fatal(err)
		}
		for name, data := range map[string]string{
			"locales/en.toml":       "more = '<More>'\nhome = 'Home'\nnot_found_title = 'Page not found'\nnot_found_message = 'Missing'\n",
			"templates/page.gohtml": `<p>{{label "more"}}</p>`,
		} {
			if err := os.WriteFile(filepath.Join(site, "themes", "demo", name), []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if collision {
			if err := os.MkdirAll(filepath.Join(site, "static", "themes", "demo"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(site, "static", "themes", "demo", "site.css"), []byte("other"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		_, err := generator.BuildVersion(context.Background(), site, "1.0.0")
		if collision {
			if err == nil || !strings.Contains(err.Error(), "collision") || !strings.Contains(err.Error(), "theme demo/public/site.css") {
				t.Fatalf("collision=%v", err)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(site, "public", "index.html"))
			if err != nil || string(data) != "<p>&lt;More&gt;</p>" {
				t.Fatalf("label not escaped: %s %v", data, err)
			}
		}
	}
}
