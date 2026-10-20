package content

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Load prepares published pages and diagnostics from a site's local content.
func Load(ctx context.Context, site string, config Config) (pages []Page, warnings []string, loadErr error) {
	root, err := os.OpenRoot(site)
	if err != nil {
		return nil, nil, fmt.Errorf("open site directory: %w", err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("close site directory: %w", err))
		}
	}()
	info, err := root.Lstat("content")
	if err != nil {
		return nil, nil, fmt.Errorf("inspect content directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, fmt.Errorf("content must be a real directory, not a symlink or file")
	}
	loadErr = fs.WalkDir(root.FS(), "content", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("read %s: %w", name, walkErr)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: content symlinks are not supported", name)
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("%s: content must be a regular file", name)
		}
		if !strings.HasSuffix(name, ".md") {
			return nil
		}
		source, err := root.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		prepared, diagnostics, err := Prepare(strings.TrimPrefix(name, "content/"), source, config)
		warnings = append(warnings, diagnostics...)
		if err != nil {
			return err
		}
		if !prepared.Draft {
			pages = append(pages, prepared)
		}
		return nil
	})
	return pages, warnings, loadErr
}
