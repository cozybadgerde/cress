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

// Create writes the starter site into dir, creating it if needed. It refuses to
// write into a non-empty directory unless force is set.
func Create(dir string, force bool) error {
	if !force {
		if err := requireEmpty(dir); err != nil {
			return err
		}
	}
	return fs.WalkDir(siteFS, builtinRoot, func(p string, d fs.DirEntry, err error) error {
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
		if err := os.WriteFile(dest, data, filePerm); err != nil {
			return fmt.Errorf("writing %s: %w", dest, err)
		}
		return nil
	})
}

// requireEmpty returns ErrExists when dir exists and holds any entries.
func requireEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("checking %s: %w", dir, err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("%s: %w", dir, ErrExists)
	}
	return nil
}
