// Package build is the site builder: it wires config, content, render, and
// theme together to turn a source tree (content + config + theme) into a
// directory of static HTML. It is the one place that performs output I/O.
package build

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/render"
	"github.com/cozybadgerde/cress/internal/theme"
)

// NotFoundFile is the site's 404 page, written at the output root. Static hosts
// look for it there regardless of how the content tree nests, so it is the one
// output path that does not mirror its source.
const NotFoundFile = "404.html"

// The synthesized 404, used when neither the content tree nor the theme
// supplies one. It goes through the theme's entry template, so a theme that
// predates cress having a 404 still gets one carrying its nav and styling.
// homeURL is the URL content gives the site's own index page, before any base
// path is applied. It is what marks a page as the home page for a theme.
const homeURL = "/"

const (
	notFoundTemplate = "404.html"
	notFoundURL      = "/404.html"
	notFoundTitle    = "Page not found"
	notFoundBody     = "# Page not found\n\n" +
		"There is nothing at this address. The page may have been renamed or " +
		"removed, or the link that brought you here may be out of date.\n\n" +
		"[Back to the home page](/)\n"
)

const outputDirPerm = 0o755

// staleWarningLimit caps how many stale files are named one by one. Past it a
// single summary line stands in, so a bulk rename cannot bury the rest of the
// build's output.
const staleWarningLimit = 10

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
	if err := guardOutput(root, outPath); err != nil {
		return nil, err
	}

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

	base := cfg.Site.BasePath
	nav, warnings := resolveNav(cfg.Nav, pages, base)
	renderer := render.New(render.WithBasePath(base))

	// The clock enters the build here and nowhere else: config.Load stays a pure
	// function of the file it reads, and every page in one build shares a year.
	site := cfg.Site
	site.Copyright = expandYear(site.Copyright, time.Now().Year())
	site.Logo = prefixURL(base, site.Logo)
	site.LogoDark = prefixURL(base, site.LogoDark)
	site.Favicon = prefixURL(base, site.Favicon)

	if err := ensureDir(outPath); err != nil {
		return nil, err
	}

	writer := &pageWriter{
		outPath:  outPath,
		thm:      thm,
		renderer: renderer,
		site:     site,
		nav:      nav,
		basePath: base,
	}

	written := make(map[string]bool, len(pages))
	rendered := 0
	for _, page := range pages {
		if page.Draft && !opts.Drafts {
			continue
		}
		if err := writer.writePage(page); err != nil {
			return nil, err
		}
		written[page.OutputPath] = true
		rendered++
	}

	// content/404.md, having produced NotFoundFile like any other page, wins.
	// Only when it did not is a 404 synthesized, and it is not counted in Pages:
	// it is not a page the author wrote.
	if !written[NotFoundFile] {
		if err := writer.writeNotFound(); err != nil {
			return nil, err
		}
		written[NotFoundFile] = true
	}

	if err := copyStatic(outPath, thm.StaticFS(), filepath.Join(root, config.StaticDir), written); err != nil {
		return nil, err
	}

	stale, err := staleFiles(outPath, written)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, staleWarnings(outPath, stale)...)

	return &Result{Pages: rendered, Output: outPath, BasePath: base, Warnings: warnings}, nil
}

// absoluteURL joins the site's base URL with a page's rooted URL. It yields ""
// when the site sets no base_url: without a host there is no absolute form, and
// an empty value lets the caller omit what needs one rather than emit a link to
// nowhere.
func (w *pageWriter) absoluteURL(pageURL string) string {
	if w.site.BaseURL == "" {
		return ""
	}
	// BaseURL carries no trailing slash and pageURL is rooted, as config and
	// content guarantee, so joining them cannot double or drop a separator. The
	// base path is already in pageURL, and BaseURL ends with it, so trim it back
	// off to avoid naming it twice.
	return strings.TrimSuffix(w.site.BaseURL, w.basePath) + pageURL
}

// fallback returns value when it has one, else the site-wide alternative.
func fallback(value, siteWide string) string {
	if value != "" {
		return value
	}
	return siteWide
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

// resolveNav resolves every navigation group against the collected pages.
func resolveNav(nav config.Nav, pages []*content.Page, basePath string) (theme.NavView, []string) {
	bySource := make(map[string]*content.Page, len(pages))
	for _, p := range pages {
		bySource[p.SourcePath] = p
	}
	main, mainWarnings := resolveNavGroup("nav.main", nav.Main, bySource, basePath)
	footer, footerWarnings := resolveNavGroup("nav.footer", nav.Footer, bySource, basePath)
	return theme.NavView{Main: main, Footer: footer}, append(mainWarnings, footerWarnings...)
}

// resolveNavGroup maps one group's entries to their page URLs. Entries pointing
// at an unknown file are dropped and reported as warnings naming the group, so
// the rest of the menu still renders.
func resolveNavGroup(group string, items []config.NavItem, bySource map[string]*content.Page, basePath string) ([]theme.NavLink, []string) {
	var (
		links    []theme.NavLink
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
		links = append(links, theme.NavLink{Title: title, URL: prefixURL(basePath, page.URL)})
	}
	return links, warnings
}

// pageWriter holds what every page render shares: the theme, the renderer, and
// the site-wide values resolved once at the start of the build. Only the page
// itself changes between calls.
type pageWriter struct {
	outPath  string
	thm      *theme.Theme
	renderer *render.Renderer
	site     config.Site
	nav      theme.NavView
	basePath string
}

// writePage renders one page through the theme and writes it to the output tree.
func (w *pageWriter) writePage(page *content.Page) error {
	body, err := w.renderer.Markdown(page.Body)
	if err != nil {
		return fmt.Errorf("%s: %w", page.SourcePath, err)
	}

	pageURL := prefixURL(w.basePath, page.URL)
	absoluteURL := w.absoluteURL(pageURL)
	description := fallback(page.Description, w.site.Description)

	head, err := renderHead(headData{Description: description, Canonical: absoluteURL})
	if err != nil {
		return fmt.Errorf("%s: %w", page.SourcePath, err)
	}

	data := theme.PageData{
		Site: w.site,
		Nav:  activeNav(w.nav, pageURL),
		Page: theme.PageView{
			Title:       page.Title,
			Description: description,
			Language:    fallback(page.Language, w.site.Language),
			URL:         pageURL,
			AbsoluteURL: absoluteURL,
			IsHome:      page.URL == homeURL,
			Meta:        page.Meta,
			HTML:        body,
			Head:        head,
		},
	}

	var buf bytes.Buffer
	if err := w.thm.Render(&buf, data); err != nil {
		return fmt.Errorf("%s: %w", page.SourcePath, err)
	}

	dest := filepath.Join(w.outPath, filepath.FromSlash(page.OutputPath))
	if err := os.MkdirAll(filepath.Dir(dest), outputDirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(dest), err)
	}
	if err := os.WriteFile(dest, buf.Bytes(), 0o644); err != nil { // #nosec G306 -- public site files
		return fmt.Errorf("writing %s: %w", dest, err)
	}
	return nil
}

// writeNotFound writes the site's 404 page at the output root. A theme's own
// 404.html template renders it when there is one; otherwise the entry template
// does, which is what makes a working 404 free for every theme rather than a
// second required template alongside page.html.
func (w *pageWriter) writeNotFound() error {
	body, err := w.renderer.Markdown([]byte(notFoundBody))
	if err != nil {
		return fmt.Errorf("%s: %w", NotFoundFile, err)
	}

	// No canonical link: a page that stands in for every address that does not
	// exist cannot name one of them as its own. It is kept out of search results
	// instead, which is the only crawl decision cress makes for the author.
	head, err := renderHead(headData{Description: w.site.Description, NoIndex: true})
	if err != nil {
		return fmt.Errorf("%s: %w", NotFoundFile, err)
	}

	pageURL := prefixURL(w.basePath, notFoundURL)
	data := theme.PageData{
		Site: w.site,
		Nav:  activeNav(w.nav, pageURL),
		Page: theme.PageView{
			Title:       notFoundTitle,
			Description: w.site.Description,
			Language:    w.site.Language,
			URL:         pageURL,
			HTML:        body,
			Head:        head,
		},
	}

	var buf bytes.Buffer
	if w.thm.HasTemplate(notFoundTemplate) {
		err = w.thm.RenderTemplate(&buf, notFoundTemplate, data)
	} else {
		err = w.thm.Render(&buf, data)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", NotFoundFile, err)
	}

	dest := filepath.Join(w.outPath, NotFoundFile)
	if err := os.WriteFile(dest, buf.Bytes(), 0o644); err != nil { // #nosec G306 -- public site files
		return fmt.Errorf("writing %s: %w", dest, err)
	}
	return nil
}

// activeNav returns a copy of nav with the entries matching currentURL flagged,
// in every group: a footer link to the current page is current too.
func activeNav(nav theme.NavView, currentURL string) theme.NavView {
	return theme.NavView{
		Main:   activeLinks(nav.Main, currentURL),
		Footer: activeLinks(nav.Footer, currentURL),
	}
}

// activeLinks returns a copy of links with the entry matching currentURL flagged.
func activeLinks(links []theme.NavLink, currentURL string) []theme.NavLink {
	out := make([]theme.NavLink, len(links))
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
