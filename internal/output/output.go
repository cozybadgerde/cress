package output

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

const directoryMode = 0o755
const fileMode = 0o644

// File is one prepared output destination and the source that owns it.
type File struct {
	Path   string
	Source string
	Data   []byte
}

// Validate rejects unsafe destinations and competing output claims.
func Validate(files []File) error {
	claims := make(map[string]string)
	directories := make(map[string]string)
	fold := cases.Fold()
	for _, file := range files {
		if !fs.ValidPath(file.Path) || file.Path == "." || strings.ContainsAny(file.Path, "\\\x00") {
			return fmt.Errorf("%s: unsafe output path %q", file.Source, file.Path)
		}
		key := norm.NFD.String(fold.String(norm.NFD.String(file.Path)))
		if key == markerPath || strings.HasPrefix(key, markerPath+"/") {
			return fmt.Errorf("%s: output path %q conflicts with the reserved Cress ownership marker", file.Source, file.Path)
		}
		if other, exists := claims[key]; exists {
			return collision(file.Path, other, file.Source)
		}
		if other, exists := directories[key]; exists {
			return collision(file.Path, other, file.Source)
		}
		for parent := path.Dir(key); parent != "."; parent = path.Dir(parent) {
			if other, exists := claims[parent]; exists {
				return collision(parent, other, file.Source)
			}
			directories[parent] = file.Source
		}
		claims[key] = file.Source
	}
	return nil
}

func collision(destination, first, second string) error {
	return fmt.Errorf("output collision at %q between %s and %s", destination, first, second)
}

// Write locks a site and publishes a complete prepared set with verified ownership.
func Write(ctx context.Context, site string, files []File) (writeErr error) {
	if err := Validate(files); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	writer, err := Acquire(ctx, site)
	if err != nil {
		return err
	}
	defer func() {
		writeErr = errors.Join(writeErr, writer.Close())
	}()
	return writer.Write(ctx, files)
}

func writeFile(root *os.Root, file File) error {
	if parent := path.Dir(file.Path); parent != "." {
		if err := root.MkdirAll(parent, directoryMode); err != nil {
			return fmt.Errorf("%s: create directory for %s: %w", file.Source, file.Path, err)
		}
	}
	target, err := root.OpenFile(file.Path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		return fmt.Errorf("%s: create %s: %w", file.Source, file.Path, err)
	}
	_, writeErr := target.Write(file.Data)
	if err := errors.Join(writeErr, target.Close()); err != nil {
		return fmt.Errorf("%s: write %s: %w", file.Source, file.Path, err)
	}
	return nil
}
