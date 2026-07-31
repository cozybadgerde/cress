package build

import (
	"bytes"
	"fmt"
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
