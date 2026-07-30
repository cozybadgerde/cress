// Package serve runs a local preview server: it builds the site once, serves
// the output over HTTP, and rebuilds whenever a source file changes. Browser
// auto-refresh is intentionally out of scope; reload the page to see changes.
package serve

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/cozybadgerde/cress/internal/build"
)

const (
	// DefaultAddr is the default listen address for the preview server.
	DefaultAddr = "localhost:1313"

	debounce        = 120 * time.Millisecond
	shutdownTimeout = 5 * time.Second
	readTimeout     = 10 * time.Second

	contentTypeHTML = "text/html; charset=utf-8"
)

// Options configures the preview server.
type Options struct {
	// Root is the site root directory. Defaults to ".".
	Root string
	// Addr is the listen address. Defaults to DefaultAddr.
	Addr string
	// Drafts includes draft pages in the preview when true.
	Drafts bool
	// Logf, when set, receives human-readable progress lines.
	Logf func(format string, args ...any)
}

// Run builds the site, serves it, and rebuilds on change until ctx is
// cancelled.
func Run(ctx context.Context, opts Options) error {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.Addr == "" {
		opts.Addr = DefaultAddr
	}
	logf := opts.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}

	buildOpts := build.Options{Root: opts.Root, Drafts: opts.Drafts}
	res, err := build.Build(buildOpts)
	if err != nil {
		return err
	}
	reportWarnings(logf, res)
	logf("built %d page(s) into %s", res.Pages, res.Output)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("starting watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()
	if err := watchSources(watcher, opts.Root); err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              opts.Addr,
		Handler:           siteHandler(res.Output, res.BasePath),
		ReadHeaderTimeout: readTimeout,
	}
	serveErr := make(chan error, 1)
	go func() {
		// The base path is read from the first build and held for the session: it is
		// the one thing the server itself is configured with rather than reading off
		// disk per request, so changing base_url needs a restart.
		logf("serving http://%s%s/ (Ctrl-C to stop)", opts.Addr, res.BasePath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	return watchLoop(ctx, watcher, srv, buildOpts, res.Output, logf, serveErr)
}

// siteHandler serves the built output, answering a miss with the site's own 404
// page so the preview matches what a static host does.
//
// When the site is rooted under basePath, the preview mounts there too, and the
// server root redirects to it. A site built for a subdirectory has that prefix
// baked into every link it emits, so serving it from "/" would preview a set of
// links that all 404: the mismatch would only show up once the site was
// published, which is the failure this exists to prevent.
func siteHandler(outPath, basePath string) http.Handler {
	files := http.FileServer(http.Dir(outPath))
	site := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files.ServeHTTP(&notFoundWriter{ResponseWriter: w, outPath: outPath}, r)
	})
	if basePath == "" {
		return site
	}

	mounted := http.StripPrefix(basePath, site)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == basePath {
			http.Redirect(w, r, basePath+"/", http.StatusFound)
			return
		}
		if !strings.HasPrefix(r.URL.Path, basePath+"/") {
			// Outside the mount point nothing is served, but the site's own 404 is
			// still the right answer: a host would serve it for this address too.
			if !writeNotFound(w, outPath) {
				http.NotFound(w, r)
			}
			return
		}
		mounted.ServeHTTP(w, r)
	})
}

// notFoundWriter swaps the file server's plain "404 page not found" body for the
// built 404 page. http.FileServer signals a miss only by calling WriteHeader, so
// intercepting that is the one way to replace the body without reimplementing
// its index lookup and redirect handling.
type notFoundWriter struct {
	http.ResponseWriter
	outPath  string
	replaced bool
}

// Unwrap hands the real writer to http.ResponseController, which is how a
// handler reaches flushing, hijacking and deadlines. Wrapping a ResponseWriter
// hides whatever optional interfaces it implements, and the loss shows up as a
// type assertion quietly returning false rather than as a build failure.
func (w *notFoundWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *notFoundWriter) WriteHeader(code int) {
	if code != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(code)
		return
	}
	if !writeNotFound(w.ResponseWriter, w.outPath) {
		w.ResponseWriter.WriteHeader(code)
		return
	}
	w.replaced = true
}

// writeNotFound answers with the built 404 page, reporting whether it could. It
// is read per miss rather than cached, so a rebuild is picked up without
// restarting the server. A build always writes the file, so a failure here means
// someone removed it mid-session and the caller answers however it must.
func writeNotFound(w http.ResponseWriter, outPath string) bool {
	// #nosec G304 -- outPath is the build's own output directory.
	page, err := os.ReadFile(filepath.Join(outPath, build.NotFoundFile))
	if err != nil {
		return false
	}
	w.Header().Set("Content-Type", contentTypeHTML)
	w.Header().Set("Content-Length", strconv.Itoa(len(page)))
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(page)
	return true
}

// Write drops the file server's own error body once the 404 page has replaced
// it, so the two are not concatenated.
func (w *notFoundWriter) Write(b []byte) (int, error) {
	if w.replaced {
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

// watchLoop rebuilds on debounced filesystem events and shuts the server down
// when ctx is cancelled.
func watchLoop(ctx context.Context, watcher *fsnotify.Watcher, srv *http.Server, buildOpts build.Options, outPath string, logf func(string, ...any), serveErr <-chan error) error {
	timer := time.NewTimer(debounce)
	timer.Stop()
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			shutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
			return srv.Shutdown(shutCtx)

		case err := <-serveErr:
			return fmt.Errorf("http server: %w", err)

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if relevantEvent(watcher, event, outPath) {
				timer.Reset(debounce)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			logf("watch error: %v", err)

		case <-timer.C:
			rebuild(buildOpts, logf)
		}
	}
}

// relevantEvent reports whether event should trigger a rebuild. A newly created
// directory is added to the watch set on the way through, so files written into
// it afterwards are noticed too.
func relevantEvent(watcher *fsnotify.Watcher, event fsnotify.Event, outPath string) bool {
	if ignoreEvent(event.Name, outPath) {
		return false
	}
	if event.Has(fsnotify.Create) {
		addIfDir(watcher, event.Name)
	}
	return true
}

// rebuild renders the site again and logs the outcome. A failed build is
// reported and swallowed: the preview server keeps serving the last good
// output so a typo in the config does not end the session.
func rebuild(buildOpts build.Options, logf func(string, ...any)) {
	res, err := build.Build(buildOpts)
	if err != nil {
		logf("rebuild failed: %v", err)
		return
	}
	reportWarnings(logf, res)
	logf("rebuilt %d page(s)", res.Pages)
}

// watchSources registers the source directories (content, themes, static, each
// recursively) plus the site root itself, so config changes are noticed. The
// output directory is deliberately not watched.
func watchSources(watcher *fsnotify.Watcher, root string) error {
	if err := watcher.Add(root); err != nil {
		return fmt.Errorf("watching %s: %w", root, err)
	}
	for _, sub := range []string{build.ContentDir, build.ThemesDir, build.StaticDir} {
		dir := filepath.Join(root, sub)
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("watching %s: %w", dir, err)
		}
	}
	return nil
}

// addIfDir starts watching a newly created directory so nested changes are seen.
func addIfDir(watcher *fsnotify.Watcher, path string) {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		_ = watcher.Add(path)
	}
}

// ignoreEvent reports whether an event path should be skipped: anything inside
// the output directory (to avoid a rebuild loop), and editor swap noise.
func ignoreEvent(path, outPath string) bool {
	if path == outPath || strings.HasPrefix(path, outPath+string(os.PathSeparator)) {
		return true
	}
	base := filepath.Base(path)
	return strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".swp")
}

// reportWarnings surfaces any non-fatal build warnings through logf.
func reportWarnings(logf func(string, ...any), res *build.Result) {
	for _, w := range res.Warnings {
		logf("warning: %s", w)
	}
}
