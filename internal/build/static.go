package build

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// copyStatic copies the theme's static assets, then the site's static/ tree,
// into destDir. Site files win on a name collision (copied last). Either source
// may be absent. Every copied path is recorded in written.
func copyStatic(destDir string, themeStatic fs.FS, siteStaticDir string, written map[string]bool) error {
	if themeStatic != nil {
		if err := copyFS(destDir, themeStatic, written); err != nil {
			return fmt.Errorf("copying theme assets: %w", err)
		}
	}
	if info, err := os.Stat(siteStaticDir); err == nil && info.IsDir() {
		if err := copyFS(destDir, os.DirFS(siteStaticDir), written); err != nil {
			return fmt.Errorf("copying static: %w", err)
		}
	}
	return nil
}

// copyFS copies every regular file in srcFS into destDir, preserving relative
// paths and recording each one in written.
//
// Only regular files are copied. Opening a symlink would follow it, and unlike
// the content tree there is no extension to narrow what that reaches, so
// static/id_rsa pointing anywhere readable would be published under that name.
// This guards the site's static directory and a theme's alike: an embedded
// theme cannot carry a link, but one in themes/ can.
func copyFS(destDir string, srcFS fs.FS, written map[string]bool) error {
	return fs.WalkDir(srcFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		dest := filepath.Join(destDir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(dest), outputDirPerm); err != nil {
			return err
		}
		if err := copyFile(srcFS, p, dest); err != nil {
			return err
		}
		written[p] = true
		return nil
	})
}

// copyFile copies a single file named src (in srcFS) to dest on disk.
func copyFile(srcFS fs.FS, src, dest string) error {
	in, err := srcFS.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dest) // #nosec G304 -- dest is under the output tree
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
