package generator_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/cozybadgerde/cress/internal/generator"
)

func TestBuildAssetsAndRelativeURLs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, layout := range []struct{ source, assets string }{
		{"about.md", "about.assets"},
		{"about/index.md", "about/index.assets"},
	} {
		t.Run(layout.source, func(t *testing.T) {
			owned := filepath.Base(layout.assets)
			site := createSite(t, map[string]string{
				"cress.toml": "[site]\npath = '/company/'\n",
				"content/index.md": "+++\ntitle = 'Home'\n+++\n" +
					"[About](" + layout.source + "?view=full#team)",
				"content/" + layout.source: "+++\ntitle = 'About'\n[image]\nsrc = '" + owned + "/nested/team.jpg'\nalt = 'Team'\n" +
					"caption = 'Photo by [Alex](" + owned + "/credits.txt)'\n+++\n" +
					"![Team](" + owned + `/nested/team.jpg "Hover"){caption="Photo by [Alex](` + owned + `/credits.txt)"}` + "\n\n" +
					"[Download](" + owned + "/manual.md) [Image](" + owned + "/nested/team.jpg?size=2#view)\n\n" +
					"```go caption=\"[Credit](" + owned + "/credits.txt)\"\nvar value = \"<&>\"\n```\n",
				"content/" + layout.assets + "/nested/team.jpg": "\x00\xff\x01unchanged\r\n",
				"content/" + layout.assets + "/manual.md":       "An opaque Markdown asset, no title required",
				"content/" + layout.assets + "/credits.txt":     "Alex",
				"content/draft.md":                              "+++\ndraft = true\n+++\n![Missing](missing.jpg)",
				"content/draft.assets/secret.jpg":               "private draft asset",
				"content/private.jpg":                           "private file",
				"static/images/logo.svg":                        "<svg>unchanged</svg>",
				"static/data/opaque.md":                         "not a content page",
			})
			result, err := generator.Build(context.Background(), site)
			if err != nil {
				t.Fatal(err)
			}
			if result.Pages != 2 || len(result.Warnings) != 1 ||
				!strings.Contains(result.Warnings[0], "site.host") {
				t.Fatalf("unexpected result: %+v", result)
			}
			for source, destination := range map[string]string{
				"content/" + layout.assets + "/nested/team.jpg": "about/nested/team.jpg",
				"content/" + layout.assets + "/manual.md":       "about/manual.md",
				"content/" + layout.assets + "/credits.txt":     "about/credits.txt",
				"static/images/logo.svg":                        "images/logo.svg",
				"static/data/opaque.md":                         "data/opaque.md",
			} {
				original, err := os.ReadFile(filepath.Join(site, source))
				if err != nil {
					t.Fatal(err)
				}
				published, err := os.ReadFile(filepath.Join(site, "public", destination))
				if err != nil || !bytes.Equal(original, published) {
					t.Fatalf("asset changed: %s -> %s: %q, %v", source, destination, published, err)
				}
			}
			for _, private := range []string{"draft/index.html", "draft/secret.jpg", "private.jpg", layout.assets} {
				if _, err := os.Lstat(filepath.Join(site, "public", private)); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("private or source-only path published: %s: %v", private, err)
				}
			}
			home, err := os.ReadFile(filepath.Join(site, "public", "index.html"))
			if err != nil || !strings.Contains(string(home), `href="/company/about/?view=full#team"`) {
				t.Fatalf("source page URL not resolved: %s, %v", home, err)
			}
			about, err := os.ReadFile(filepath.Join(site, "public", "about", "index.html"))
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{
				`src="/company/about/nested/team.jpg" alt="Team"`,
				`<figcaption>Photo by <a href="/company/about/credits.txt">Alex</a></figcaption>`,
				`href="/company/about/manual.md"`,
				`href="/company/about/nested/team.jpg?size=2#view"`,
				`<figcaption><a href="/company/about/credits.txt">Credit</a></figcaption>`,
				`<span class="kd">var</span>`,
			} {
				if !strings.Contains(string(about), want) {
					t.Errorf("missing %q in %s", want, about)
				}
			}
		})
	}
}

func TestBuildSourceReferenceErrorsPreserveOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, body := range []string{
		"[Missing](missing.md)",
		"[Draft](draft.md)",
		"![Draft asset](draft.assets/secret.jpg)",
		"![Missing asset](index.assets/missing.jpg)",
		"[Private](private.txt)",
		"![Page](about.md)",
		"[Escape](../../outside.jpg)",
		`![Photo](/photo.jpg){caption="[Missing](missing.md)"}`,
		"```go caption=\"[Draft](draft.md)\"\nvar code = 1\n```",
	} {
		t.Run(body, func(t *testing.T) {
			site := createSite(t, map[string]string{
				"content/index.md":                "+++\ntitle = 'Home'\n+++\n" + body,
				"content/about.md":                "+++\ntitle = 'About'\n+++",
				"content/draft.md":                "+++\ndraft = true\n+++",
				"content/draft.assets/secret.jpg": "secret",
				"content/private.txt":             "private",
				"public/index.html":               "previous output",
			})
			_, err := generator.Build(context.Background(), site)
			if err == nil || !strings.Contains(err.Error(), "content/index.md") {
				t.Fatalf("error = %v, want referring source", err)
			}
			previous, err := os.ReadFile(filepath.Join(site, "public", "index.html"))
			if err != nil || string(previous) != "previous output" {
				t.Fatalf("previous output changed: %q, %v", previous, err)
			}
		})
	}
}

func TestBuildAssetCollisionsAndOrphans(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, test := range []struct {
		name    string
		sources map[string]string
		want    []string
	}{
		{
			name:    "orphan",
			sources: map[string]string{"content/orphan.assets/file.txt": "unowned"},
			want:    []string{"content/orphan.assets", "orphan"},
		},
		{
			name:    "asset page",
			sources: map[string]string{"content/about.assets/index.html": "collides"},
			want:    []string{"collision", "content/about.md", "content/about.assets/index.html"},
		},
		{
			name:    "static page",
			sources: map[string]string{"static/about/index.html": "collides"},
			want:    []string{"collision", "content/about.md", "static/about/index.html"},
		},
		{
			name: "static owned",
			sources: map[string]string{
				"static/about/team.jpg": "shared", "content/about.assets/team.jpg": "owned",
			},
			want: []string{"collision", "static/about/team.jpg", "content/about.assets/team.jpg"},
		},
		{
			name: "case alias",
			sources: map[string]string{
				"static/about/Team.jpg": "shared", "content/about.assets/team.jpg": "owned",
			},
			want: []string{"collision", "static/about/Team.jpg", "content/about.assets/team.jpg"},
		},
		{
			name: "unicode alias",
			sources: map[string]string{
				"static/about/é.jpg": "shared", "content/about.assets/e\u0301.jpg": "owned",
			},
			want: []string{"collision", "static/about/é.jpg", "content/about.assets/e\u0301.jpg"},
		},
		{
			name:    "static file directory",
			sources: map[string]string{"static/about": "file"},
			want:    []string{"collision", "static/about", "content/about.md"},
		},
		{
			name: "owned file directory",
			sources: map[string]string{
				"content/about.assets/team": "file", "static/about/team/photo.jpg": "nested",
			},
			want: []string{"collision", "content/about.assets/team", "static/about/team/photo.jpg"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.sources["content/about.md"] = "+++\ntitle = 'About'\n+++"
			site := createSite(t, test.sources)
			_, err := generator.Build(context.Background(), site)
			for _, want := range test.want {
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("error = %v, want %q", err, want)
				}
			}
			if _, err := os.Lstat(filepath.Join(site, "public")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("output created before validation: %v", err)
			}
		})
	}
}

func TestBuildRefusesAssetSymlinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, name := range []string{"static", "static/link", "content/about.assets", "content/about.assets/link",
		"content/draft.assets/link"} {
		t.Run(name, func(t *testing.T) {
			site := createSite(t, map[string]string{
				"content/about.md": "+++\ntitle = 'About'\n+++",
				"content/draft.md": "+++\ndraft = true\n+++",
			})
			target := t.TempDir()
			if err := os.MkdirAll(filepath.Dir(filepath.Join(site, name)), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, filepath.Join(site, name)); err != nil {
				t.Fatal(err)
			}
			if _, err := generator.Build(context.Background(), site); err == nil {
				t.Fatal("asset symlink accepted")
			}
			if _, err := os.Lstat(filepath.Join(site, "public")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("output created despite symlink: %v", err)
			}
		})
	}
}

func TestBuildRefusesSpecialAssets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, name := range []string{"static/pipe", "content/about.assets/pipe"} {
		t.Run(name, func(t *testing.T) {
			site := createSite(t, map[string]string{"content/about.md": "+++\ntitle = 'About'\n+++"})
			filename := filepath.Join(site, name)
			if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := syscall.Mkfifo(filename, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := generator.Build(context.Background(), site); err == nil || !strings.Contains(err.Error(), name) {
				t.Fatalf("special asset accepted or source omitted: %v", err)
			}
		})
	}
	site := createSite(t, map[string]string{
		"content/about.md": "+++\ntitle = 'About'\n+++",
		"static":           "not a directory",
	})
	if _, err := generator.Build(context.Background(), site); err == nil || !strings.Contains(err.Error(), "static") {
		t.Fatalf("non-directory static accepted: %v", err)
	}
}

func TestBuildDisablesHighlighting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{
		"cress.toml":      "[markdown]\nhighlight = false",
		"content/page.md": "+++\ntitle = 'Page'\n+++\n```go caption=\"Example\"\nvar text = \"<&>\"\n```",
	})
	if _, err := generator.Build(context.Background(), site); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(filepath.Join(site, "public", "page", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), `<code class="language-go">var text = &#34;&lt;&amp;&gt;&#34;`) ||
		strings.Contains(string(output), "<span") || !strings.Contains(string(output), "<figcaption>Example</figcaption>") {
		t.Fatalf("highlighting switch or code caption failed: %s", output)
	}
}

func TestBuildTitleImageSourceErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, fields := range []string{
		"src = 'missing.jpg'\nalt = ''",
		"src = 'draft.assets/secret.jpg'\nalt = ''",
		"src = 'page.md'\nalt = ''",
		"src = '/team.jpg'\nalt = ''\ncaption = '[Missing](missing.md)'",
		"src = '/team.jpg'\nalt = ''\ncaption = '**Invalid**'",
	} {
		t.Run(fields, func(t *testing.T) {
			site := createSite(t, map[string]string{
				"content/page.md":                 "+++\ntitle = 'Page'\n[image]\n" + fields + "\n+++",
				"content/draft.md":                "+++\ndraft = true\n+++",
				"content/draft.assets/secret.jpg": "private",
				"public/index.html":               "previous output",
			})
			if _, err := generator.Build(context.Background(), site); err == nil ||
				!strings.Contains(err.Error(), "content/page.md: image.") {
				t.Fatalf("image error lacks context: %v", err)
			}
			previous, err := os.ReadFile(filepath.Join(site, "public", "index.html"))
			if err != nil || string(previous) != "previous output" {
				t.Fatalf("previous output changed: %s, %v", previous, err)
			}
		})
	}
}
