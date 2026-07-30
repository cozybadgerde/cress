package theme

import (
	"html/template"

	"github.com/cozybadgerde/cress/internal/config"
)

// PageData is the data a theme's template receives for one page.
//
// These field names are cress's theme API. Every template ever written against
// cress reaches for .Site, .Nav and .Page by exactly these names, so renaming
// or removing one silently breaks every third-party theme: treat it as a
// breaking change, the same as a CLI flag. The builder assembles the value;
// this package defines it, because it already owns the template contract.
type PageData struct {
	// Site is the [site] table of cress.toml, with the {year} token in
	// Copyright already expanded. Config owns those field names, but a template
	// reads them (.Site.Title, .Site.Logo, ...) so they are part of this
	// contract too. Every one of them may be empty except where config
	// documents otherwise, so a template guards each.
	Site config.Site
	// Nav is the resolved navigation, one slice per configured menu.
	Nav NavView
	// Page is the page being rendered.
	Page PageView
}

// NavView is the resolved navigation as a template sees it: one slice per menu,
// so a theme can place the primary and secondary links independently.
type NavView struct {
	// Main is the [nav.main] group in authoring order. Empty when the site
	// configures no main menu, so {{ with .Nav.Main }} skips the whole block.
	Main []NavLink
	// Footer is the [nav.footer] group, on the same terms as Main.
	Footer []NavLink
}

// NavLink is one resolved navigation entry.
type NavLink struct {
	// Title is the link label: the [nav] entry's key, or the target page's
	// title when that key is empty. Plain text, escaped by the template.
	Title string
	// URL is the target's root-relative URL ("/about.html", or "/" for the home
	// page). Never empty: an entry whose target does not exist is dropped with
	// a build warning rather than rendered as a dead link.
	URL string
	// Active reports whether this entry points at the page being rendered. It
	// is set per page and in every group, so a footer link to the current page
	// is active too.
	Active bool
}

// PageView is a page as a template sees it: its metadata plus rendered HTML.
type PageView struct {
	// Title is the page title: front-matter "title", else the first H1, else
	// the file's base name. Plain text, escaped by the template.
	Title string
	// URL is the page's root-relative URL ("/guide/setup.html", or "/guide/"
	// for an index file).
	URL string
	// Meta is the page's front matter, passed through verbatim so a theme can
	// read keys cress itself does not define. Empty when the page has none.
	Meta map[string]any
	// HTML is the rendered Markdown body. It is markup rather than text, and
	// typed template.HTML to say so: {{ .Page.HTML }} emits it as-is. Escaping
	// it again would show readers their own page source.
	HTML template.HTML
}
