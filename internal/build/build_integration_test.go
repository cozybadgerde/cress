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

	assertMarkers(t, "index.html", readFile(t, filepath.Join(out, "index.html")), []marker{
		{"<title>Welcome · My cozy site</title>", "the composed title"},
		{"--accent: #9cb43b", "the injected accent color"},
		{`<link rel="icon" href="/favicon.png"`, "the default favicon link"},
		{`class="site-logo"`, "the nav logo"},
		{`href="/about.html"`, "the nav link to about"},
		{`class="footer-nav"`, "the footer nav"},
		{`href="/imprint.html"`, "the footer nav link to imprint"},
		{"Fresh little sites, fast.", "the rendered body content"},
	})

	assertMarkers(t, "about.html", readFile(t, filepath.Join(out, "about.html")), []marker{
		{`aria-current="page"`, "its own nav entry marked active"},
		{`<code class="language-go">`, "the fenced code block's language class"},
	})
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

// marker is one substring a rendered document must contain, with a
// human-readable name for the failure message.
type marker struct{ substr, desc string }

// assertMarkers checks every marker against doc, reporting each miss
// separately so one failure does not mask the others. The document is dumped
// once at the end when anything is missing, since the site it was built from
// lives in a t.TempDir() that is gone by the time the failure is read.
func assertMarkers(t *testing.T, name, doc string, markers []marker) {
	t.Helper()
	missing := false
	for _, m := range markers {
		if !strings.Contains(doc, m.substr) {
			t.Errorf("%s is missing %s (%q)", name, m.desc, m.substr)
			missing = true
		}
	}
	if missing {
		t.Logf("%s was:\n%s", name, doc)
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
