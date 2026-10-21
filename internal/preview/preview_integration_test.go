package preview

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/output"
)

const testTimeout = 10 * time.Second

type terminal struct {
	mu      sync.Mutex
	text    strings.Builder
	changed chan struct{}
}

func newTerminal() *terminal { return &terminal{changed: make(chan struct{}, 1)} }

func (terminal *terminal) Write(data []byte) (int, error) {
	terminal.mu.Lock()
	defer terminal.mu.Unlock()
	count, err := terminal.text.Write(data)
	select {
	case terminal.changed <- struct{}{}:
	default:
	}
	return count, err
}

func (terminal *terminal) snapshot() string {
	terminal.mu.Lock()
	defer terminal.mu.Unlock()
	return terminal.text.String()
}

func waitFor(t *testing.T, terminal *terminal, contains string) string {
	t.Helper()
	timeout := time.NewTimer(testTimeout)
	defer timeout.Stop()
	for {
		if text := terminal.snapshot(); strings.Contains(text, contains) {
			return text
		}
		select {
		case <-terminal.changed:
		case <-timeout.C:
			t.Fatalf("timeout waiting for %q:\n%s", contains, terminal.snapshot())
		}
	}
}

func writeSource(t *testing.T, site, name, data string) {
	t.Helper()
	filename := filepath.Join(site, name)
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func startTestPreview(t *testing.T, options Options, build buildFunc) (string, *terminal, *terminal, func()) {
	t.Helper()
	stdout, stderr := newTerminal(), newTerminal()
	options.Output, options.Errors = stdout, stderr
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, options, build) }()
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(testTimeout):
			t.Error("preview did not shut down")
		}
	}
	t.Cleanup(stop)
	text := waitFor(t, stdout, "Preview http://")
	url := regexp.MustCompile(`http://127\.0\.0\.1:\d+`).FindString(text)
	if url == "" {
		t.Fatal("missing loopback URL: " + text)
	}
	return url, stdout, stderr, stop
}

func getPage(t *testing.T, address string) (int, string) {
	t.Helper()
	client := &http.Client{Timeout: testTimeout}
	response, err := client.Get(address)
	if err != nil {
		t.Fatal(err)
	}
	data, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil {
		t.Fatalf("read response: %v, %v", readErr, closeErr)
	}
	return response.StatusCode, string(data)
}

func TestPreviewSessionsWatchRetryAndIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem and HTTP integration test")
	}
	site := t.TempDir()
	writeSource(t, site, "cress.toml", "[site]\nhost = 'https://example.org'\npath = '/company/'")
	writeSource(t, site, "content/index.md", "+++\ntitle = 'Home'\n+++\nOriginal")
	writeSource(t, site, "content/private.md", "+++\ntitle = 'Private'\ndraft = true\n+++\nPrivate preview body")
	options := Options{Directory: site, Version: "dev", Port: 0}
	first, stdout, stderr, stopFirst := startTestPreview(t, options, generator.BuildPreview)
	options.Drafts = true
	second, secondOut, secondErr, stopSecond := startTestPreview(t, options, generator.BuildPreview)
	if !strings.Contains(stdout.snapshot(), "drafts excluded") || !strings.Contains(secondOut.snapshot(), "drafts included") {
		t.Fatal("draft mode not communicated")
	}
	for _, test := range []struct {
		address string
		status  int
		text    string
	}{
		{first + "/", http.StatusOK, "Original"},
		{first + "/company/private/", http.StatusNotFound, ""},
		{second + "/company/private/", http.StatusOK, "Private preview body"},
		{first + "/company/.cress-output.json", http.StatusNotFound, ""},
		{first + "/.cress/output.json", http.StatusNotFound, ""},
	} {
		status, body := getPage(t, test.address)
		if status != test.status || !strings.Contains(body, test.text) {
			t.Fatalf("%s = %d, %s", test.address, status, body)
		}
	}
	// Read-only serving must not hold the normal build's writer lock.
	if _, err := generator.Build(context.Background(), site); err != nil {
		t.Fatal(err)
	}
	writer, err := output.Acquire(context.Background(), site)
	if err != nil {
		t.Fatal(err)
	}
	writeSource(t, site, "content/index.md", "+++\ntitle = 'Home'\n+++\nAfter contention")
	waitFor(t, stderr, "retry pending")
	waitFor(t, secondErr, "retry pending")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, stdout, "Rebuilt")
	waitFor(t, secondOut, "Rebuilt")
	for _, address := range []string{first, second} {
		_, body := getPage(t, address+"/company/")
		if !strings.Contains(body, "After contention") {
			t.Fatal("blocked edit was lost: " + body)
		}
	}
	writeSource(t, site, "content/index.md", "missing title")
	waitFor(t, stderr, "rebuild failed")
	_, body := getPage(t, first+"/company/")
	if !strings.Contains(body, "After contention") {
		t.Fatal("invalid edit replaced last good output")
	}
	writeSource(t, site, "content/index.md", "+++\ntitle = 'Home'\n+++\nCorrected")
	writeSource(t, site, "content/new/index.md", "+++\ntitle = 'New'\n+++\nNested new page")
	awaitBody(t, first+"/company/new/", "Nested new page")
	writeSource(t, site, "content/new/index.md", "+++\ntitle = 'New'\n+++\nNested edited page")
	awaitBody(t, first+"/company/new/", "Nested edited page")
	writeSource(t, site, "cress.next", "[site]\nhost = 'https://example.org'\npath = '/changed/'")
	if err := os.Rename(filepath.Join(site, "cress.next"), filepath.Join(site, "cress.toml")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, stdout, "Preview URL changed: "+first+"/changed/")
	awaitBody(t, first+"/changed/", "Corrected")
	deployed, err := os.ReadFile(filepath.Join(site, "public/index.html"))
	if err != nil || !strings.Contains(string(deployed), "Original") || strings.Contains(string(deployed), "Private preview body") {
		t.Fatalf("automatic rebuild modified deployable output: %s, %v", deployed, err)
	}
	stopSecond()
	stopFirst()
}

func awaitBody(t *testing.T, address, want string) {
	t.Helper()
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		_, body := getPage(t, address)
		if strings.Contains(body, want) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s did not serve %q", address, want)
}

func TestEditsDuringBuildAreNotLost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem and HTTP integration test")
	}
	site := t.TempDir()
	writeSource(t, site, "content/index.md", "+++\ntitle = 'Home'\n+++\nInitial")
	built, release := make(chan struct{}), make(chan struct{})
	count := 0
	build := func(ctx context.Context, request generator.PreviewRequest) (generator.Result, error) {
		count++
		result, err := generator.BuildPreview(ctx, request)
		if count == 2 {
			close(built)
			select {
			case <-release:
			case <-ctx.Done():
			}
		}
		return result, err
	}
	address, _, _, stop := startTestPreview(t, Options{Directory: site, Version: "dev"}, build)
	writeSource(t, site, "content/index.md", "+++\ntitle = 'Home'\n+++\nFirst edit")
	select {
	case <-built:
	case <-time.After(testTimeout):
		t.Fatal("rebuild did not start")
	}
	writeSource(t, site, "content/index.md", "+++\ntitle = 'Home'\n+++\nLatest edit")
	close(release)
	awaitBody(t, address+"/", "Latest edit")
	stop()
}

func TestStartupContentionAndOccupiedPort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem and listener integration test")
	}
	site := t.TempDir()
	writeSource(t, site, "content/index.md", "+++\ntitle = 'Home'\n+++")
	options := Options{Directory: site, Version: "dev", Output: io.Discard, Errors: io.Discard}
	writer, err := output.Acquire(context.Background(), site)
	if err != nil {
		t.Fatal(err)
	}
	err = Run(context.Background(), options)
	if closeErr := writer.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if !errors.Is(err, output.ErrBusy) {
		t.Fatalf("initial preview did not fail fast: %v", err)
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	options.Port = listener.Addr().(*net.TCPAddr).Port
	err = Run(context.Background(), options)
	if closeErr := listener.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err == nil || !strings.Contains(err.Error(), "listen for preview") {
		t.Fatalf("occupied port did not fail: %v", err)
	}
}

func TestUnverifiedOutputStopsPreviewAndPreservesEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem and HTTP integration test")
	}
	site := t.TempDir()
	writeSource(t, site, "content/index.md", "+++\ntitle = 'Home'\n+++\nVerified")
	stdout, stderr := newTerminal(), newTerminal()
	var directory string
	var marker []byte
	count := 0
	build := func(ctx context.Context, request generator.PreviewRequest) (generator.Result, error) {
		count++
		if count == 2 {
			// Break ownership, not the site inputs, to exercise the fatal serving boundary.
			if err := os.WriteFile(filepath.Join(directory, "public/.cress-output.json"), []byte("altered"), 0o644); err != nil {
				return generator.Result{}, err
			}
		}
		result, err := generator.BuildPreview(ctx, request)
		if count == 1 {
			directory = request.Session.Directory()
			marker, err = os.ReadFile(filepath.Join(directory, "public/.cress-output.json"))
		}
		return result, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- run(ctx, Options{Directory: site, Version: "dev", Output: stdout, Errors: stderr}, build)
	}()
	waitFor(t, stdout, "Preview http://")
	writeSource(t, site, "content/index.md", "+++\ntitle = 'Home'\n+++\nChange")
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "no verified usable output") ||
			!strings.Contains(err.Error(), directory) {
			t.Fatalf("fatal preview failure missing: %v", err)
		}
	case <-time.After(testTimeout):
		t.Fatal("unverified preview kept running")
	}
	if _, err := os.Stat(filepath.Join(directory, "public/index.html")); err != nil {
		t.Fatalf("recovery evidence deleted: %v", err)
	}
	// Restore only the exact ownership evidence saved by this test before reclamation.
	if err := os.WriteFile(filepath.Join(directory, "public/.cress-output.json"), marker, 0o644); err != nil {
		t.Fatal(err)
	}
	session, _, err := output.NewSession(context.Background(), site)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(directory); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("verified inactive preview not reclaimed: %v", err)
	}
}
