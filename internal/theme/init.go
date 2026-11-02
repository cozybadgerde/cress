package theme

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
)

const directoryMode = 0o755
const fileMode = 0o644

func sortedFiles(files map[string][]byte) []string { return slices.Sorted(maps.Keys(files)) }

// Init copies the embedded skeleton into a new, non-overwriting local package.
func Init(ctx context.Context, site, name string) (initErr error) {
	if !ValidName(name) {
		return fmt.Errorf("invalid theme name %q; use lowercase ASCII letters, digits, and hyphens, starting with a letter", name)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	root, err := os.OpenRoot(site)
	if err != nil {
		return fmt.Errorf("open theme initialization site: %w", err)
	}
	defer func() { initErr = errors.Join(initErr, root.Close()) }()
	info, err := root.Lstat("themes")
	if errors.Is(err, fs.ErrNotExist) {
		if err := root.Mkdir("themes", directoryMode); err != nil {
			return fmt.Errorf("create themes directory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("inspect themes: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("themes must be a real directory, not a symlink or file")
	}
	destination := path.Join("themes", name)
	if err := root.Mkdir(destination, directoryMode); err != nil {
		return fmt.Errorf("create new theme %s (existing targets are never overwritten): %w", name, err)
	}
	err = fs.WalkDir(embedded, "skeleton", func(source string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		relative := strings.TrimPrefix(source, "skeleton/")
		if source == "skeleton" {
			relative = "."
		}
		target := path.Join(destination, relative)
		if entry.IsDir() {
			if relative == "." {
				return nil
			}
			return root.Mkdir(target, directoryMode)
		}
		data, err := embedded.ReadFile(source)
		if err != nil {
			return err
		}
		if relative == "theme.json" {
			data, err = json.MarshalIndent(Manifest{Name: name, Version: "1.0.0", Cress: Compatibility{Minimum: developmentVersion, Major: 1}}, "", "  ")
			if err != nil {
				return err
			}
			data = append(data, '\n')
		}
		if relative == "templates/shared.gohtml" {
			data, err = insertStylesheet(data)
			if err != nil {
				return err
			}
		}
		file, err := root.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileMode)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(data)
		return errors.Join(writeErr, file.Close())
	})
	if err != nil {
		return fmt.Errorf("theme initialization incomplete at %s; inspect before retrying: %w", destination, err)
	}
	return nil
}

func insertStylesheet(source []byte) ([]byte, error) {
	const anchor = `{{block "theme-styles" .}}{{end}}`
	if count := strings.Count(string(source), anchor); count != 1 {
		return nil, fmt.Errorf("embedded templates/shared.gohtml: expected exactly one theme-styles anchor, found %d", count)
	}
	return []byte(strings.Replace(string(source), anchor,
		`<link rel="stylesheet" href="{{asset "site.css"}}">`, 1)), nil
}
