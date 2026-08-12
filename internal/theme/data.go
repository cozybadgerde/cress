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
	// Copyright already expanded and every in-site URL in it (Logo, LogoDark,
	// Favicon) rooted under Site.BasePath. Config owns those field names, but a
	// template reads them (.Site.Title, .Site.Logo, ...) so they are part of this
	// contract too. Every one of them may be empty except where config
	// documents otherwise, so a template guards each.
	//
	// A theme must prefix its own asset links with .Site.BasePath, because those
	// are the one set of URLs cress does not emit:
	//
	//	<link rel="stylesheet" href="{{ .Site.BasePath }}/style.css" />
	//
	// It is empty for a site at a domain root, which is why the form above is
	// also correct there. A theme that hardcodes "/style.css" instead renders
	// correctly at a root and silently unstyled under a subdirectory, with
	// nothing to warn the site's author which of the two is at fault.
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
	// page), already rooted under Site.BasePath. Never empty: an entry whose
	// target does not exist is dropped with a build warning rather than rendered
	// as a dead link.
	URL string
	// Active reports whether this entry points at the page being rendered. It
	// is set per page and in every group, so a footer link to the current page
	// is active too.
	Active bool
}

// PageView is a page as a template sees it: its metadata plus rendered HTML.
//
// Several values here have a namesake on Site, and they are not the same thing:
// Site holds what the author configured, PageView holds what this page resolved
// to. A template renders the PageView one and reads the Site one only when it
// deliberately wants the site-wide value.
type PageView struct {
	// Title is the page title: front-matter "title", else the first H1, else
	// the file's base name. Plain text, escaped by the template.
	Title string
	// Description is the page's meta description: front-matter "description",
	// else Site.Description. Empty when neither is set, so a template guards it.
	// Already included in Head; read it directly only to arrange it differently.
	Description string
	// Language is the page's language tag: front-matter "language", else
	// Site.Language. Never empty, and belongs in the document's lang attribute:
	//
	//	<html lang="{{ .Page.Language }}">
	//
	// Head cannot carry it, since it is an attribute rather than an element.
	Language string
	// Image is the page's lead image: front-matter "image", else Site.Image,
	// already rooted under Site.BasePath. Empty when neither is set, so a
	// template guards it:
	//
	//	{{ with .Page.Image }}
	//	<figure><img src="{{ . }}" alt="{{ $.Page.ImageAlt }}" />
	//	{{ with $.Page.ImageCaption }}<figcaption>{{ . }}</figcaption>{{ end }}
	//	</figure>
	//	{{ end }}
	//
	// It is the theme's to place, size and crop, or to ignore.
	Image string
	// ImageAlt describes Image for a reader who cannot see it, and is empty when
	// the author left the image decorative. Emit it either way: an empty alt
	// attribute is what marks an image as decoration, while no alt attribute at
	// all leaves a screen reader reading out the file name.
	ImageAlt string
	// ImageCaption is the line to render under Image, and is empty when there is
	// none. It carries the text an image has to show everyone rather than only
	// the people who cannot see it: an attribution its licence demands, or a
	// disclosure the law does. A theme that renders Image should render this too,
	// or a site that depends on it has no way to comply.
	//
	// Both belong to whichever image Image resolved to. A page naming its own
	// image never inherits the site's caption, so what a template renders is
	// always about the picture beside it.
	ImageCaption string
	// AbsoluteURL is the page's full URL, Site.BaseURL joined with URL. Empty
	// when the site sets no base_url, because there is no host to build it from;
	// a template that renders it must guard it.
	AbsoluteURL string
	// IsHome marks the page served at the site root, the one rendered from
	// content/index.md. False for every other page, and for all of them when a
	// site has no index of its own. Themes use it for the things that read
	// differently on a front page, such as dropping the site name from a title
	// that already is the site name.
	IsHome bool
	// Head is the metadata block cress renders for this page: the meta
	// description and the canonical link, each omitted when it has no value.
	// Emit it inside <head> and let cress own what goes in it:
	//
	//	<head>
	//	  <meta charset="utf-8" />
	//	  {{ .Page.Head }}
	//	  <link rel="stylesheet" href="{{ .Site.BasePath }}/style.css" />
	//	</head>
	//
	// It holds only what describes the page. The document's own furniture, the
	// title, charset, viewport, icon and stylesheets, stays the theme's, so a
	// theme keeps control of its <head> and gains the metadata for one line.
	// A theme that would rather arrange the metadata itself reads Description
	// and AbsoluteURL and omits this; rendering both emits each tag twice.
	//
	// Already-escaped markup, like HTML: do not escape it again.
	Head template.HTML
	// URL is the page's root-relative URL ("/guide/setup.html", or "/guide/"
	// for an index file), already rooted under Site.BasePath.
	URL string
	// Meta is the page's front matter, passed through verbatim so a theme can
	// read keys cress itself does not define. Empty when the page has none.
	Meta map[string]any
	// HTML is the rendered Markdown body. It is markup rather than text, and
	// typed template.HTML to say so: {{ .Page.HTML }} emits it as-is. Escaping
	// it again would show readers their own page source.
	HTML template.HTML
}
