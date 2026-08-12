// Package config loads cress's TOML site configuration (cress.toml): the site
// metadata (title, base URL, theme) and the explicit navigation, grouped into a
// primary and a secondary menu.
//
// It also names the conventions a site is built from, which are facts about a
// site's shape rather than values read from the file: the config file name, the
// directories a site is made of, and the defaults an unset key falls back to.
// They live here because this is the package every other one can depend on
// without acquiring anything else.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// FileName is the conventional config file name at the root of a cress site.
const FileName = "cress.toml"

// Conventional site directory names, relative to the site root.
//
// They describe the shape of a cress site rather than the behaviour of any one
// command, which is why they live beside FileName rather than with the builder.
// Three packages need the same vocabulary: build reads content, themes and
// static and writes output; clean refuses to empty the first three; serve
// watches them. Owning them here keeps those packages from depending on each
// other to learn what a directory is called.
const (
	ContentDir = "content"
	StaticDir  = "static"
	ThemesDir  = "themes"
	OutputDir  = "public"
)

// DefaultTheme is the theme used when the config leaves it unset. It resolves
// to the embedded default theme shipped in the binary.
const DefaultTheme = "cress"

// DefaultLanguage is the language tag used when the config leaves it unset. It
// matches what cress emitted before the key existed, so an existing site keeps
// declaring the language it already declared.
const DefaultLanguage = "en"

// exampleAccent is the color named in the error message for an invalid accent.
// It is an example, not a default: an unset accent stays unset so the theme's
// own accent applies.
const exampleAccent = "#4f7a4a"

// exampleBaseURL is the base_url named in the error message for one that is not
// an absolute URL. Like exampleAccent it is an example, not a default: an unset
// base_url stays unset.
const exampleBaseURL = "https://example.com"

// The TOML tables holding the navigation. Each group under [nav] is a table of
// label = content path; the group names are fixed, so a mistyped one is caught
// by the unknown-key check rather than silently rendering nothing.
const (
	navTable  = "nav"
	navMain   = "main"
	navFooter = "footer"
)

// ErrNotFound is returned by Load when the config file does not exist. Callers
// distinguish it (with errors.Is) to point the user at `cress init`.
var ErrNotFound = errors.New("config not found")

// Config is a cress site's top-level configuration.
type Config struct {
	Site Site
	// Nav is the site navigation, built from the [nav] tables.
	Nav Nav
	// Markdown is how the site's Markdown is rendered, from the [markdown] table.
	Markdown Markdown
}

// Markdown holds the rendering options that apply to every page's Markdown. It
// is its own table rather than a few more keys under [site], because [site]
// holds the site's identity and branding and how Markdown is turned into HTML
// is neither.
type Markdown struct {
	// Highlight tokenizes fenced code blocks, so a theme's stylesheet has spans
	// to color. It defaults to off: turning it on rewrites the HTML of every code
	// block on the site, and cress supplies the tokens either way rather than any
	// of the colors.
	Highlight bool `toml:"highlight"`
}

// Nav is the site navigation, split into the menus a theme renders separately.
// Both groups are optional and independent; each keeps the order its entries
// were written in.
type Nav struct {
	// Main is the primary menu, from [nav.main]. Themes render it in the header.
	Main []NavItem
	// Footer is the secondary menu, from [nav.footer], for the links that belong
	// out of the way: legal pages, social links, and the like.
	Footer []NavItem
}

// Site holds the metadata shared by every rendered page.
type Site struct {
	// Title is the site name, shown in the header and page titles.
	Title string `toml:"title"`
	// Description is an optional site tagline, exposed to templates (e.g. for a
	// meta description tag).
	Description string `toml:"description"`
	// BaseURL is the site's canonical root (e.g. "https://example.com", or
	// "https://user.github.io/project" for a project page). It is exposed to
	// templates for absolute links, and its path component decides where in-site
	// links are rooted. Empty is a normal setting rather than a missing one: the
	// links cress emits are root-relative, so a site with no BaseURL serves
	// correctly from any domain that points at it. Only a subdirectory needs this
	// key. Load rejects a value that is not an absolute URL, since a host mistaken
	// for a path ("example.com/site") would quietly misroot every link.
	BaseURL string `toml:"base_url"`
	// BasePath is the path component of BaseURL, cleaned to either "" (the site
	// sits at a domain root) or a rooted path with no trailing slash ("/project").
	// It is derived rather than configured, so cress.toml cannot set it and the
	// two cannot disagree. Every in-site URL cress emits already carries it;
	// templates need it only for a theme's own asset links.
	BasePath string `toml:"-"`
	// Theme names the theme to render with. Empty resolves to DefaultTheme.
	Theme string `toml:"theme"`
	// Language is the site's language tag (e.g. "en", "de", "en-GB"), rendered as
	// the document's lang attribute and overridable per page. Empty resolves to
	// DefaultLanguage: every document declares a language, since a wrong or
	// missing one misleads screen readers and translation tools alike. The value
	// is passed through as written; cress does not validate it against the tag
	// registry, so a typo reaches the output.
	Language string `toml:"language"`
	// Logo is a path or URL to a logo image, rendered in the navigation. Empty
	// means no logo is rendered, unless the theme supplies its own fallback.
	// Typically a file in static/, e.g. "/logo.svg".
	Logo string `toml:"logo"`
	// LogoDark replaces Logo when the reader prefers a dark color scheme. Empty
	// means Logo applies to both. A logo is site identity rather than theme
	// identity, so no theme default can stand in for it: artwork legible on one
	// background needs this to be legible on the other. It is meaningless
	// without Logo, which Load rejects rather than silently dropping.
	LogoDark string `toml:"logo_dark"`
	// Favicon is a path or URL to a favicon. Empty means no icon link is
	// rendered, unless the theme supplies its own fallback. Typically a file in
	// static/, e.g. "/favicon.png".
	Favicon string `toml:"favicon"`
	// Image is a path or URL to the site's lead image, used for a page that names
	// none of its own. It sits between Logo and Favicon, which are chrome, and
	// the images an author writes into a page body, which are content: a theme
	// decides whether to render it, where, and cropped how. Empty means pages
	// have no lead image unless they name one. Typically a file in static/, e.g.
	// "/social.webp".
	Image string `toml:"image"`
	// ImageAlt describes Image for a reader who cannot see it. Empty makes the
	// image decorative, which is a real choice and the wrong one for artwork
	// carrying meaning.
	ImageAlt string `toml:"image_alt"`
	// ImageCaption is the line rendered under Image, for the text an image has to
	// show everyone: an attribution its licence demands, or a disclosure the law
	// does. Empty renders no caption.
	//
	// It travels with Image rather than resolving on its own. A page that names
	// its own image never inherits this, because a credit attached to a different
	// picture is a false claim rather than a missing one.
	ImageCaption string `toml:"image_caption"`
	// Accent is the CSS accent color (a hex value like "#4f7a4a") the theme uses
	// as a highlight. Empty means the theme's own accent applies, including any
	// per-scheme variant it defines; cress supplies no default of its own.
	Accent string `toml:"accent"`
	// AccentDark replaces Accent when the reader prefers a dark color scheme.
	// Empty means Accent applies to both. No single color clears WCAG AA contrast
	// against a light and a dark background at once, so an accent legible in one
	// scheme needs this to be legible in the other.
	AccentDark string `toml:"accent_dark"`
	// Footer is the site's own footer message, replacing the theme's default
	// credit. It is plain text: templates escape it, so markup in it renders as
	// the characters it is made of rather than as HTML.
	Footer string `toml:"footer"`
	// Copyright is an optional copyright notice for the footer. Empty omits it.
	// The builder expands a "{year}" token to the year the site was built, so an
	// author need not hardcode a year that then goes stale.
	Copyright string `toml:"copyright"`
}

// NavItem is one entry in the site navigation. Title is the link label; Path
// points at a content file relative to the content directory (e.g. "about.md"),
// which the builder resolves to the page's output URL.
type NavItem struct {
	Title string
	Path  string
}

// document is the on-disk shape decoded from cress.toml.
type document struct {
	Site     Site        `toml:"site"`
	Nav      navDocument `toml:"nav"`
	Markdown Markdown    `toml:"markdown"`
}

// navDocument is the on-disk [nav] table: one sub-table per menu, each mapping
// label -> content path. TOML tables are unordered, so Load recovers the
// authoring order from the parse metadata.
type navDocument struct {
	Main   map[string]string `toml:"main"`
	Footer map[string]string `toml:"footer"`
}

// Load reads and parses the config at path. A missing file yields ErrNotFound
// (wrapped); unknown keys are rejected as likely typos rather than silently
// dropped.
func Load(path string) (*Config, error) {
	// #nosec G304 -- path is the site's own config, chosen by the user.
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s: %w", path, ErrNotFound)
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var doc document
	md, err := toml.Decode(string(raw), &doc)
	if err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, k := range undecoded {
			keys[i] = k.String()
		}
		return nil, fmt.Errorf("config %s: unknown key(s): %s", path, strings.Join(keys, ", "))
	}

	cfg := &Config{
		Site: doc.Site,
		Nav: Nav{
			Main:   orderedNav(md, navMain, doc.Nav.Main),
			Footer: orderedNav(md, navFooter, doc.Nav.Footer),
		},
		Markdown: doc.Markdown,
	}
	cfg.normalize()
	if err := cfg.resolveBaseURL(path); err != nil {
		return nil, err
	}
	if err := validateAccents(path, cfg.Site); err != nil {
		return nil, err
	}
	if err := validateLogos(path, cfg.Site); err != nil {
		return nil, err
	}
	if err := validateTheme(path, cfg.Site); err != nil {
		return nil, err
	}
	return cfg, nil
}

// resolveBaseURL owns the whole base_url story: it decides whether the key was
// set at all, validates it, canonicalizes BaseURL, and derives BasePath from the
// path component. Keeping the trailing-slash trim here rather than in normalize
// is what lets an error quote the author's own text; trimming first meant
// reporting a value nobody wrote.
//
// A blank value, or one that is nothing but slashes, counts as unset: naming no
// site root and naming the root of a domain are the same statement, and the
// theme key already treats blank as unset. A value that does name something must
// be an absolute URL, because the alternative is worse than an error. Parsing
// "example.com/site" leniently yields a path of "example.com/site", so every
// link in the site would be emitted under a directory that does not exist, and
// the only symptom would be a site that 404s everywhere once published.
func (c *Config) resolveBaseURL(path string) error {
	raw := strings.TrimSpace(c.Site.BaseURL)
	if strings.Trim(raw, "/") == "" {
		c.Site.BaseURL = ""
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("config %s: invalid base_url %q: %w", path, raw, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("config %s: base_url %q is not an absolute URL (want a scheme and host, like %q)", path, raw, exampleBaseURL)
	}
	c.Site.BaseURL = strings.TrimRight(raw, "/")
	c.Site.BasePath = cleanBasePath(u.Path)
	return nil
}

// cleanBasePath reduces a URL path to the form the builder prefixes with: empty
// for a site at a domain root, else rooted with no trailing slash. Both ends
// matter, since every URL cress emits already starts with a slash and nothing
// should end up doubling it.
func cleanBasePath(p string) string {
	p = strings.Trim(p, "/")
	if p == "" {
		return ""
	}
	return "/" + p
}

// validateLogos rejects a dark logo with no light one to pair it with. The two
// are a pair by construction: the dark variant is offered as an alternative to
// Logo for one color scheme, so on its own it names a scheme to swap in for and
// nothing to swap out. Reporting it beats rendering no logo at all and leaving
// the author to guess which of the two keys was the problem.
func validateLogos(path string, site Site) error {
	if site.LogoDark != "" && site.Logo == "" {
		return fmt.Errorf("config %s: logo_dark is set without logo; a dark logo replaces the light one and cannot stand alone", path)
	}
	return nil
}

// validateAccents rejects a set accent that is not a hex color. The values are
// interpolated into the theme's stylesheet, so restricting their shape keeps
// arbitrary text out of the rendered CSS. An unset one is left unset rather
// than defaulted: the theme is the only place that knows its own backgrounds,
// and therefore the only place that can pick a legible accent for each scheme.
// validateTheme rejects a theme name that is anything but a single directory
// name. The value is joined onto themes/ to find the theme on disk, and
// filepath.Join resolves ".." rather than refusing it, so "../../elsewhere"
// would select a directory outside the site: its templates would render every
// page, and its static/ would be copied into the published output. A theme name
// is the one config value that chooses which code runs, so it is worth checking
// as closely as the accent that reaches a stylesheet.
func validateTheme(path string, site Site) error {
	name := site.Theme
	if name == "" {
		return nil
	}
	if !ValidThemeName(name) {
		return fmt.Errorf("config %s: invalid theme %q (want a directory name under %s/, not a path)", path, name, ThemesDir)
	}
	return nil
}

// ValidThemeName reports whether name is usable as a theme: a single directory
// name and nothing else. Scaffolding a theme joins the name onto themes/ just as
// resolving one does, so both ask here rather than keeping a rule each, which
// would leave a name cress writes and a name cress refuses to load.
func ValidThemeName(name string) bool {
	return name != "" && name == filepath.Base(name) && name != "." && name != ".." &&
		!strings.ContainsRune(name, '/') && !strings.ContainsRune(name, filepath.Separator)
}

func validateAccents(path string, site Site) error {
	for _, accent := range []struct{ key, value string }{
		{"accent", site.Accent},
		{"accent_dark", site.AccentDark},
	} {
		if accent.value != "" && !isHexColor(accent.value) {
			return fmt.Errorf("config %s: invalid %s color %q (want a hex value like %q)", path, accent.key, accent.value, exampleAccent)
		}
	}
	return nil
}

// orderedNav rebuilds one navigation group in the order its entries appear in
// the source, using the parse metadata (a decoded map alone loses that order).
func orderedNav(md toml.MetaData, group string, entries map[string]string) []NavItem {
	if len(entries) == 0 {
		return nil
	}
	nav := make([]NavItem, 0, len(entries))
	seen := make(map[string]bool, len(entries))
	for _, key := range md.Keys() {
		if len(key) != 3 || key[0] != navTable || key[1] != group {
			continue
		}
		label := key[2]
		if seen[label] {
			continue
		}
		seen[label] = true
		nav = append(nav, NavItem{Title: label, Path: entries[label]})
	}
	return nav
}

// normalize fills in defaults for fields left unset. base_url is not among them:
// resolveBaseURL parses it, and normalizing it here first would corrupt the text
// an error message needs to quote.
func (c *Config) normalize() {
	if strings.TrimSpace(c.Site.Theme) == "" {
		c.Site.Theme = DefaultTheme
	}
	c.Site.Language = strings.TrimSpace(c.Site.Language)
	if c.Site.Language == "" {
		c.Site.Language = DefaultLanguage
	}
}

// isHexColor reports whether s is a CSS hex color: #RGB, #RRGGBB, or #RRGGBBAA.
// The accent is interpolated into a stylesheet, so restricting it to this shape
// keeps arbitrary text out of the rendered CSS.
func isHexColor(s string) bool {
	if len(s) == 0 || s[0] != '#' {
		return false
	}
	digits := s[1:]
	switch len(digits) {
	case 3, 6, 8:
	default:
		return false
	}
	for _, r := range digits {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}
