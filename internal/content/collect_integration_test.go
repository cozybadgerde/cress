package content_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/content"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestCollect_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "index.md"), "---\ntitle: Home\n---\n# Home\n")
	writeFile(t, filepath.Join(root, "about.md"), "# About us\n")
	writeFile(t, filepath.Join(root, "guide", "index.md"), "# Guide\n")
	writeFile(t, filepath.Join(root, "guide", "setup.md"), "---\ntitle: Setup\ndraft: true\n---\nbody\n")
	writeFile(t, filepath.Join(root, "notes.txt"), "ignored, not markdown\n")

	pages, err := content.Collect(root)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(pages) != 4 {
		t.Fatalf("collected %d pages, want 4 (non-.md ignored)", len(pages))
	}

	// WalkDir yields lexical order, so the slice is deterministic.
	want := []struct {
		source, out, url, title string
		draft                   bool
	}{
		{"about.md", "about.html", "/about.html", "About us", false},
		{"guide/index.md", "guide/index.html", "/guide/", "Guide", false},
		{"guide/setup.md", "guide/setup.html", "/guide/setup.html", "Setup", true},
		{"index.md", "index.html", "/", "Home", false},
	}
	for i, w := range want {
		p := pages[i]
		if p.SourcePath != w.source {
			t.Errorf("pages[%d].SourcePath = %q, want %q", i, p.SourcePath, w.source)
		}
		if p.OutputPath != w.out {
			t.Errorf("pages[%d].OutputPath = %q, want %q", i, p.OutputPath, w.out)
		}
		if p.URL != w.url {
			t.Errorf("pages[%d].URL = %q, want %q", i, p.URL, w.url)
		}
		if p.Title != w.title {
			t.Errorf("pages[%d].Title = %q, want %q", i, p.Title, w.title)
		}
		if p.Draft != w.draft {
			t.Errorf("pages[%d].Draft = %v, want %v", i, p.Draft, w.draft)
		}
	}
}

// A page's layout is read verbatim and never checked here: whether a theme
// defines one is the builder's question, so this package reports what the
// author wrote and nothing more.
func TestCollect_layout_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "blank.md"), "---\nlayout: \"   \"\n---\nbody\n")
	writeFile(t, filepath.Join(root, "named.md"), "---\nlayout: landing\n---\nbody\n")
	writeFile(t, filepath.Join(root, "numeric.md"), "---\nlayout: 42\n---\nbody\n")
	writeFile(t, filepath.Join(root, "padded.md"), "---\nlayout: \" wide \"\n---\nbody\n")
	writeFile(t, filepath.Join(root, "unknown.md"), "---\nlayout: no-theme-has-this\n---\nbody\n")
	writeFile(t, filepath.Join(root, "unset.md"), "---\ntitle: Plain\n---\nbody\n")

	pages, err := content.Collect(root)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	// WalkDir yields lexical order, so the slice is deterministic.
	want := map[string]string{
		"blank.md":   "",
		"named.md":   "landing",
		"numeric.md": "",
		"padded.md":  "wide",
		"unknown.md": "no-theme-has-this",
		"unset.md":   "",
	}
	if len(pages) != len(want) {
		t.Fatalf("collected %d pages, want %d", len(pages), len(want))
	}
	for _, p := range pages {
		if got := p.Layout; got != want[p.SourcePath] {
			t.Errorf("%s: Layout = %q, want %q", p.SourcePath, got, want[p.SourcePath])
		}
	}
}

func TestCollectMissingRoot_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	_, err := content.Collect(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected an error for a missing content root")
	}
}

// A symlink is not a page. The walk does not follow one, but reading it would,
// so a link left to resolve would publish whatever it points at: in a build run
// over content somebody else can add to, that is any file the process can read.
func TestCollectSkipsSymlinks_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("SECRET"), 0o600); err != nil {
		t.Fatalf("writing secret: %v", err)
	}

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "real.md"), "# Real\n")
	if err := os.Symlink(outside, filepath.Join(root, "leak.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	// A link to a file inside the tree is skipped too: the rule is the file type,
	// not where it points, because deciding that needs the target resolved.
	if err := os.Symlink(filepath.Join(root, "real.md"), filepath.Join(root, "alias.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	pages, err := content.Collect(root)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("collected %d page(s), want only the regular file", len(pages))
	}
	if pages[0].SourcePath != "real.md" {
		t.Errorf("collected %q, want real.md", pages[0].SourcePath)
	}
	for _, page := range pages {
		if strings.Contains(string(page.Body), "SECRET") {
			t.Errorf("a symlink target reached the page body: %s", page.Body)
		}
	}
}
