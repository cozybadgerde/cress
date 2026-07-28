// Package scaffold writes a starter cress site: a minimal cress.toml and a
// couple of content pages, enough for `cress build` to produce a site with the
// built-in theme. The starter files are embedded in the binary.
package scaffold

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:builtin
var siteFS embed.FS

const (
	builtinRoot = "builtin"
	dirPerm     = 0o755
	filePerm    = 0o644
)

// ErrExists is returned by Create when the target directory already contains
// files and force was not set.
var ErrExists = errors.New("target directory not empty")

// Create writes the starter site into dir, creating it if needed, and returns
// the starter files it left alone because something was already there (in slash
// form, relative to dir). It never overwrites, so scaffolding cannot destroy
// work, and it refuses a directory holding anything but dot-entries unless
// force is set.
func Create(dir string, force bool) ([]string, error) {
	if !force {
		if err := requireEmpty(dir); err != nil {
			return nil, err
		}
	}
	var skipped []string
	err := fs.WalkDir(siteFS, builtinRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(builtinRoot, p)
		if err != nil {
			return err
		}
		dest := filepath.Join(dir, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, dirPerm)
		}
		data, err := siteFS.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), dirPerm); err != nil {
			return err
		}
		wrote, err := writeNew(dest, data)
		if err != nil {
			return err
		}
		if !wrote {
			skipped = append(skipped, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return skipped, nil
}

// writeNew writes data to path unless a file is already there, reporting
// whether it wrote. The exclusive open is what makes `cress init --force` mean
// "proceed despite the non-empty directory" rather than "replace my content":
// an existing file is never truncated, not even between the check and the
// write.
func writeNew(path string, data []byte) (bool, error) {
	// #nosec G304 -- path is under the target directory the user named.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePerm)
	if errors.Is(err, fs.ErrExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("writing %s: %w", path, err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return false, fmt.Errorf("writing %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return false, fmt.Errorf("writing %s: %w", path, err)
	}
	return true, nil
}

// requireEmpty returns ErrExists when dir holds any entry that is not a
// dot-entry. Dotfiles do not count: `git init` before `cress init` is the
// natural way to start, and a lone .git must not push the user onto --force.
func requireEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("checking %s: %w", dir, err)
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			return fmt.Errorf("%s: %w", dir, ErrExists)
		}
	}
	return nil
}
