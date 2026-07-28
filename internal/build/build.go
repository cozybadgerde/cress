// Package build is the site builder: it wires config, content, render, and
// theme together to turn a source tree (content + config + theme) into a
// directory of static HTML. It is the one place that performs output I/O.
package build

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/render"
	"github.com/cozybadgerde/cress/internal/theme"
)

// Conventional site directory names, relative to the site root.
const (
	ContentDir = "content"
	StaticDir  = "static"
	ThemesDir  = "themes"
	OutputDir  = "public"
)

const outputDirPerm = 0o755

// staleWarningLimit caps how many stale files are named one by one. Past it a
// single summary line stands in, so a bulk rename cannot bury the rest of the
// build's output.
const staleWarningLimit = 10

// Options configures a build.
type Options struct {
	// Root is the site root directory (holds cress.toml). Defaults to ".".
	Root string
	// Output is the directory to write into. A relative path is taken under
	// Root; defaults to OutputDir.
	Output string
	// Drafts includes pages marked draft when true.
	Drafts bool
}

// Result summarizes a completed build.
type Result struct {
	Pages    int
	Output   string
	Warnings []string
}

// renderContext is the data passed to a theme's template for one page.
type renderContext struct {
	Site config.Site
	Nav  navView
	Page pageView
}

// navView is the resolved navigation as a template sees it: one slice per menu,
// so a theme can place the primary and secondary links independently.
type navView struct {
	Main   []navLink
	Footer []navLink
}

// navLink is one resolved navigation entry; Active marks the current page.
type navLink struct {
	Title  string
	URL    string
	Active bool
}

// pageView is a page as seen by a template: its metadata plus rendered HTML.
type pageView struct {
	Title string
	URL   string
	Meta  map[string]any
	HTML  template.HTML
}

// Build renders the site described by opts and returns a summary. It only ever
// creates and overwrites files: nothing in the output directory is deleted, so
// a build can never destroy content it did not produce. Files left over from an
// earlier build are reported as warnings instead.
func Build(opts Options) (*Result, error) {
	root := opts.Root
	if root == "" {
		root = "."
	}
	outName := opts.Output
	if outName == "" {
		outName = OutputDir
	}
	outPath := outName
	if !filepath.IsAbs(outName) {
		outPath = filepath.Join(root, outName)
	}
	if err := guardOutput(root, outPath); err != nil {
		return nil, err
	}

	cfg, err := config.Load(filepath.Join(root, config.FileName))
	if err != nil {
		return nil, err
	}
	pages, err := content.Collect(filepath.Join(root, ContentDir))
	if err != nil {
		return nil, fmt.Errorf("collecting content: %w", err)
	}
	thm, err := theme.Resolve(root, ThemesDir, cfg.Site.Theme)
	if err != nil {
		return nil, err
	}

	nav, warnings := resolveNav(cfg.Nav, pages)
	renderer := render.New()

	if err := ensureDir(outPath); err != nil {
		return nil, err
	}

	written := make(map[string]bool, len(pages))
	rendered := 0
	for _, page := range pages {
		if page.Draft && !opts.Drafts {
			continue
		}
		if err := writePage(outPath, thm, renderer, cfg.Site, nav, page); err != nil {
			return nil, err
		}
		written[page.OutputPath] = true
		rendered++
	}

	if err := copyStatic(outPath, thm.StaticFS(), filepath.Join(root, StaticDir), written); err != nil {
		return nil, err
	}

	stale, err := staleFiles(outPath, written)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, staleWarnings(outPath, stale)...)

	return &Result{Pages: rendered, Output: outPath, Warnings: warnings}, nil
}

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
	if absOut == filepath.Join(absRoot, ContentDir) {
		return fmt.Errorf("refusing to build into the content directory %s", absOut)
	}
	return nil
}

// resolveNav resolves every navigation group against the collected pages.
func resolveNav(nav config.Nav, pages []*content.Page) (navView, []string) {
	bySource := make(map[string]*content.Page, len(pages))
	for _, p := range pages {
		bySource[p.SourcePath] = p
	}
	main, mainWarnings := resolveNavGroup("nav.main", nav.Main, bySource)
	footer, footerWarnings := resolveNavGroup("nav.footer", nav.Footer, bySource)
	return navView{Main: main, Footer: footer}, append(mainWarnings, footerWarnings...)
}

// resolveNavGroup maps one group's entries to their page URLs. Entries pointing
// at an unknown file are dropped and reported as warnings naming the group, so
// the rest of the menu still renders.
func resolveNavGroup(group string, items []config.NavItem, bySource map[string]*content.Page) ([]navLink, []string) {
	var (
		links    []navLink
		warnings []string
	)
	for _, item := range items {
		page, ok := bySource[filepath.ToSlash(item.Path)]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("%s entry %q points at missing content %q", group, item.Title, item.Path))
			continue
		}
		title := item.Title
		if title == "" {
			title = page.Title
		}
		links = append(links, navLink{Title: title, URL: page.URL})
	}
	return links, warnings
}

// writePage renders one page through the theme and writes it to the output tree.
func writePage(outPath string, thm *theme.Theme, renderer *render.Renderer, site config.Site, nav navView, page *content.Page) error {
	body, err := renderer.Markdown(page.Body)
	if err != nil {
		return fmt.Errorf("%s: %w", page.SourcePath, err)
	}

	ctx := renderContext{
		Site: site,
		Nav:  activeNav(nav, page.URL),
		Page: pageView{Title: page.Title, URL: page.URL, Meta: page.Meta, HTML: body},
	}

	var buf bytes.Buffer
	if err := thm.Render(&buf, ctx); err != nil {
		return fmt.Errorf("%s: %w", page.SourcePath, err)
	}

	dest := filepath.Join(outPath, filepath.FromSlash(page.OutputPath))
	if err := os.MkdirAll(filepath.Dir(dest), outputDirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(dest), err)
	}
	if err := os.WriteFile(dest, buf.Bytes(), 0o644); err != nil { // #nosec G306 -- public site files
		return fmt.Errorf("writing %s: %w", dest, err)
	}
	return nil
}

// activeNav returns a copy of nav with the entries matching currentURL flagged,
// in every group: a footer link to the current page is current too.
func activeNav(nav navView, currentURL string) navView {
	return navView{
		Main:   activeLinks(nav.Main, currentURL),
		Footer: activeLinks(nav.Footer, currentURL),
	}
}

// activeLinks returns a copy of links with the entry matching currentURL flagged.
func activeLinks(links []navLink, currentURL string) []navLink {
	out := make([]navLink, len(links))
	copy(out, links)
	for i := range out {
		out[i].Active = out[i].URL == currentURL
	}
	return out
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

// copyFS copies every file in srcFS into destDir, preserving relative paths and
// recording each one in written.
func copyFS(destDir string, srcFS fs.FS, written map[string]bool) error {
	return fs.WalkDir(srcFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
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
