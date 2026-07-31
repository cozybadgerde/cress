package build

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cozybadgerde/cress/internal/config"
)

const outputDirPerm = 0o755

// staleWarningLimit caps how many stale files are named one by one. Past it a
// single summary line stands in, so a bulk rename cannot bury the rest of the
// build's output.
const staleWarningLimit = 10

// guardOutput refuses to use an output path that would clobber the site itself.
func guardOutput(root, outPath string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	absOut, err := filepath.Abs(outPath)
	if err != nil {
		return err
	}
	if absOut == absRoot {
		return fmt.Errorf("refusing to build into the site root %s", absRoot)
	}
	if absOut == filepath.Join(absRoot, config.ContentDir) {
		return fmt.Errorf("refusing to build into the content directory %s", absOut)
	}
	return nil
}

// ensureDir creates destDir if it does not exist, leaving any existing contents
// alone. The builder deliberately has no counterpart that clears it: no rule
// about which paths are safe to delete can hold across every machine, CI runner
// and platform, and not deleting needs no such rule. Removing output is
// `cress clean`'s job, where the user asking is the consent.
func ensureDir(destDir string) error {
	if err := os.MkdirAll(destDir, outputDirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", destDir, err)
	}
	return nil
}

// staleFiles reports files under outPath that this build did not write, in
// slash form relative to outPath. Because a build never deletes, a renamed or
// removed page leaves its old file behind; listing those is the only way the
// user finds out it is still being served.
//
// Dot-entries are skipped. A .git directory (building into a gh-pages worktree)
// or a hand-placed .nojekyll belongs to the user rather than to the build, and
// would otherwise be reported on every run until it was ignored out of habit.
//
// clean.collect walks the same tree and deliberately does the opposite, because
// `cress clean` means empty and a keep-list would grow with every host. The two
// are the halves of one policy: change the dotfile rule here and that is the
// other half to change with it.
func staleFiles(outPath string, written map[string]bool) ([]string, error) {
	var stale []string
	err := filepath.WalkDir(outPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != outPath && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(outPath, path)
		if err != nil {
			return err
		}
		if slash := filepath.ToSlash(rel); !written[slash] {
			stale = append(stale, slash)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", outPath, err)
	}
	return stale, nil
}

// staleWarnings turns stale output paths into build warnings, naming each file
// so it can be acted on, and collapsing to a count once there are too many to
// read.
func staleWarnings(outPath string, stale []string) []string {
	if len(stale) == 0 {
		return nil
	}
	if len(stale) > staleWarningLimit {
		return []string{fmt.Sprintf("%d file(s) in %s were not written by this build", len(stale), outPath)}
	}
	warnings := make([]string, 0, len(stale))
	for _, f := range stale {
		warnings = append(warnings, fmt.Sprintf("%s was not written by this build", filepath.Join(outPath, filepath.FromSlash(f))))
	}
	return warnings
}
