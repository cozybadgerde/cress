package content_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/content"
)

func TestLoadConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, kind := range []string{"missing", "valid", "malformed", "directory", "symlink", "dangling symlink", "unreadable"} {
		t.Run(kind, func(t *testing.T) {
			site := t.TempDir()
			filename := filepath.Join(site, "cress.toml")
			switch kind {
			case "valid", "unreadable":
				if err := os.WriteFile(filename, []byte("[site]\ntitle = 'Configured'"), 0o600); err != nil {
					t.Fatal(err)
				}
				if kind == "unreadable" {
					if os.Geteuid() == 0 {
						t.Skip("root can read mode-restricted files")
					}
					if err := os.Chmod(filename, 0); err != nil {
						t.Fatal(err)
					}
				}
			case "malformed":
				if err := os.WriteFile(filename, []byte("[site"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(filename, 0o700); err != nil {
					t.Fatal(err)
				}
			case "symlink", "dangling symlink":
				target := filepath.Join(t.TempDir(), "target.toml")
				if kind == "symlink" {
					if err := os.WriteFile(target, []byte("[site]\ntitle = 'Do not follow'"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(target, filename); err != nil {
					t.Fatal(err)
				}
			}
			config, warnings, err := content.LoadConfig(context.Background(), site)
			switch kind {
			case "missing", "valid":
				if err != nil || len(warnings) != 0 || config.Site.Language != "en" {
					t.Fatalf("config = %+v, warnings = %v, error = %v", config, warnings, err)
				}
				if kind == "valid" && config.Site.Title != "Configured" {
					t.Fatalf("configuration was not loaded: %+v", config)
				}
			default:
				if err == nil || !strings.Contains(err.Error(), "cress.toml") {
					t.Fatalf("error = %v, want explicit configuration failure", err)
				}
			}
		})
	}
}

func TestLoadConfigCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := content.LoadConfig(ctx, t.TempDir()); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want cancellation", err)
	}
}
