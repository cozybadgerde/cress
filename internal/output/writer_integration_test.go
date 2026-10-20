package output

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

var errInjected = errors.New("injected filesystem failure")

func acquireWriter(t *testing.T, site string) *Writer {
	t.Helper()
	writer, err := Acquire(context.Background(), site)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := writer.Close(); err != nil {
			t.Error(err)
		}
	})
	return writer
}

func preparedFile(name, body string) File {
	return File{Path: name, Source: "test source " + name, Data: []byte(body)}
}

func assertFile(t *testing.T, site, name, want string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(site, name))
	if err != nil || string(data) != want {
		t.Fatalf("%s = %q, %v; want %q", name, data, err, want)
	}
}

func TestReplacementAndStagingFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	writer := acquireWriter(t, site)
	if err := writer.Write(context.Background(), []File{preparedFile("old.html", "old")}); err != nil {
		t.Fatal(err)
	}
	err := writer.Write(context.Background(), []File{preparedFile(strings.Repeat("x", 256), "invalid")})
	if err == nil {
		t.Fatal("oversized path accepted")
	}
	assertFile(t, site, "public/old.html", "old")
	if err := os.WriteFile(filepath.Join(site, "public/manual.txt"), []byte("manual"), fileMode); err != nil {
		t.Fatal(err)
	}
	if err := writer.Write(context.Background(), []File{preparedFile("new.html", "new")}); err != nil {
		t.Fatal(err)
	}
	assertFile(t, site, "public/new.html", "new")
	for _, name := range []string{"public/old.html", "public/manual.txt"} {
		if _, err := os.Lstat(filepath.Join(site, name)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("stale file %s remains: %v", name, err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(site, stateDirectory))
	if err != nil || len(entries) != 2 {
		t.Fatalf("unexpected retained state: %v, %v", entries, err)
	}
}

func TestPublicationFailuresAndRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, failure := range []string{"backup", "publish", "restore", "cleanup"} {
		t.Run(failure, func(t *testing.T) {
			site := t.TempDir()
			writer := acquireWriter(t, site)
			if err := writer.Write(context.Background(), []File{preparedFile("index.html", "old")}); err != nil {
				t.Fatal(err)
			}
			writer.rename = func(old, new string) error {
				if (failure == "backup" && old == "public") ||
					((failure == "publish" || failure == "restore") && strings.Contains(old, "/stage-")) ||
					(failure == "restore" && strings.Contains(old, "/backup-")) {
					return errInjected
				}
				return writer.root.Rename(old, new)
			}
			if failure == "cleanup" {
				writer.removeAll = func(string) error { return errInjected }
			}
			err := writer.Write(context.Background(), []File{preparedFile("index.html", "new")})
			if !errors.Is(err, errInjected) {
				t.Fatalf("error = %v, want explicit injected failure", err)
			}
			switch failure {
			case "cleanup":
				assertFile(t, site, "public/index.html", "new")
				assertFile(t, site, writer.state.Pending.backup()+"/index.html", "old")
			case "restore":
				if !strings.Contains(err.Error(), "restore previous output failed") ||
					!strings.Contains(err.Error(), "publish staged output") {
					t.Fatalf("both failures not reported: %v", err)
				}
				assertFile(t, site, writer.state.Pending.backup()+"/index.html", "old")
			default:
				assertFile(t, site, "public/index.html", "old")
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			recovered := acquireWriter(t, site)
			want := "old"
			if failure == "cleanup" {
				want = "new"
			}
			assertFile(t, site, "public/index.html", want)
			if recovered.state.Pending != nil {
				t.Fatal("recovery transaction remains")
			}
		})
	}
}

func TestInterruptionRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, point := range []string{"recorded", "empty stage", "partial stage", "ready", "backed up", "promoted", "confirmed"} {
		t.Run(point, func(t *testing.T) {
			site := t.TempDir()
			writer := acquireWriter(t, site)
			if err := writer.Write(context.Background(), []File{preparedFile("index.html", "old")}); err != nil {
				t.Fatal(err)
			}
			pending := &transaction{
				Token: strings.Repeat("a", tokenBytes*2), Previous: writer.state.Current,
				HadPublic: true, Phase: "staging",
			}
			writer.state.Pending = pending
			if err := writer.saveState(); err != nil {
				t.Fatal(err)
			}
			switch point {
			case "recorded":
			case "empty stage":
				if err := writer.root.Mkdir(pending.stage(), privateMode); err != nil {
					t.Fatal(err)
				}
			default:
				if err := writer.stage(context.Background(), []File{preparedFile("index.html", "new")}); err != nil {
					t.Fatal(err)
				}
				if point != "partial stage" {
					pending.Phase = "ready"
					if err := writer.saveState(); err != nil {
						t.Fatal(err)
					}
				}
				if point == "backed up" || point == "promoted" || point == "confirmed" {
					if err := writer.root.Rename("public", pending.backup()); err != nil {
						t.Fatal(err)
					}
				}
				if point == "promoted" || point == "confirmed" {
					if err := writer.root.Rename(pending.stage(), "public"); err != nil {
						t.Fatal(err)
					}
				}
				if point == "confirmed" {
					pending.Phase = "published"
					writer.state.Current = pending.Token
					if err := writer.saveState(); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			recovered := acquireWriter(t, site)
			want := "old"
			if point == "promoted" || point == "confirmed" {
				want = "new"
			}
			assertFile(t, site, "public/index.html", want)
			if recovered.state.Pending != nil {
				t.Fatal("transaction not cleared")
			}
		})
	}
}

func TestAmbiguousStatePreservesEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, kind := range []string{"state symlink", "lock symlink", "orphan stage", "next state", "marker symlink", "foreign marker", "broken record"} {
		t.Run(kind, func(t *testing.T) {
			site := t.TempDir()
			writer := acquireWriter(t, site)
			if err := writer.Write(context.Background(), []File{preparedFile("index.html", "keep")}); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "state symlink", "lock symlink", "marker symlink":
				name := statePath
				switch kind {
				case "lock symlink":
					name = lockPath
				case "marker symlink":
					name = "public/" + markerPath
				}
				if err := os.Rename(filepath.Join(site, name), filepath.Join(site, "keep")); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(site, "keep"), filepath.Join(site, name)); err != nil {
					t.Fatal(err)
				}
			case "orphan stage":
				if err := os.Mkdir(filepath.Join(site, ".cress/stage-unrecognized"), privateMode); err != nil {
					t.Fatal(err)
				}
			case "next state":
				if err := os.WriteFile(filepath.Join(site, nextStatePath), []byte("evidence"), privateFileMode); err != nil {
					t.Fatal(err)
				}
			case "foreign marker":
				if err := os.WriteFile(filepath.Join(site, "public", markerPath),
					[]byte(`{"version":1,"token":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`), fileMode); err != nil {
					t.Fatal(err)
				}
			case "broken record":
				if err := os.WriteFile(filepath.Join(site, statePath), []byte(`{}`), privateFileMode); err != nil {
					t.Fatal(err)
				}
			}
			err := Write(context.Background(), site, []File{preparedFile("index.html", "replace")})
			if err == nil {
				t.Fatal("unverified state accepted")
			}
			assertFile(t, site, "public/index.html", "keep")
			if kind == "next state" {
				assertFile(t, site, nextStatePath, "evidence")
			}
		})
	}
}

func TestWriterLockAliasesAndProcessExit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process integration test")
	}
	if site := os.Getenv("CRESS_TEST_LOCK_SITE"); site != "" {
		writer, err := Acquire(context.Background(), site)
		if os.Getenv("CRESS_TEST_LOCK_BUSY") == "true" {
			if !errors.Is(err, ErrBusy) {
				t.Fatalf("competing process did not observe contention: %v", err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		// Process termination must release the lock even without Close.
		_ = writer
		os.Exit(0)
	}
	site := t.TempDir()
	writer := acquireWriter(t, site)
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(site, alias); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{site, alias, filepath.Join(alias, ".")} {
		if _, err := Acquire(context.Background(), name); !errors.Is(err, ErrBusy) {
			t.Fatalf("alias %q bypassed writer: %v", name, err)
		}
	}
	competitor := exec.Command(os.Args[0], "-test.run=^TestWriterLockAliasesAndProcessExit$")
	competitor.Env = append(os.Environ(), "CRESS_TEST_LOCK_SITE="+alias, "CRESS_TEST_LOCK_BUSY=true")
	if data, err := competitor.CombinedOutput(); err != nil {
		t.Fatalf("competing writer: %v\n%s", err, data)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(os.Args[0], "-test.run=^TestWriterLockAliasesAndProcessExit$")
	command.Env = append(os.Environ(), "CRESS_TEST_LOCK_SITE="+site)
	if data, err := command.CombinedOutput(); err != nil {
		t.Fatalf("child writer: %v\n%s", err, data)
	}
	acquireWriter(t, site)
}

func TestPrivateStateRefusesUnsafeTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, kind := range []string{"directory symlink", "directory file", "public permissions", "lock FIFO"} {
		t.Run(kind, func(t *testing.T) {
			site, protected := t.TempDir(), t.TempDir()
			if err := os.WriteFile(filepath.Join(protected, "keep"), []byte("protected"), fileMode); err != nil {
				t.Fatal(err)
			}
			directory := filepath.Join(site, stateDirectory)
			switch kind {
			case "directory symlink":
				if err := os.Symlink(protected, directory); err != nil {
					t.Fatal(err)
				}
			case "directory file":
				if err := os.WriteFile(directory, []byte("keep"), privateFileMode); err != nil {
					t.Fatal(err)
				}
			default:
				mode := os.FileMode(privateMode)
				if kind == "public permissions" {
					mode = directoryMode
				}
				if err := os.Mkdir(directory, mode); err != nil {
					t.Fatal(err)
				}
				if kind == "lock FIFO" {
					if err := syscall.Mkfifo(filepath.Join(site, lockPath), privateFileMode); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err := Acquire(context.Background(), site); err == nil {
				t.Fatal("unsafe private state accepted")
			}
			assertFile(t, protected, "keep", "protected")
			entries, err := os.ReadDir(protected)
			if err != nil || len(entries) != 1 {
				t.Fatalf("writer followed unsafe state: %v, %v", entries, err)
			}
		})
	}
}

func TestStateInstallationFailurePreservesPublishedOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	writer := acquireWriter(t, site)
	if err := writer.Write(context.Background(), []File{preparedFile("index.html", "old")}); err != nil {
		t.Fatal(err)
	}
	writer.rename = func(old, new string) error {
		if old == nextStatePath && writer.state.Pending != nil && writer.state.Pending.Phase == "published" {
			return errInjected
		}
		return writer.root.Rename(old, new)
	}
	err := writer.Write(context.Background(), []File{preparedFile("index.html", "new")})
	if !errors.Is(err, errInjected) || !strings.Contains(err.Error(), "new output is published") {
		t.Fatalf("publication-state failure hidden: %v", err)
	}
	assertFile(t, site, "public/index.html", "new")
	backup := writer.state.Pending.backup()
	assertFile(t, site, backup+"/index.html", "old")
	if err := writer.Write(context.Background(), []File{preparedFile("index.html", "retry")}); err == nil ||
		!strings.Contains(err.Error(), nextStatePath) {
		t.Fatalf("same writer discarded unconfirmed state: %v", err)
	}
	assertFile(t, site, backup+"/index.html", "old")
	assertFile(t, site, "public/index.html", "new")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Acquire(context.Background(), site); err == nil ||
		!strings.Contains(err.Error(), nextStatePath) {
		t.Fatalf("unconfirmed next state was not preserved: %v", err)
	}
	assertFile(t, site, backup+"/index.html", "old")
	assertFile(t, site, "public/index.html", "new")
}

func TestRecoveryRejectsChangedBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	writer := acquireWriter(t, site)
	if err := writer.Write(context.Background(), []File{preparedFile("index.html", "old")}); err != nil {
		t.Fatal(err)
	}
	writer.removeAll = func(string) error { return errInjected }
	if err := writer.Write(context.Background(), []File{preparedFile("index.html", "new")}); !errors.Is(err, errInjected) {
		t.Fatal(err)
	}
	backup := writer.state.Pending.backup()
	if err := writer.root.WriteFile(backup+"/"+markerPath, []byte("unverifiable"), fileMode); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Acquire(context.Background(), site); err == nil || !strings.Contains(err.Error(), "preserved recovery evidence") {
		t.Fatalf("changed backup accepted or guidance missing: %v", err)
	}
	assertFile(t, site, backup+"/index.html", "old")
	assertFile(t, site, "public/index.html", "new")
}
