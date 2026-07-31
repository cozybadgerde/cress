package clean_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/clean"
)

// site writes a minimal cress site and returns its root. Only the config file
// matters: clean never reads it, but its presence is what marks the directory
// as a site.
func site(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "cress.toml"), []byte("[site]\ntitle = \"x\"\n"), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	return root
}

// write creates a file under root, making its parent directories as needed.
func write(t *testing.T, root, rel string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("writing %s: %v", rel, err)
	}
}

func TestPrepareCollects_integration(t *testing.T) {
	root := site(t)
	for _, rel := range []string{
		"public/index.html",
		"public/404.html",
		"public/guides/writing.html",
		"public/style.css",
	} {
		write(t, root, rel)
	}

	plan, err := clean.Prepare(clean.Options{Root: root})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if plan.Empty() {
		t.Fatal("plan is empty, want four files")
	}

	want := []string{"404.html", "guides/writing.html", "index.html", "style.css"}
	if strings.Join(plan.Files, ",") != strings.Join(want, ",") {
		t.Errorf("Files = %v, want %v (sorted, slash-separated)", plan.Files, want)
	}
	if !filepath.IsAbs(plan.Output) {
		t.Errorf("Output = %q, want an absolute path", plan.Output)
	}
}

// Preparing must not touch the disk, or a dry run would delete what it claims
// only to be listing.
func TestPrepareRemovesNothing_integration(t *testing.T) {
	root := site(t)
	write(t, root, "public/index.html")

	if _, err := clean.Prepare(clean.Options{Root: root}); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "public", "index.html")); err != nil {
		t.Errorf("Prepare removed a file: %v", err)
	}
}

func TestExecute_integration(t *testing.T) {
	root := site(t)
	for _, rel := range []string{"public/index.html", "public/guides/writing.html", "public/.nojekyll"} {
		write(t, root, rel)
	}

	plan, err := clean.Prepare(clean.Options{Root: root})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := plan.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Clean means empty. A dot-entry is not an exception, because a keep-list is
	// a promise that grows with every host and never shrinks.
	rest, err := os.ReadDir(filepath.Join(root, "public"))
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if len(rest) != 0 {
		t.Errorf("output still holds %d entrie(s), want none", len(rest))
	}
	// The directory itself stays, so the next build writes straight into it.
	if _, err := os.Stat(filepath.Join(root, "public")); err != nil {
		t.Errorf("output directory should survive: %v", err)
	}
}

// The site's own sources sit next to the output, and nothing about emptying one
// may reach the other.
func TestExecuteLeavesSources_integration(t *testing.T) {
	root := site(t)
	write(t, root, "content/index.md")
	write(t, root, "static/logo.svg")
	write(t, root, "public/index.html")

	plan, err := clean.Prepare(clean.Options{Root: root})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := plan.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	for _, rel := range []string{"cress.toml", "content/index.md", "static/logo.svg"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s should have survived: %v", rel, err)
		}
	}
}

// A clean is safe to repeat: the second run has nothing to do and says so
// rather than failing, so `cress clean && cress build` works every time.
func TestPrepareMissingOutput_integration(t *testing.T) {
	plan, err := clean.Prepare(clean.Options{Root: site(t)})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if !plan.Empty() {
		t.Error("a missing output directory should yield an empty plan")
	}
	if err := plan.Execute(); err != nil {
		t.Errorf("Execute on an empty plan: %v", err)
	}
}

func TestPrepareEmptyOutput_integration(t *testing.T) {
	root := site(t)
	if err := os.Mkdir(filepath.Join(root, "public"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	plan, err := clean.Prepare(clean.Options{Root: root})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if !plan.Empty() {
		t.Error("an empty output directory should yield an empty plan")
	}
}

// Every one of these would destroy the site or something that is not the site.
// Prepare has to refuse before it collects anything, so no caller can reach
// Execute with one of them.
func TestPrepareGuards_integration(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{"the site root", ".", "is the site root"},
		{"the parent of the site root", "..", "contains the site root"},
		{"the content directory", "content", "inside the content directory"},
		{"a directory under content", "content/nested", "inside the content directory"},
		{"the static directory", "static", "inside the static directory"},
		{"the themes directory", "themes", "inside the themes directory"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := site(t)
			write(t, root, "content/index.md")

			_, err := clean.Prepare(clean.Options{Root: root, Output: tc.output})
			if err == nil {
				t.Fatalf("Prepare(%q) succeeded, want a refusal", tc.output)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

// A directory with no config is not a site. This is what stops a clean run from
// the wrong working directory: an unrelated "public" next door is likelier than
// a real site missing its config.
func TestPrepareNotASite_integration(t *testing.T) {
	root := t.TempDir()
	write(t, root, "public/index.html")

	_, err := clean.Prepare(clean.Options{Root: root})
	if err == nil {
		t.Fatal("Prepare succeeded outside a site, want a refusal")
	}
	if !strings.Contains(err.Error(), "not a cress site") {
		t.Errorf("error = %q, want it to say the directory is not a cress site", err)
	}
}

// A guard refuses the site's own directories, not any output path outside the
// site: a build that wrote somewhere else must still be cleanable.
func TestPrepareOutsideSite_integration(t *testing.T) {
	root := site(t)
	elsewhere := t.TempDir()
	write(t, elsewhere, "index.html")

	plan, err := clean.Prepare(clean.Options{Root: root, Output: elsewhere})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if plan.Empty() {
		t.Error("plan is empty, want the file in the external output directory")
	}
}
