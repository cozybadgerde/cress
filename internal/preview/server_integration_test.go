package preview

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRootReplacementPreservesOwnership(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, outcome := range []string{"missing", "success", "old root closed"} {
		t.Run(outcome, func(t *testing.T) {
			old, err := os.OpenRoot(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := old.Close(); err != nil {
					t.Error(err)
				}
			})
			server := &pageServer{root: old}
			t.Cleanup(func() {
				if err := server.root.Close(); err != nil {
					t.Error(err)
				}
			})
			directory := t.TempDir()
			if outcome == "missing" {
				directory = filepath.Join(directory, "absent")
			}
			if outcome == "old root closed" {
				if err := old.Close(); err != nil {
					t.Fatal(err)
				}
			}
			server.mu.Lock()
			err = server.replaceRoot(directory)
			server.mu.Unlock()
			if outcome == "missing" {
				if !errors.Is(err, os.ErrNotExist) || server.root != old {
					t.Fatalf("failed replacement lost original root: %v", err)
				}
				if _, err := old.Stat("."); err != nil {
					t.Fatalf("original root no longer usable: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := server.root.Stat("."); err != nil {
				t.Fatalf("replacement root unusable: %v", err)
			}
			if _, err := old.Stat("."); !errors.Is(err, os.ErrClosed) {
				t.Fatalf("old root not closed: %v", err)
			}
		})
	}
}
