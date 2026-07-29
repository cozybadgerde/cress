package scaffold_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/scaffold"
)

func TestCreate_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dir := t.TempDir()
	if _, err := scaffold.Create(dir, false); err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, rel := range []string{
		"cress.toml",
		filepath.Join("content", "index.md"),
		filepath.Join("content", "about.md"),
		filepath.Join("content", "styleguide.md"),
		filepath.Join("content", "imprint.md"),
		filepath.Join("content", "privacy.md"),
		filepath.Join("static", "logo.svg"),
		filepath.Join("static", "favicon.png"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("missing scaffolded file %s: %v", rel, err)
		}
	}

	cfg, err := os.ReadFile(filepath.Join(dir, "cress.toml")) // #nosec G304 -- test path
	if err != nil {
		t.Fatalf("read cress.toml: %v", err)
	}
	for _, table := range []string{"[nav.main]", "[nav.footer]"} {
		if !strings.Contains(string(cfg), table) {
			t.Errorf("scaffolded cress.toml missing %s table:\n%s", table, cfg)
		}
	}
}

func TestCreateRefusesNonEmpty_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, err := scaffold.Create(dir, false); !errors.Is(err, scaffold.ErrExists) {
		t.Fatalf("Create into non-empty dir: err = %v, want ErrExists", err)
	}

	skipped, err := scaffold.Create(dir, true)
	if err != nil {
		t.Fatalf("Create with force: %v", err)
	}
	if len(skipped) != 0 {
		t.Errorf("nothing of the starter site was there yet, but skipped = %v", skipped)
	}
	if _, err := os.Stat(filepath.Join(dir, "cress.toml")); err != nil {
		t.Errorf("force should have written the site: %v", err)
	}
}

func TestCreateForceKeepsExistingFiles_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// A site someone has already worked on: two starter files carry their
	// content now, and a third is missing entirely.
	dir := t.TempDir()
	mine := map[string]string{
		"cress.toml":                         "title = \"My actual site\"\n",
		filepath.Join("content", "index.md"): "# Two years of work\n",
		filepath.Join("content", "about.md"): "# About me\n",
	}
	for rel, body := range mine {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("seed dir for %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("seed %s: %v", rel, err)
		}
	}

	skipped, err := scaffold.Create(dir, true)
	if err != nil {
		t.Fatalf("Create with force: %v", err)
	}

	for rel, want := range mine {
		got, err := os.ReadFile(filepath.Join(dir, rel)) // #nosec G304 -- test path
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if string(got) != want {
			t.Errorf("--force overwrote %s:\n got %q\nwant %q", rel, got, want)
		}
	}
	if len(skipped) != len(mine) {
		t.Errorf("skipped = %v, want the %d pre-existing files reported", skipped, len(mine))
	}
	// The files that were not there still get written.
	if _, err := os.Stat(filepath.Join(dir, "content", "privacy.md")); err != nil {
		t.Errorf("force should still write the missing starter files: %v", err)
	}
}

func TestCreateIgnoresDotEntries_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// `git init` before `cress init` is the natural way to start a site, and it
	// leaves a .git behind. That must not require --force.
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatalf("seed .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("public/\n"), 0o644); err != nil {
		t.Fatalf("seed .gitignore: %v", err)
	}

	if _, err := scaffold.Create(dir, false); err != nil {
		t.Fatalf("Create into a dir holding only dot-entries: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cress.toml")); err != nil {
		t.Errorf("the site should have been scaffolded: %v", err)
	}
}
