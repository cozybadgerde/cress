package preview

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/output"
)

const (
	// DefaultPort is the loopback port used unless explicitly overridden.
	DefaultPort     = 8080
	debounceDelay   = 100 * time.Millisecond
	retryDelay      = 500 * time.Millisecond
	headerTimeout   = 5 * time.Second
	shutdownTimeout = 5 * time.Second
	maxPort         = 65535
	loopbackHost    = "127.0.0.1"
)

// Options fixes the site's preview mode and terminal destinations.
type Options struct {
	Directory string
	Version   string
	Drafts    bool
	Port      int
	Output    io.Writer
	Errors    io.Writer
}

// Run serves isolated output until cancellation or a fatal preview failure.
func Run(ctx context.Context, options Options) (runErr error) {
	return run(ctx, options, generator.BuildPreview)
}

type buildFunc func(context.Context, generator.PreviewRequest) (generator.Result, error)

func run(ctx context.Context, options Options, build buildFunc) (runErr error) {
	if options.Port < 0 || options.Port > maxPort {
		return fmt.Errorf("preview port must be between 0 and %d", maxPort)
	}
	site, err := filepath.Abs(options.Directory)
	if err != nil {
		return err
	}
	site, err = filepath.EvalSymlinks(site)
	if err != nil {
		return err
	}
	session, warnings, err := output.NewSession(ctx, site)
	for _, warning := range warnings {
		if _, reportErr := fmt.Fprintln(options.Errors, "warning: "+warning); reportErr != nil {
			if session != nil {
				reportErr = errors.Join(reportErr, session.Close())
			}
			return errors.Join(err, reportErr)
		}
	}
	if err != nil {
		return err
	}
	defer func() { runErr = errors.Join(runErr, session.Close()) }()
	watcher, err := newSourceWatcher(site)
	if err != nil {
		return fmt.Errorf("start source watching: %w", err)
	}
	watchCtx, cancelWatch := context.WithCancel(ctx)
	go watcher.run(watchCtx)
	defer func() {
		cancelWatch()
		runErr = errors.Join(runErr, watcher.close())
	}()
	listener, err := net.Listen("tcp4", net.JoinHostPort(loopbackHost, strconv.Itoa(options.Port)))
	if err != nil {
		return fmt.Errorf("listen for preview: %w", err)
	}
	pages := &pageServer{}
	defer func() {
		pages.mu.Lock()
		defer pages.mu.Unlock()
		if pages.root != nil {
			runErr = errors.Join(runErr, pages.root.Close())
		}
	}()
	request := generator.PreviewRequest{Version: options.Version, Drafts: options.Drafts, Session: session}
	result, err := build(ctx, request)
	if reportErr := reportWarnings(options.Errors, result.Warnings); reportErr != nil || err != nil {
		return errors.Join(err, reportErr, listener.Close())
	}
	pages.root, err = os.OpenRoot(filepath.Join(session.Directory(), "public"))
	if err != nil {
		return errors.Join(err, listener.Close())
	}
	pages.prefix = result.Prefix
	server := &http.Server{Handler: pages, ReadHeaderTimeout: headerTimeout}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			shutdownErr = errors.Join(shutdownErr, server.Close())
		}
		serveErr := <-served
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
		runErr = errors.Join(runErr, shutdownErr, serveErr)
	}()
	address := "http://" + listener.Addr().String()
	mode := "drafts excluded"
	if options.Drafts {
		mode = "drafts included"
	}
	if _, err := fmt.Fprintf(options.Output, "Preview %s%s (%s). Refresh manually; Ctrl-C to stop.\n", address, pages.prefix, mode); err != nil {
		return err
	}
	token := result.Token
	timer := time.NewTimer(debounceDelay)
	timer.Stop()
	defer timer.Stop()
	var due <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return nil
		case watchErr := <-watcher.errors:
			return watchErr
		case serveErr := <-served:
			served <- serveErr
			return fmt.Errorf("preview server stopped unexpectedly: %w", serveErr)
		case <-watcher.changes:
			timer.Reset(debounceDelay)
			due = timer.C
		case <-due:
			due = nil
			pages.mu.Lock()
			next, buildErr := build(ctx, request)
			if ctx.Err() != nil {
				pages.mu.Unlock()
				return nil
			}
			if next.Token == "" {
				closeErr := pages.root.Close()
				pages.root = nil
				pages.mu.Unlock()
				return errors.Join(fmt.Errorf("preview stopped: no verified usable output: %w", buildErr), closeErr)
			}
			previousPrefix := pages.prefix
			if next.Token != token {
				if err := pages.replaceRoot(filepath.Join(session.Directory(), "public")); err != nil {
					pages.mu.Unlock()
					return err
				}
				pages.prefix, token = next.Prefix, next.Token
			}
			pages.mu.Unlock()
			if pages.prefix != previousPrefix {
				if _, err := fmt.Fprintf(options.Output, "Preview URL changed: %s%s (%s).\n", address, pages.prefix, mode); err != nil {
					return errors.Join(buildErr, err)
				}
			}
			if err := reportWarnings(options.Errors, next.Warnings); err != nil {
				return errors.Join(buildErr, err)
			}
			if buildErr != nil {
				message := "rebuild failed; serving verified output: "
				if errors.Is(buildErr, output.ErrBusy) {
					message = "rebuild blocked; retry pending: "
					timer.Reset(retryDelay)
					due = timer.C
				}
				if _, err := fmt.Fprintln(options.Errors, message+buildErr.Error()); err != nil {
					return err
				}
			} else if _, err := fmt.Fprintf(options.Output, "Rebuilt %d page(s): %s%s (%s).\n", next.Pages, address, pages.prefix, mode); err != nil {
				return err
			}
		}
	}
}

func reportWarnings(writer io.Writer, warnings []string) error {
	for _, warning := range warnings {
		if _, err := fmt.Fprintln(writer, "warning: "+warning); err != nil {
			return err
		}
	}
	return nil
}
