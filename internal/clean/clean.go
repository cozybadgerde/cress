// Package clean empties a built site's output directory. It is the only
// package that deletes, and it acts only when a user names the command:
// build never removes anything, so nothing can reach this code unattended.
//
// Deleting is the whole point here, which is why the guards are stricter than
// the ones build applies to the same path. A build that lands somewhere
// unexpected leaves files to tidy up; a clean that does destroys sources.
package clean

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cozybadgerde/cress/internal/config"
)

// Options configures a clean.
type Options struct {
	// Root is the site root directory (holds cress.toml). Defaults to ".".
	Root string
	// Output is the directory to empty. A relative path is taken under Root;
	// defaults to config.OutputDir.
	Output string
}

// Plan is what a clean would remove: resolved, checked, and not yet carried
// out. Preparing and executing are separate steps so that a dry run and a
// confirmation prompt describe exactly the work the deletion then performs,
// rather than an estimate of it.
type Plan struct {
	// Output is the absolute path of the directory to empty.
	Output string
	// Files lists every file that would go, relative to Output and sorted. It is
	// what a dry run prints and what the count reports, because a file count is
	// what an author recognizes their site by.
	Files []string
	// entries are the top-level names to remove, which is where deletion actually
	// happens: removing a directory takes everything under it.
	entries []string
}

// Empty reports whether there is nothing to remove, either because the output
// directory holds no entries or because it does not exist. A clean is then a
// no-op rather than an error, so running it twice is safe.
func (p *Plan) Empty() bool { return len(p.entries) == 0 }

// Prepare resolves the output directory, refuses it if removing it would
// destroy the site, and collects what a clean would take. It deletes nothing.
func Prepare(opts Options) (*Plan, error) {
	root := opts.Root
	if root == "" {
		root = "."
	}
	outName := opts.Output
	if outName == "" {
		outName = config.OutputDir
	}
	outPath := outName
	if !filepath.IsAbs(outName) {
		outPath = filepath.Join(root, outName)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	absOut, err := filepath.Abs(outPath)
	if err != nil {
		return nil, err
	}
	if err := guard(absRoot, absOut); err != nil {
		return nil, err
	}

	files, entries, err := collect(absOut)
	if err != nil {
		return nil, err
	}
	return &Plan{Output: absOut, Files: files, entries: entries}, nil
}

// Execute removes every top-level entry the plan found. The output directory
// itself stays, so the next build writes into it without recreating it.
func (p *Plan) Execute() error {
	for _, name := range p.entries {
		if err := os.RemoveAll(filepath.Join(p.Output, name)); err != nil {
			return fmt.Errorf("removing %s: %w", filepath.Join(p.Output, name), err)
		}
	}
	return nil
}

// guard refuses an output path whose removal would take the site with it. Each
// message names what the path is rather than what was about to happen to it, so
// a caller can put its own verb in front.
func guard(absRoot, absOut string) error {
	// A directory without a config file is not a site, which is what catches a
	// clean run from the wrong directory: an unrelated "public" next door is far
	// likelier than a real site with no cress.toml. Nothing here reads the file,
	// so a site whose config is broken can still be cleaned.
	if _, err := os.Stat(filepath.Join(absRoot, config.FileName)); err != nil {
		return fmt.Errorf("%s is not a cress site: no %s", absRoot, config.FileName)
	}
	if absOut == absRoot {
		return fmt.Errorf("output path %s is the site root", absOut)
	}
	if within(absOut, absRoot) {
		return fmt.Errorf("output path %s contains the site root", absOut)
	}
	for _, dir := range []string{config.ContentDir, config.StaticDir, config.ThemesDir} {
		srcDir := filepath.Join(absRoot, dir)
		if absOut == srcDir || within(srcDir, absOut) {
			return fmt.Errorf("output path %s is inside the %s directory", absOut, dir)
		}
	}
	return nil
}

// within reports whether path lies strictly under dir.
func within(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, "..")
}

// collect lists every file under outPath and the top-level entries holding
// them. A missing output directory yields nothing, not an error: there is
// simply nothing to clean.
//
// Dot-entries are included, unlike in build.staleFiles, which walks the same
// tree and skips them because a .nojekyll the user placed is not the build's to
// report. Here the opposite holds: clean means empty, and a keep-list is a
// promise that grows with every host and never shrinks. The two are the halves
// of one policy, so a change to either wants a look at the other.
func collect(outPath string) (files, entries []string, err error) {
	dirEntries, err := os.ReadDir(outPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("reading %s: %w", outPath, err)
	}
	for _, entry := range dirEntries {
		entries = append(entries, entry.Name())
	}

	err = filepath.WalkDir(outPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(outPath, path)
		if relErr != nil {
			return relErr
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walking %s: %w", outPath, err)
	}
	sort.Strings(files)
	return files, entries, nil
}
