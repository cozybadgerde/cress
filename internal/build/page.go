package build

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/render"
	"github.com/cozybadgerde/cress/internal/theme"
)

// homeURL is the URL content gives the site's own index page, before any base
// path is applied. It is what marks a page as the home page for a theme.
const homeURL = "/"

// htmlExt is the extension every theme template carries, which is what turns a
// front-matter layout name into the file a theme defines it in.
const htmlExt = ".html"

// The synthesized 404, used when neither the content tree nor the theme
// supplies one. It goes through the theme's entry template, so a theme that
// predates cress having a 404 still gets one carrying its nav and styling.
const (
	notFoundTemplate = "404.html"
	notFoundURL      = "/404.html"
	notFoundTitle    = "Page not found"
	notFoundBody     = "# Page not found\n\n" +
		"There is nothing at this address. The page may have been renamed or " +
		"removed, or the link that brought you here may be out of date.\n\n" +
		"[Back to the home page](/)\n"
)

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

// leadImage resolves the page's lead image with the alt text and caption that
// belong to it, rooted under the base path.
//
// The three are chosen together rather than one key at a time, which is what
// separates this from description and language above. Resolving them
// separately would let a page that named its own image inherit the site
// image's caption, attaching an attribution or an AI disclosure to a picture it
// was never written about. A missing credit is a gap; a credit on the wrong
// image is a false claim, and the second is the one worth designing out.
//
// The page's own path is rooted here while the site-wide one arrives rooted
// already, since siteData prefixed it once for the whole build.
func (w *pageWriter) leadImage(page *content.Page) (url, alt, caption string) {
	if page.Image != "" {
		return prefixURL(w.basePath, page.Image), page.ImageAlt, page.ImageCaption
	}
	return w.site.Image, w.site.ImageAlt, w.site.ImageCaption
}

// pageWriter holds what every page render shares: the theme, the renderer, and
// the site-wide values resolved once at the start of the build. Only the page
// itself changes between calls.
//
// warnings accumulates what a page render found worth reporting but not worth
// failing over. Collecting it here rather than returning it per page keeps
// writePage's signature about whether the write succeeded.
type pageWriter struct {
	outPath  string
	thm      *theme.Theme
	renderer *render.Renderer
	site     config.Site
	nav      theme.NavView
	basePath string
	warnings []string
}

// layoutTemplate maps a front-matter layout name to the template file a theme
// would define for it. The name is free-form: cress promises no vocabulary, so
// which layouts exist is the theme's to decide and document.
func layoutTemplate(layout string) string { return layout + htmlExt }

// resolveLayout picks the template page renders through. A page naming a layout
// its theme does not define falls back to the entry template and is reported,
// which follows nav resolution: a build that is missing something still
// produces a site, and says what was missing.
//
// The name needs no sanitizing. It is looked up in the set of layout file names
// the theme itself declared, so this opens no path, and an absurd layout misses
// like any other unknown name. A partial misses too: it is not a layout, so a
// page cannot reach one by asking for it.
func (w *pageWriter) resolveLayout(page *content.Page) string {
	if page.Layout == "" {
		return ""
	}
	name := layoutTemplate(page.Layout)
	if w.thm.HasLayout(name) {
		return name
	}
	w.warnings = append(w.warnings, missingLayoutWarning(page.SourcePath, page.Layout, w.thm.Name()))
	return ""
}

// missingLayoutWarning phrases an unresolvable layout, naming the page that
// asked, what it asked for, and the theme that had no answer.
func missingLayoutWarning(sourcePath, layout, themeName string) string {
	return fmt.Sprintf("%s names layout %q, which theme %q does not define", sourcePath, layout, themeName)
}

// renderLayout writes data through the theme template named name, or through
// the entry template when name is empty. Every caller that may or may not have
// a template to reach for goes through here, so "fall back to page.html" is
// written once.
func renderLayout(w io.Writer, thm *theme.Theme, name string, data theme.PageData) error {
	if name == "" {
		return thm.Render(w, data)
	}
	return thm.RenderTemplate(w, name, data)
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
	image, imageAlt, imageCaption := w.leadImage(page)

	head, err := renderHead(headData{Description: description, Canonical: absoluteURL})
	if err != nil {
		return fmt.Errorf("%s: %w", page.SourcePath, err)
	}

	data := theme.PageData{
		Site: w.site,
		Nav:  activeNav(w.nav, pageURL),
		Page: theme.PageView{
			Title:        page.Title,
			Description:  description,
			Language:     fallback(page.Language, w.site.Language),
			Image:        image,
			ImageAlt:     imageAlt,
			ImageCaption: imageCaption,
			URL:          pageURL,
			AbsoluteURL:  absoluteURL,
			IsHome:       page.URL == homeURL,
			Meta:         page.Meta,
			HTML:         body,
			Head:         head,
		},
	}

	var buf bytes.Buffer
	if err := renderLayout(&buf, w.thm, w.resolveLayout(page), data); err != nil {
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
	name := ""
	if w.thm.HasLayout(notFoundTemplate) {
		name = notFoundTemplate
	}
	if err := renderLayout(&buf, w.thm, name, data); err != nil {
		return fmt.Errorf("%s: %w", NotFoundFile, err)
	}

	dest := filepath.Join(w.outPath, NotFoundFile)
	if err := os.WriteFile(dest, buf.Bytes(), 0o644); err != nil { // #nosec G306 -- public site files
		return fmt.Errorf("writing %s: %w", dest, err)
	}
	return nil
}
