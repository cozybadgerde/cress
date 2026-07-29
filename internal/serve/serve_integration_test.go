package serve_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cozybadgerde/cress/internal/scaffold"
	"github.com/cozybadgerde/cress/internal/serve"
)

func TestServe_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	addr := freeAddr(t)

	ctx, cancel := context.WithCancel(context.Background())
	// Safety net: a mid-test t.Fatal still cancels the server. cancel is
	// idempotent, and runErr is buffered, so the goroutine never leaks.
	t.Cleanup(cancel)

	runErr := make(chan error, 1)
	go func() {
		runErr <- serve.Run(ctx, serve.Options{Root: root, Addr: addr})
	}()

	base := "http://" + addr

	// The initial build is served.
	if body := waitForBody(t, base+"/", "Welcome"); body == "" {
		t.Fatal("initial page never became available")
	}

	// Editing content triggers a rebuild that the server then serves.
	edit := "---\ntitle: Changed\n---\n\n# Changed heading MARKER_V2\n"
	if err := os.WriteFile(filepath.Join(root, "content", "index.md"), []byte(edit), 0o644); err != nil {
		t.Fatalf("edit content: %v", err)
	}
	if body := waitForBody(t, base+"/", "MARKER_V2"); body == "" {
		t.Fatal("rebuild after edit was never served")
	}

	// A miss is answered with the site's own 404 page, the way a static host
	// does, instead of the file server's plain-text default.
	assertNotFoundPage(t, base+"/nope.html")

	// Cancelling shuts the server down cleanly.
	cancel()
	select {
	case err := <-runErr:
		if err != nil {
			t.Errorf("serve.Run returned %v, want nil on cancel", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve.Run did not shut down after cancel")
	}
}

// freeAddr returns a localhost address that is free at the moment of the call.
func freeAddr(t *testing.T) string {
	t.Helper()
	var lc net.ListenConfig
	l, err := lc.Listen(context.Background(), "tcp", "localhost:0")
	if err != nil {
		t.Fatalf("reserving a port: %v", err)
	}
	addr := l.Addr().String()
	if err := l.Close(); err != nil {
		t.Fatalf("closing reserved listener: %v", err)
	}
	return addr
}

// assertNotFoundPage checks that url answers with the built 404 page rather
// than net/http's plain-text default.
func assertNotFoundPage(t *testing.T, url string) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(url) //nolint:noctx // short-lived test request with a client timeout
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("reading body of %s: %v", url, err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET %s = %d, want %d", url, resp.StatusCode, http.StatusNotFound)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(string(body), "Page not found") {
		t.Errorf("body = %q, want the built 404 page", body)
	}
	// The file server writes its own body after signalling the miss; if that is
	// not discarded the two appear concatenated.
	if strings.Contains(string(body), "404 page not found") {
		t.Error("the file server's plain-text 404 leaked into the response")
	}
}

// waitForBody polls url until the response body contains want, and returns the
// matching body. It returns "" if want never appears within the deadline.
func waitForBody(t *testing.T, url, want string) string {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url) //nolint:noctx // short-lived test poll with a client timeout
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if strings.Contains(string(body), want) {
				return string(body)
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return ""
}
