package output

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestSession(t *testing.T, site string) *Session {
	t.Helper()
	session, warnings, err := NewSession(context.Background(), site)
	if err != nil {
		t.Fatalf("new session: %v (%v)", err, warnings)
	}
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Error(err)
		}
	})
	return session
}

func TestSessionIsolationAndOrphanReclamation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	if err := os.Mkdir(filepath.Join(site, "public"), directoryMode); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(site, "public/keep"), []byte("deployable"), fileMode); err != nil {
		t.Fatal(err)
	}
	active := newTestSession(t, site)
	orphan := newTestSession(t, site)
	for _, session := range []*Session{active, orphan} {
		token, err := session.Attempt(context.Background(), func(writer *Writer) error {
			return writer.Write(context.Background(), []File{preparedFile("index.html", session.token)})
		})
		if err != nil || !validToken(token) {
			t.Fatalf("generate: %s, %v", token, err)
		}
	}
	info, err := os.Stat(active.directory)
	if err != nil || info.Mode().Perm() != privateMode {
		t.Fatalf("nonprivate session: %v, %v", info, err)
	}
	if err := orphan.lease.Close(); err != nil {
		t.Fatal(err)
	}
	orphan.lease = nil
	unknown := filepath.Join(filepath.Dir(active.directory), "unrelated")
	if err := os.WriteFile(unknown, []byte("keep"), privateFileMode); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(unknown); err != nil {
			t.Error(err)
		}
	})
	next, warnings, err := NewSession(context.Background(), site)
	if err != nil || len(warnings) != 1 || !strings.Contains(warnings[0], "unrelated") {
		t.Fatalf("new session = %v, %v", err, warnings)
	}
	if err := next.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(orphan.directory); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("verified orphan remains: %v", err)
	}
	assertFile(t, site, "public/keep", "deployable")
	assertFile(t, active.directory, "public/index.html", active.token)
	assertFile(t, filepath.Dir(active.directory), "unrelated", "keep")
	directory := active.directory
	if err := active.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(directory); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("orderly shutdown retained session: %v", err)
	}
}

func TestSessionPublicationFailureUsability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, failure := range []string{"publish", "restore", "cleanup"} {
		t.Run(failure, func(t *testing.T) {
			site := t.TempDir()
			session := newTestSession(t, site)
			first, err := session.Attempt(context.Background(), func(writer *Writer) error {
				return writer.Write(context.Background(), []File{preparedFile("index.html", "old")})
			})
			if err != nil {
				t.Fatal(err)
			}
			token, err := session.Attempt(context.Background(), func(writer *Writer) error {
				writer.rename = func(old, new string) error {
					if ((failure == "publish" || failure == "restore") && strings.Contains(old, "/stage-")) ||
						(failure == "restore" && strings.Contains(old, "/backup-")) {
						return errInjected
					}
					return writer.root.Rename(old, new)
				}
				if failure == "cleanup" {
					writer.removeAll = func(string) error { return errInjected }
				}
				return writer.Write(context.Background(), []File{preparedFile("index.html", "new")})
			})
			if !errors.Is(err, errInjected) {
				t.Fatalf("failure hidden: %v", err)
			}
			switch failure {
			case "publish":
				if token != first {
					t.Fatal("restored output was not usable")
				}
				assertFile(t, session.directory, "public/index.html", "old")
			case "cleanup":
				if !validToken(token) || token == first {
					t.Fatal("confirmed new output was not usable")
				}
				assertFile(t, session.directory, "public/index.html", "new")
			case "restore":
				if token != "" {
					t.Fatal("failed restoration reported usable output")
				}
			}
		})
	}
}

func TestSessionPreservesAmbiguousEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	session, _, err := NewSession(context.Background(), site)
	if err != nil {
		t.Fatal(err)
	}
	directory := session.directory
	extra := filepath.Join(directory, "unknown")
	if err := os.WriteFile(extra, []byte("keep"), privateFileMode); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err == nil {
		t.Fatal("ambiguous cleanup succeeded")
	}
	assertFile(t, directory, "unknown", "keep")
	_, _, err = NewSession(context.Background(), site)
	if err == nil || !strings.Contains(err.Error(), directory) {
		t.Fatalf("ambiguous orphan not reported: %v", err)
	}
	assertFile(t, directory, "unknown", "keep")
	if err := os.Remove(extra); err != nil {
		t.Fatal(err)
	}
	reclaimed := newTestSession(t, site)
	if reclaimed.directory == directory {
		t.Fatal("session identity reused")
	}
}

func TestSessionBusyAttemptRetainsVerifiedOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	session := newTestSession(t, site)
	token, err := session.Attempt(context.Background(), func(writer *Writer) error {
		return writer.Write(context.Background(), []File{preparedFile("index.html", "old")})
	})
	if err != nil {
		t.Fatal(err)
	}
	writer := acquireWriter(t, site)
	got, err := session.Attempt(context.Background(), func(*Writer) error {
		t.Fatal("busy attempt generated anyway")
		return nil
	})
	if !errors.Is(err, ErrBusy) || got != token {
		t.Fatalf("contention lost usable output: %s, %v", got, err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestChangedSessionIdentityAndMissingOutputArePreserved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, kind := range []string{"identity", "missing output"} {
		t.Run(kind, func(t *testing.T) {
			site := t.TempDir()
			session, _, err := NewSession(context.Background(), site)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := session.Attempt(context.Background(), func(writer *Writer) error {
				return writer.Write(context.Background(), []File{preparedFile("index.html", "keep")})
			}); err != nil {
				t.Fatal(err)
			}
			record, err := os.ReadFile(filepath.Join(session.directory, sessionRecord))
			if err != nil {
				t.Fatal(err)
			}
			if kind == "identity" {
				if err := os.WriteFile(filepath.Join(session.directory, sessionRecord), []byte("{}"), privateFileMode); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Rename(filepath.Join(session.directory, "public"), filepath.Join(site, "saved-preview")); err != nil {
				t.Fatal(err)
			}
			if token, err := session.Verify(); err == nil || token != "" {
				t.Fatalf("changed session still verified: %s, %v", token, err)
			}
			if err := session.Close(); err == nil {
				t.Fatal("unverifiable session was deleted")
			}
			if _, err := os.Stat(filepath.Join(session.directory, sessionRecord)); err != nil {
				t.Fatalf("evidence disappeared: %v", err)
			}
			if kind == "identity" {
				if err := os.WriteFile(filepath.Join(session.directory, sessionRecord), record, privateFileMode); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Rename(filepath.Join(site, "saved-preview"), filepath.Join(session.directory, "public")); err != nil {
				t.Fatal(err)
			}
			reclaimed := newTestSession(t, site)
			if reclaimed.directory == session.directory {
				t.Fatal("reclaimed session identity was reused")
			}
		})
	}
}
