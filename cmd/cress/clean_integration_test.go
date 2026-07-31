package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// execClean executes `cress clean` against a scaffolded, built site, feeding it
// stdin and capturing its output. A strings.Reader is never a terminal, which
// is exactly the non-interactive case the command has to handle.
func execClean(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newRootCommand()
	cmd.Writer = &out
	cmd.ErrWriter = io.Discard
	cmd.Reader = strings.NewReader(stdin)
	err := cmd.Run(context.Background(), append([]string{"cress", "clean"}, args...))
	return out.String(), err
}

// builtSite scaffolds and builds a site, returning its root.
func builtSite(t *testing.T) string {
	t.Helper()
	site := t.TempDir()
	if err := run(t, "init", site); err != nil {
		t.Fatalf("cress init: %v", err)
	}
	if err := run(t, "build", "--source", site); err != nil {
		t.Fatalf("cress build: %v", err)
	}
	return site
}

// outputEntries counts what is left in the site's output directory.
func outputEntries(t *testing.T, site string) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(site, "public"))
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	return len(entries)
}

func TestCLICleanForce_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := builtSite(t)
	out, err := execClean(t, "", "--source", site, "--force")
	if err != nil {
		t.Fatalf("cress clean --force: %v", err)
	}
	if !strings.Contains(out, "removed") {
		t.Errorf("output should report what went, got: %s", out)
	}
	if n := outputEntries(t, site); n != 0 {
		t.Errorf("output still holds %d entrie(s), want none", n)
	}
}

// Without a terminal there is nobody to answer, so the command has to refuse
// rather than hang a CI job or read end-of-input as a decision.
func TestCLICleanNonInteractive_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := builtSite(t)
	before := outputEntries(t, site)

	_, err := execClean(t, "", "--source", site)
	if err == nil {
		t.Fatal("cress clean without a terminal succeeded, want a refusal")
	}
	for _, want := range []string{"--force", "--dry-run"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should point at %s, got: %v", want, err)
		}
	}
	if n := outputEntries(t, site); n != before {
		t.Errorf("output changed from %d to %d entrie(s), want no deletion", before, n)
	}
}

// /dev/null is a character device, so it passes for a terminal on the cheap
// test and only gives itself away by answering nothing. Left unhandled,
// `cress clean </dev/null` in a CI job prompts nobody, reads end-of-input as a
// no, and exits 0 having done nothing, which reads as success.
func TestCLICleanDevNull_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("opening %s: %v", os.DevNull, err)
	}
	defer func() { _ = devNull.Close() }()

	site := builtSite(t)
	before := outputEntries(t, site)

	var out bytes.Buffer
	cmd := newRootCommand()
	cmd.Writer = &out
	cmd.ErrWriter = io.Discard
	cmd.Reader = devNull

	err = cmd.Run(context.Background(), []string{"cress", "clean", "--source", site})
	if err == nil {
		t.Fatal("cress clean </dev/null succeeded, want a refusal")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("error should point at --force, got: %v", err)
	}
	if n := outputEntries(t, site); n != before {
		t.Errorf("output changed from %d to %d entrie(s), want no deletion", before, n)
	}
}

func TestCLICleanDryRun_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := builtSite(t)
	before := outputEntries(t, site)

	// A dry run deletes nothing, so it needs no terminal and no --force.
	out, err := execClean(t, "", "--source", site, "--dry-run")
	if err != nil {
		t.Fatalf("cress clean --dry-run: %v", err)
	}
	for _, want := range []string{"would remove", "index.html", "nothing removed"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q, got: %s", want, out)
		}
	}
	if n := outputEntries(t, site); n != before {
		t.Errorf("output changed from %d to %d entrie(s), want no deletion", before, n)
	}
}

// Running clean twice is safe: the second run has nothing to do and says so.
func TestCLICleanTwice_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := builtSite(t)
	if _, err := execClean(t, "", "--source", site, "--force"); err != nil {
		t.Fatalf("first clean: %v", err)
	}

	out, err := execClean(t, "", "--source", site, "--force")
	if err != nil {
		t.Fatalf("second clean: %v", err)
	}
	if !strings.Contains(out, "nothing to remove") {
		t.Errorf("a second clean should report nothing to do, got: %s", out)
	}
}

// An empty output directory needs no terminal either: there is nothing to
// consent to, so the command must not refuse for want of one.
func TestCLICleanEmptyNeedsNoTerminal_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	if err := run(t, "init", site); err != nil {
		t.Fatalf("cress init: %v", err)
	}

	out, err := execClean(t, "", "--source", site)
	if err != nil {
		t.Fatalf("cress clean on an unbuilt site: %v", err)
	}
	if !strings.Contains(out, "nothing to remove") {
		t.Errorf("output should report nothing to do, got: %s", out)
	}
}

// A guard is a refusal, not a prompt: --force buys past the question, never
// past the check that the target is not the site itself.
func TestCLICleanForceDoesNotBypassGuards_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := builtSite(t)
	_, err := execClean(t, "", "--source", site, "--output", "content", "--force")
	if err == nil {
		t.Fatal("cress clean -o content --force succeeded, want a refusal")
	}
	if !strings.Contains(err.Error(), "refusing to clean") {
		t.Errorf("error should say it refuses to clean, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(site, "content", "index.md")); statErr != nil {
		t.Errorf("content should have survived: %v", statErr)
	}
}
