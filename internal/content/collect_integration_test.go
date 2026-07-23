package content_test

import (
	"os"
	"path/filepath"
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

func TestCollectMissingRoot_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	_, err := content.Collect(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected an error for a missing content root")
	}
}
