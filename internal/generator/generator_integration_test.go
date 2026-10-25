package generator_test

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/output"
)

func createSite(t *testing.T, sources map[string]string) string {
	t.Helper()
	site := t.TempDir()
	if err := os.Mkdir(filepath.Join(site, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range sources {
		filename := filepath.Join(site, name)
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return site
}

func TestBuildSite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{
		"content/index.md":       "+++\ntitle = '<Home>'\nextra = 'warn'\n+++\nWelcome",
		"content/about.md":       "+++\ndraft = true\n+++\nPrivate duplicate",
		"content/about/index.md": "+++\ntitle = 'About'\n+++\n**About us**",
		"content/blog/first.md":  "+++\ntitle = 'First'\n+++\n<script>trusted()</script>",
		"content/draft.md":       "+++\ndraft = true\n+++\nPrivate body",
		"content/ignored.MD":     "Not a lowercase Markdown file",
		"content/photo.jpg":      "Private outside an asset directory",
		"cress.toml":             "[theme.custom]\nprivate = 'configuration'",
		"themes/private.gohtml":  "private template",
		"static/shared.txt":      "shared static file",
	})
	result, err := generator.Build(context.Background(), site)
	if err != nil {
		t.Fatal(err)
	}
	if result.Pages != 3 || len(result.Warnings) != 2 ||
		!strings.Contains(result.Warnings[0], `unknown frontmatter field "extra"`) {
		t.Fatalf("unexpected result: %+v", result)
	}
	var published []string
	err = filepath.WalkDir(filepath.Join(site, "public"), func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			relative, err := filepath.Rel(filepath.Join(site, "public"), name)
			if err != nil {
				return err
			}
			published = append(published, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(published)
	if !slices.Equal(published, []string{".cress-output.json", "about/index.html", "blog/first/index.html", "index.html", "robots.txt", "shared.txt", "themes/poppy/site.css"}) {
		t.Fatalf("unexpected published files: %v", published)
	}
	home, err := os.ReadFile(filepath.Join(site, "public", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(home), "<h1>&lt;Home&gt;</h1>") ||
		!strings.Contains(string(home), "<p>Welcome</p>") {
		t.Fatalf("unexpected home output:\n%s", home)
	}
	first, err := os.ReadFile(filepath.Join(site, "public", "blog", "first", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), "<script>trusted()</script>") {
		t.Fatal("trusted HTML was not preserved")
	}
}

func TestBuildErrorsDoNotCreateOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, source := range []string{
		"Body without a title",
		"+++\ntitle = [\n+++",
		"+++\ntitle = 1\ndraft = true\n+++",
	} {
		t.Run(source, func(t *testing.T) {
			site := createSite(t, map[string]string{
				"content/a-valid.md": "+++\ntitle = 'Valid'\n+++\nPrepared first",
				"content/z-error.md": source,
			})
			if _, err := generator.Build(context.Background(), site); err == nil {
				t.Fatal("invalid source accepted")
			}
			if _, err := os.Lstat(filepath.Join(site, "public")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("output created despite invalid source: %v", err)
			}
		})
	}
}

func TestBuildErrorsPreservePreviousOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{
		"content/index.md":  "+++\ntitle = 123\n+++",
		"public/index.html": "previous successful output",
	})
	if _, err := generator.Build(context.Background(), site); err == nil {
		t.Fatal("invalid input accepted")
	}
	previous, err := os.ReadFile(filepath.Join(site, "public", "index.html"))
	if err != nil || string(previous) != "previous successful output" {
		t.Fatalf("previous output changed: %q, %v", previous, err)
	}
}

func TestBuildCollisions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, pair := range [][2]string{
		{"about.md", "about/index.md"},
		{"About.md", "about/index.md"},
		{"café.md", "cafe\u0301/index.md"},
		{"index.md", "index.html.md"},
	} {
		t.Run(strings.Join(pair[:], " and "), func(t *testing.T) {
			site := createSite(t, map[string]string{
				"content/" + pair[0]: "+++\ntitle = 'First'\n+++",
				"content/" + pair[1]: "+++\ntitle = 'Second'\n+++",
			})
			_, err := generator.Build(context.Background(), site)
			if err == nil || !strings.Contains(err.Error(), "collision") ||
				!strings.Contains(err.Error(), "content/"+pair[0]) ||
				!strings.Contains(err.Error(), "content/"+pair[1]) {
				t.Fatalf("error = %v, want collision with both sources", err)
			}
			if _, err := os.Lstat(filepath.Join(site, "public")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("output created despite collision: %v", err)
			}
		})
	}
}

func TestBuildRefusesContentSymlinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, kind := range []string{"root", "file", "directory", "dangling"} {
		t.Run(kind, func(t *testing.T) {
			site := t.TempDir()
			outside := createSite(t, map[string]string{"content/secret.md": "+++\ntitle = 'Secret'\n+++"})
			if kind == "root" {
				if err := os.Symlink(filepath.Join(outside, "content"), filepath.Join(site, "content")); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Mkdir(filepath.Join(site, "content"), 0o755); err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(outside, "content", "secret.md")
				switch kind {
				case "directory":
					target = filepath.Join(outside, "content")
				case "dangling":
					target = filepath.Join(outside, "absent")
				}
				if err := os.Symlink(target, filepath.Join(site, "content", "link.md")); err != nil {
					t.Fatal(err)
				}
			}
			_, err := generator.Build(context.Background(), site)
			if err == nil || !strings.Contains(err.Error(), "symlink") {
				t.Fatalf("error = %v, want explicit symlink refusal", err)
			}
			if _, err := os.Lstat(filepath.Join(site, "public")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("output created despite unsafe content: %v", err)
			}
		})
	}
}

func TestBuildMissingContentAndCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	if _, err := generator.Build(context.Background(), t.TempDir()); err == nil {
		t.Fatal("missing content directory accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	site := createSite(t, nil)
	if _, err := generator.Build(ctx, site); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want cancellation", err)
	}
	if _, err := os.Lstat(filepath.Join(site, "public")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("output created despite cancellation: %v", err)
	}
}

func TestBuildWithNoPublishedPages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, drafts := range []bool{false, true} {
		name := "empty content"
		var sources map[string]string
		if drafts {
			name = "draft-only content"
			sources = map[string]string{"content/private.md": "+++\ndraft = true\n+++\nPrivate"}
		}
		t.Run(name, func(t *testing.T) {
			site := createSite(t, sources)
			result, err := generator.Build(context.Background(), site)
			if err != nil || result.Pages != 0 {
				t.Fatalf("result = %+v, error = %v", result, err)
			}
			entries, err := os.ReadDir(filepath.Join(site, "public"))
			if err != nil || len(entries) != 3 ||
				entries[0].Name() != ".cress-output.json" || entries[1].Name() != "robots.txt" || entries[2].Name() != "themes" {
				t.Fatalf("unpublished output = %v, error = %v", entries, err)
			}

		})
	}
}

func TestRebuildPreservesOwnedOutputOnInvalidInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{"content/index.md": "+++\ntitle = 'Home'\n+++\nOriginal"})
	if _, err := generator.Build(context.Background(), site); err != nil {
		t.Fatal(err)
	}
	previous, err := os.ReadFile(filepath.Join(site, "public/index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(site, "content/index.md"), []byte("invalid without title"), 0o644); err != nil {
		t.Fatal(err)
	}
	writer, err := output.Acquire(context.Background(), site)
	if err != nil {
		t.Fatal(err)
	}
	_, buildErr := generator.Build(context.Background(), site)
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(buildErr, output.ErrBusy) {
		t.Fatalf("build did not lock before loading invalid inputs: %v", buildErr)
	}
	if _, err := generator.Build(context.Background(), site); err == nil || !strings.Contains(err.Error(), "title") {
		t.Fatalf("invalid input not reported after lock release: %v", err)
	}
	current, err := os.ReadFile(filepath.Join(site, "public/index.html"))
	if err != nil || !bytes.Equal(current, previous) {
		t.Fatalf("owned output changed on generation error: %q, %v", current, err)
	}
}
