package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// run executes a fresh root command with the given args (program name added),
// discarding its output.
func run(t *testing.T, args ...string) error {
	t.Helper()
	cmd := newRootCommand()
	cmd.Writer = io.Discard
	cmd.ErrWriter = io.Discard
	return cmd.Run(context.Background(), append([]string{"cress"}, args...))
}

func TestCLIInitThenBuild_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	site := t.TempDir()
	if err := run(t, "init", site); err != nil {
		t.Fatalf("cress init: %v", err)
	}
	if _, err := os.Stat(filepath.Join(site, "cress.toml")); err != nil {
		t.Fatalf("init did not scaffold cress.toml: %v", err)
	}

	if err := run(t, "build", "--source", site); err != nil {
		t.Fatalf("cress build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(site, "public", "index.html")); err != nil {
		t.Errorf("build did not produce public/index.html: %v", err)
	}
}

func TestCLIUnknownCommand_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if err := run(t, "frobnicate"); err == nil {
		t.Fatal("expected an error for an unknown command")
	}
}
