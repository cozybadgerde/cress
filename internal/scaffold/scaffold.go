package scaffold

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cozybadgerde/cress/internal/theme"
)

//go:embed site
var sources embed.FS

// Init creates a ready-to-build site without adopting or overwriting files.
func Init(ctx context.Context, directory, name string) (initErr error) {
	if !theme.IsBuiltin(name) {
		return fmt.Errorf("site init theme must be poppy or clover, got %q", name)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	target, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	info, err := os.Lstat(target)
	if errors.Is(err, fs.ErrNotExist) {
		parent, err := filepath.EvalSymlinks(filepath.Dir(target))
		if err != nil {
			return fmt.Errorf("resolve existing site parent: %w", err)
		}
		target = filepath.Join(parent, filepath.Base(target))
		if err := os.Mkdir(target, 0o755); err != nil {
			return fmt.Errorf("create site directory: %w", err)
		}
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("site init target must be a real directory, not a symlink or file")
	}
	root, err := os.OpenRoot(target)
	if err != nil {
		return err
	}
	defer func() { initErr = errors.Join(initErr, root.Close()) }()
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("site init refuses nonempty directory %s; choose a new or empty target", target)
	}
	err = fs.WalkDir(sources, "site", func(source string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if source == "site" {
			return nil
		}
		relative := strings.TrimPrefix(source, "site/")
		if relative == "gitignore" {
			relative = ".gitignore"
		}
		if entry.IsDir() {
			return root.Mkdir(relative, 0o755)
		}
		data, err := sources.ReadFile(source)
		if err != nil {
			return err
		}
		if relative == "cress.toml" {
			data = []byte(strings.Replace(string(data), `name = "poppy"`, `name = "`+name+`"`, 1))
		}
		file, err := root.OpenFile(relative, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(data)
		return errors.Join(writeErr, file.Close())
	})
	if err != nil {
		return fmt.Errorf("site initialization incomplete at %s; inspect before retrying: %w", target, err)
	}
	return nil
}
