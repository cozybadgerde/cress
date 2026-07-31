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
// files and allowExisting was not set.
var ErrExists = errors.New("target directory not empty")

// Create writes the starter site into dir, creating it if needed, and returns
// the starter files it left alone because something was already there (in slash
// form, relative to dir).
//
// It never overwrites, so scaffolding cannot destroy work, and that holds
// whatever allowExisting says. All that flag decides is whether a directory
// already holding files is refused: it relaxes a precondition rather than
// granting permission to write over anything.
func Create(dir string, allowExisting bool) ([]string, error) {
	if !allowExisting {
		if err := requireEmpty(dir); err != nil {
			return nil, err
		}
	}
	w := &siteWriter{dir: dir}
	if err := fs.WalkDir(siteFS, builtinRoot, w.copyEntry); err != nil {
		return nil, err
	}
	return w.skipped, nil
}

// siteWriter copies the embedded starter site into a target directory,
// collecting the files it left alone. The accumulator is what earns it a type:
// as a closure over a local it shared a scope with the walk itself, so a reader
// tracked the traversal, the per-entry work, and the result at once.
type siteWriter struct {
	dir     string
	skipped []string
}

// copyEntry writes one entry of the embedded tree into the target directory. It
// is an fs.WalkDirFunc, so it reports a name it skipped rather than returning
// it: a starter file that is already there is left untouched and recorded.
func (w *siteWriter) copyEntry(p string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(builtinRoot, p)
	if err != nil {
		return err
	}
	dest := filepath.Join(w.dir, rel)
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
		w.skipped = append(w.skipped, filepath.ToSlash(rel))
	}
	return nil
}

// writeNew writes data to path unless a file is already there, reporting
// whether it wrote. The exclusive open is what makes `cress init
// --allow-existing` mean "scaffold alongside what is here" rather than "replace
// it": an existing file is never truncated, not even between the check and the
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
// natural way to start, and a lone .git must not push the user onto a flag.
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
