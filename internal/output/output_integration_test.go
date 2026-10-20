package output_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/output"
)

func TestWriteNewAndEmptyOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, existing := range []bool{false, true} {
		name := "new"
		if existing {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			site := t.TempDir()
			if existing {
				if err := os.Mkdir(filepath.Join(site, "public"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			files := []output.File{
				{Path: "index.html", Source: "content/index.md", Data: []byte("Home")},
				{Path: "about/index.html", Source: "content/about.md", Data: []byte("About")},
			}
			if err := output.Write(context.Background(), site, files); err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				data, err := os.ReadFile(filepath.Join(site, "public", file.Path))
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != string(file.Data) {
					t.Fatalf("%s = %q, want %q", file.Path, data, file.Data)
				}
			}
			if err := output.Write(context.Background(), site, files); err != nil {
				t.Fatalf("subsequent build failed: %v", err)
			}
		})
	}
}

func TestRefuseUnsafeOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, kind := range []string{"nonempty", "file", "symlink", "dangling symlink", "content symlink"} {
		t.Run(kind, func(t *testing.T) {
			site := t.TempDir()
			target := filepath.Join(site, "public")
			protected := filepath.Join(site, "protected")
			if err := os.Mkdir(protected, 0o755); err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(protected, "keep")
			if err := os.WriteFile(marker, []byte("unchanged"), 0o644); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "nonempty":
				if err := os.Mkdir(target, 0o755); err != nil {
					t.Fatal(err)
				}
				marker = filepath.Join(target, ".hidden")
				if err := os.WriteFile(marker, []byte("unchanged"), 0o644); err != nil {
					t.Fatal(err)
				}
			case "file":
				if err := os.WriteFile(target, []byte("unchanged"), 0o644); err != nil {
					t.Fatal(err)
				}
				marker = target
			case "symlink":
				if err := os.Symlink(protected, target); err != nil {
					t.Fatal(err)
				}
			case "dangling symlink":
				if err := os.Symlink(filepath.Join(site, "absent"), target); err != nil {
					t.Fatal(err)
				}
			case "content symlink":
				if err := os.Mkdir(filepath.Join(site, "content"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("content", target); err != nil {
					t.Fatal(err)
				}
			}
			files := []output.File{{Path: "index.html", Source: "content/index.md", Data: []byte("new")}}
			if err := output.Write(context.Background(), site, files); err == nil {
				t.Fatal("unsafe output accepted")
			}
			data, err := os.ReadFile(marker)
			if err != nil || string(data) != "unchanged" {
				t.Fatalf("protected content changed: %q, %v", data, err)
			}
			if _, err := os.Lstat(filepath.Join(protected, "index.html")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("unexpected output in protected directory: %v", err)
			}
			if kind == "content symlink" {
				entries, err := os.ReadDir(filepath.Join(site, "content"))
				if err != nil || len(entries) != 0 {
					t.Fatalf("output was written into content: %v, %v", entries, err)
				}
			}
		})
	}
}

func TestWriteValidationAndCancellationPreserveTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	files := []output.File{{Path: "../escape", Source: "content/index.md"}}
	if err := output.Write(context.Background(), site, files); err == nil {
		t.Fatal("unsafe destination accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := output.Write(ctx, site, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want cancellation", err)
	}
	if _, err := os.Lstat(filepath.Join(site, "public")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("output was created before validation or despite cancellation: %v", err)
	}
}

func TestWriteFilesystemFailureIsExplicit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	files := []output.File{{
		Path: strings.Repeat("x", 256) + "/index.html", Source: "content/oversized.md", Data: []byte("Body"),
	}}
	err := output.Write(context.Background(), site, files)
	if err == nil || !strings.Contains(err.Error(), "previous public unchanged") ||
		!strings.Contains(err.Error(), "content/oversized.md") {
		t.Fatalf("error = %v, want source and preservation guidance", err)
	}
}
