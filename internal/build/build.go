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
	Nav  []navLink
	Page pageView
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

// Build renders the site described by opts and returns a summary. It replaces
// the output directory's contents.
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

	if err := resetDir(outPath); err != nil {
		return nil, err
	}

	written := 0
	for _, page := range pages {
		if page.Draft && !opts.Drafts {
			continue
		}
		if err := writePage(outPath, thm, renderer, cfg.Site, nav, page); err != nil {
			return nil, err
		}
		written++
	}

	if err := copyStatic(outPath, thm.StaticFS(), filepath.Join(root, StaticDir)); err != nil {
		return nil, err
	}

	return &Result{Pages: written, Output: outPath, Warnings: warnings}, nil
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

// resolveNav maps each configured nav entry's content path to its page URL.
// Entries pointing at an unknown file are dropped and reported as warnings.
func resolveNav(items []config.NavItem, pages []*content.Page) ([]navLink, []string) {
	bySource := make(map[string]*content.Page, len(pages))
	for _, p := range pages {
		bySource[p.SourcePath] = p
	}
	var (
		links    []navLink
		warnings []string
	)
	for _, item := range items {
		page, ok := bySource[filepath.ToSlash(item.Path)]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("nav entry %q points at missing content %q", item.Title, item.Path))
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
func writePage(outPath string, thm *theme.Theme, renderer *render.Renderer, site config.Site, nav []navLink, page *content.Page) error {
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

// activeNav returns a copy of nav with the entry matching currentURL flagged.
func activeNav(nav []navLink, currentURL string) []navLink {
	out := make([]navLink, len(nav))
	copy(out, nav)
	for i := range out {
		out[i].Active = out[i].URL == currentURL
	}
	return out
}

// resetDir removes destDir and recreates it empty.
func resetDir(destDir string) error {
	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("clearing %s: %w", destDir, err)
	}
	if err := os.MkdirAll(destDir, outputDirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", destDir, err)
	}
	return nil
}

// copyStatic copies the theme's static assets, then the site's static/ tree,
// into destDir. Site files win on a name collision (copied last). Either source
// may be absent.
func copyStatic(destDir string, themeStatic fs.FS, siteStaticDir string) error {
	if themeStatic != nil {
		if err := copyFS(destDir, themeStatic); err != nil {
			return fmt.Errorf("copying theme assets: %w", err)
		}
	}
	if info, err := os.Stat(siteStaticDir); err == nil && info.IsDir() {
		if err := copyFS(destDir, os.DirFS(siteStaticDir)); err != nil {
			return fmt.Errorf("copying static: %w", err)
		}
	}
	return nil
}

// copyFS copies every file in srcFS into destDir, preserving relative paths.
func copyFS(destDir string, srcFS fs.FS) error {
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
		return copyFile(srcFS, p, dest)
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
