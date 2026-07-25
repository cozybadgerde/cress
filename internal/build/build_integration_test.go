package build_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/build"
	"github.com/cozybadgerde/cress/internal/scaffold"
)

func TestBuild_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if res.Pages != 5 {
		t.Errorf("rendered %d pages, want 5 (index, about, styleguide, imprint, privacy)", res.Pages)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", res.Warnings)
	}

	out := filepath.Join(root, build.OutputDir)
	// style.css comes from the theme; favicon.png and logo.png from the site's
	// static/ tree that `cress init` scaffolds.
	for _, name := range []string{"index.html", "about.html", "style.css", "favicon.png", "logo.png"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("missing output %s: %v", name, err)
		}
	}

	index := readFile(t, filepath.Join(out, "index.html"))
	if !strings.Contains(index, "<title>Welcome · My cozy site</title>") {
		t.Errorf("index.html missing composed title:\n%s", index)
	}
	if !strings.Contains(index, "--accent: #9cb43b") {
		t.Error("index.html missing injected accent color")
	}
	if !strings.Contains(index, `<link rel="icon" href="/favicon.png"`) {
		t.Error("index.html missing default favicon link")
	}
	if !strings.Contains(index, `class="site-logo"`) {
		t.Error("index.html missing nav logo")
	}
	if !strings.Contains(index, `href="/about.html"`) {
		t.Error("index.html missing nav link to about")
	}
	if !strings.Contains(index, `class="footer-nav"`) || !strings.Contains(index, `href="/imprint.html"`) {
		t.Errorf("index.html missing the footer nav:\n%s", index)
	}
	if !strings.Contains(index, "Fresh little sites, fast.") {
		t.Error("index.html missing rendered body content")
	}

	about := readFile(t, filepath.Join(out, "about.html"))
	if !strings.Contains(about, `aria-current="page"`) {
		t.Error("about.html should mark its own nav entry active")
	}
	if !strings.Contains(about, `<code class="language-go">`) {
		t.Error("about.html should render the fenced code block with a language class")
	}
}

func TestBuild_draftsAndStatic_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	writeSiteFile(t, filepath.Join(root, "content", "secret.md"), "---\ntitle: Secret\ndraft: true\n---\nhidden\n")
	writeSiteFile(t, filepath.Join(root, "static", "robots.txt"), "User-agent: *\n")

	// Default build: drafts excluded, site static copied.
	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if res.Pages != 5 {
		t.Errorf("rendered %d pages, want 5 (draft excluded)", res.Pages)
	}
	out := filepath.Join(root, build.OutputDir)
	if _, err := os.Stat(filepath.Join(out, "secret.html")); !os.IsNotExist(err) {
		t.Errorf("draft page should not be rendered (stat err = %v)", err)
	}
	if _, err := os.Stat(filepath.Join(out, "robots.txt")); err != nil {
		t.Errorf("site static file should be copied: %v", err)
	}

	// With drafts enabled the page appears.
	res, err = build.Build(build.Options{Root: root, Drafts: true})
	if err != nil {
		t.Fatalf("build --drafts: %v", err)
	}
	if res.Pages != 6 {
		t.Errorf("rendered %d pages with drafts, want 6", res.Pages)
	}
	if _, err := os.Stat(filepath.Join(out, "secret.html")); err != nil {
		t.Errorf("draft page should render with --drafts: %v", err)
	}
}

func TestBuild_refusesSiteRootOutput(t *testing.T) {
	root := t.TempDir()
	if err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	if _, err := build.Build(build.Options{Root: root, Output: "."}); err == nil {
		t.Fatal("expected an error when output is the site root")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path) // #nosec G304 -- test-controlled path
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

func writeSiteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
