package preview

import (
	"context"
	"testing"
	"time"
)

func TestWatchingIgnoresOutputAndKeepsSourceEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping native filesystem event integration test")
	}
	site := t.TempDir()
	writeSource(t, site, "content/index.md", "initial")
	watcher, err := newSourceWatcher(site)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go watcher.run(ctx)
	defer func() {
		cancel()
		if err := watcher.close(); err != nil {
			t.Error(err)
		}
	}()
	writeSource(t, site, "public/index.html", "generated")
	writeSource(t, site, ".cress/output.json", "state")
	select {
	case <-watcher.changes:
		t.Fatal("generated output triggered a rebuild")
	case err := <-watcher.errors:
		t.Fatal(err)
	case <-time.After(300 * time.Millisecond):
	}
	writeSource(t, site, "content/nested/page.md", "source")
	select {
	case <-watcher.changes:
	case err := <-watcher.errors:
		t.Fatal(err)
	case <-time.After(testTimeout):
		t.Fatal("new nested source did not trigger a rebuild")
	}
}
