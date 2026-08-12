// Package build is the site builder: it wires config, content, render, and
// theme together to turn a source tree (content + config + theme) into a
// directory of static HTML. It is the one place that performs output I/O.
package build

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/render"
	"github.com/cozybadgerde/cress/internal/theme"
	"github.com/cozybadgerde/cress/internal/version"
)

// NotFoundFile is the site's 404 page, written at the output root. Static hosts
// look for it there regardless of how the content tree nests, so it is the one
// output path that does not mirror its source.
const NotFoundFile = "404.html"

// yearToken is the placeholder in [site].copyright that stands in for the year
// the site is built, so a notice does not go stale every January.
const yearToken = "{year}"

// Options configures a build.
type Options struct {
	// Root is the site root directory (holds cress.toml). Defaults to ".".
	Root string
	// Output is the directory to write into. A relative path is taken under
	// Root; defaults to config.OutputDir.
	Output string
	// Drafts includes pages marked draft when true.
	Drafts bool
}

// Result summarizes a completed build.
type Result struct {
	Pages  int
	Output string
	// BasePath is the path the built site is rooted under, from the config's
	// base_url, or empty for a site at a domain root. The preview server reads it
	// to mount the output where the emitted links expect to find it.
	BasePath string
	Warnings []string
}

// Build renders the site described by opts and returns a summary. It only ever
// creates and overwrites files: nothing in the output directory is deleted, so
// a build can never destroy content it did not produce. Files left over from an
// earlier build are reported as warnings instead.
func Build(opts Options) (*Result, error) {
	root, outPath, err := resolvePaths(opts)
	if err != nil {
		return nil, err
	}
	in, err := loadInputs(root)
	if err != nil {
		return nil, err
	}

	base := in.cfg.Site.BasePath
	nav, warnings := resolveNav(in.cfg.Nav, in.pages, base)
	warnings = append(warnings, shadowWarnings(root, in.cfg.Site.Theme)...)
	warnings = append(warnings, themeWarnings(in.thm)...)
	writer := &pageWriter{
		outPath:  outPath,
		thm:      in.thm,
		renderer: render.New(renderOptions(in.cfg, base)...),
		site:     siteData(in.cfg.Site, base),
		nav:      nav,
		basePath: base,
	}

	if err := ensureDir(outPath); err != nil {
		return nil, err
	}
	rendered, written, err := writeAll(writer, in.pages, opts.Drafts)
	if err != nil {
		return nil, err
	}
	if err := copyStatic(outPath, in.thm.StaticFS(), filepath.Join(root, config.StaticDir), written); err != nil {
		return nil, err
	}

	stale, err := staleFiles(outPath, written)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, writer.warnings...)
	return &Result{
		Pages:    rendered,
		Output:   outPath,
		BasePath: base,
		Warnings: append(warnings, staleWarnings(outPath, stale)...),
	}, nil
}

// resolvePaths settles where the build reads from and writes to, and refuses an
// output path that would clobber the site itself. It runs first, because every
// later phase is named relative to what it returns.
func resolvePaths(opts Options) (root, outPath string, err error) {
	root = opts.Root
	if root == "" {
		root = "."
	}
	outName := opts.Output
	if outName == "" {
		outName = config.OutputDir
	}
	outPath = outName
	if !filepath.IsAbs(outName) {
		outPath = filepath.Join(root, outName)
	}
	if err := guardOutput(root, outPath); err != nil {
		return "", "", err
	}
	return root, outPath, nil
}

// renderOptions turns the site's config into the renderer's options, rooting
// the author's links under base.
//
// An option left off is what leaves the renderer as it was, so a site that
// configures nothing gets the goldmark cress has always built rather than one
// assembled from defaults that happen to agree with it.
func renderOptions(cfg *config.Config, base string) []render.Option {
	opts := []render.Option{render.WithBasePath(base)}
	if cfg.Markdown.Highlight {
		opts = append(opts, render.WithHighlighting())
	}
	return opts
}

// inputs is the three sources a build reads. They are loaded together, before
// anything is written, so a site that cannot be built fails without leaving a
// half-written output tree behind.
type inputs struct {
	cfg   *config.Config
	pages []*content.Page
	thm   *theme.Theme
}

// loadInputs reads the config, the content tree, and the theme for the site
// rooted at root.
func loadInputs(root string) (*inputs, error) {
	cfg, err := config.Load(filepath.Join(root, config.FileName))
	if err != nil {
		return nil, err
	}
	pages, err := content.Collect(filepath.Join(root, config.ContentDir))
	if err != nil {
		return nil, fmt.Errorf("collecting content: %w", err)
	}
	thm, err := theme.Resolve(root, config.ThemesDir, cfg.Site.Theme)
	if err != nil {
		return nil, err
	}
	return &inputs{cfg: cfg, pages: pages, thm: thm}, nil
}

// themeWarnings reports what the theme's own metadata has to say to the cress
// running it, each line prefixed with the theme it came from so it reads like
// every other build warning.
//
// The running version enters the build here, which is what keeps internal/theme
// a pure function of the directory it reads: the comparison takes the version as
// an argument rather than reaching for the package that holds it.
func themeWarnings(thm *theme.Theme) []string {
	found := thm.Meta().Warnings(version.Version)
	warnings := make([]string, 0, len(found))
	for _, w := range found {
		warnings = append(warnings, fmt.Sprintf("theme %q: %s", thm.Name(), w))
	}
	return warnings
}

// shadowWarnings reports a themes/ directory that has taken a built-in theme's
// name. The directory wins, which is the documented rule and stays the rule;
// what is worth saying is that it happened at all.
//
// Nothing in a built page reveals which of the two rendered it, so the silent
// version of this is a site whose author reads one theme and publishes another.
// The likeliest way in is `cress theme init <name>` with a name already in use,
// which scaffolds the starter theme over a design somebody chose.
func shadowWarnings(root, name string) []string {
	if !theme.ShadowsBuiltin(root, config.ThemesDir, name) {
		return nil
	}
	return []string{fmt.Sprintf(
		"theme %q was loaded from %s, which shadows the built-in theme of the same name; rename the directory to use the built-in one",
		name, filepath.Join(config.ThemesDir, name))}
}

// siteData prepares the site metadata every page is rendered with: the branding
// URLs rooted under base, and the copyright year expanded.
//
// The clock enters the build here and nowhere else. config.Load stays a pure
// function of the file it reads, and every page in one build shares a year.
func siteData(site config.Site, base string) config.Site {
	site.Copyright = expandYear(site.Copyright, time.Now().Year())
	site.Logo = prefixURL(base, site.Logo)
	site.LogoDark = prefixURL(base, site.LogoDark)
	site.Favicon = prefixURL(base, site.Favicon)
	return site
}

// writeAll renders every page that is not a draft and then guarantees a 404,
// reporting how many pages the author wrote and which output paths this build
// owns. That second return is what tells staleFiles which files are not ours.
func writeAll(w *pageWriter, pages []*content.Page, drafts bool) (int, map[string]bool, error) {
	written := make(map[string]bool, len(pages))
	rendered := 0
	for _, page := range pages {
		if page.Draft && !drafts {
			continue
		}
		if err := w.writePage(page); err != nil {
			return 0, nil, err
		}
		written[page.OutputPath] = true
		rendered++
	}

	// content/404.md, having produced NotFoundFile like any other page, wins.
	// Only when it did not is a 404 synthesized, and it is not counted in Pages:
	// it is not a page the author wrote.
	if !written[NotFoundFile] {
		if err := w.writeNotFound(); err != nil {
			return 0, nil, err
		}
		written[NotFoundFile] = true
	}
	return rendered, written, nil
}

// prefixURL roots u under basePath, leaving anything that is not a link into
// this site's own root untouched: a relative path, an absolute URL, and a
// protocol-relative URL all address something the base path does not govern.
// This is the same rule the renderer applies to links inside content, kept here
// for the URLs the builder emits itself.
func prefixURL(basePath, u string) string {
	if basePath == "" || !strings.HasPrefix(u, "/") || strings.HasPrefix(u, "//") {
		return u
	}
	return basePath + u
}

// expandYear replaces yearToken in a copyright notice with year. It takes the
// year rather than reading the clock so that the substitution stays pure and
// the caller decides what "now" means.
func expandYear(copyright string, year int) string {
	return strings.ReplaceAll(copyright, yearToken, strconv.Itoa(year))
}
