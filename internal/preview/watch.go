package preview

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type sourceWatcher struct {
	watcher *fsnotify.Watcher
	site    string
	changes chan struct{}
	errors  chan error
	done    chan struct{}
}

func watchesSource(relative string) bool {
	if relative == "cress.toml" {
		return true
	}
	for _, name := range []string{"content", "static", "themes"} {
		if relative == name || strings.HasPrefix(relative, name+"/") {
			return true
		}
	}
	return false
}

func newSourceWatcher(site string) (*sourceWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	source := &sourceWatcher{
		watcher: watcher, site: site, changes: make(chan struct{}, 1),
		errors: make(chan error, 1), done: make(chan struct{}),
	}
	if err := source.syncDirectories(); err != nil {
		return nil, errors.Join(err, watcher.Close())
	}
	return source, nil
}

func (source *sourceWatcher) syncDirectories() error {
	wanted := map[string]bool{source.site: true}
	for _, name := range []string{"content", "static", "themes"} {
		err := filepath.WalkDir(filepath.Join(source.site, name), func(name string, entry fs.DirEntry, err error) error {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			if err != nil {
				return err
			}
			if entry.IsDir() {
				wanted[name] = true
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("enumerate source watches: %w", err)
		}
	}
	current := make(map[string]bool)
	for _, name := range source.watcher.WatchList() {
		current[name] = true
		if !wanted[name] {
			if err := source.watcher.Remove(name); err != nil && !errors.Is(err, fsnotify.ErrNonExistentWatch) {
				return fmt.Errorf("remove source watch %s: %w", name, err)
			}
		}
	}
	for name := range wanted {
		if !current[name] {
			if err := source.watcher.Add(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("watch %s: %w", name, err)
			}
		}
	}
	return nil
}

func (source *sourceWatcher) run(ctx context.Context) {
	defer close(source.done)
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-source.watcher.Events:
			if !ok {
				source.errors <- fmt.Errorf("source watcher closed unexpectedly")
				return
			}
			relative, err := filepath.Rel(source.site, event.Name)
			if err != nil {
				source.errors <- err
				return
			}
			if relative == "." && event.Has(fsnotify.Remove|fsnotify.Rename) {
				source.errors <- fmt.Errorf("site directory moved or removed")
				return
			}
			if !watchesSource(filepath.ToSlash(relative)) {
				continue
			}
			if event.Has(fsnotify.Create | fsnotify.Remove | fsnotify.Rename) {
				if err := source.syncDirectories(); err != nil {
					source.errors <- err
					return
				}
			}
			select {
			case source.changes <- struct{}{}:
			default:
			}
		case err, ok := <-source.watcher.Errors:
			if !ok {
				err = fmt.Errorf("source watcher error channel closed unexpectedly")
			}
			source.errors <- fmt.Errorf("source watching failed: %w", err)
			return
		}
	}
}

func (source *sourceWatcher) close() error {
	err := source.watcher.Close()
	<-source.done
	return err
}
