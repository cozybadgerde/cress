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
	if err := scaffold.Create(dir, false); err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, rel := range []string{
		"cress.toml",
		filepath.Join("content", "index.md"),
		filepath.Join("content", "about.md"),
		filepath.Join("content", "styleguide.md"),
		filepath.Join("static", "logo.png"),
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
	if !strings.Contains(string(cfg), "[nav]") {
		t.Errorf("scaffolded cress.toml missing [nav] table:\n%s", cfg)
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

	if err := scaffold.Create(dir, false); !errors.Is(err, scaffold.ErrExists) {
		t.Fatalf("Create into non-empty dir: err = %v, want ErrExists", err)
	}

	if err := scaffold.Create(dir, true); err != nil {
		t.Fatalf("Create with force: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cress.toml")); err != nil {
		t.Errorf("force should have written the site: %v", err)
	}
}
