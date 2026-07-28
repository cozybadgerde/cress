// Package config loads cress's TOML site configuration (cress.toml): the site
// metadata (title, base URL, theme) and the explicit navigation, grouped into a
// primary and a secondary menu.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// FileName is the conventional config file name at the root of a cress site.
const FileName = "cress.toml"

// DefaultTheme is the theme used when the config leaves it unset. It resolves
// to the embedded default theme shipped in the binary.
const DefaultTheme = "cress"

// exampleAccent is the color named in the error message for an invalid accent.
// It is an example, not a default: an unset accent stays unset so the theme's
// own accent applies.
const exampleAccent = "#4f7a4a"

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
	// BaseURL is the site's canonical root (e.g. "https://example.com"). It is
	// exposed to templates for absolute links; in-site links stay root-relative.
	BaseURL string `toml:"base_url"`
	// Theme names the theme to render with. Empty resolves to DefaultTheme.
	Theme string `toml:"theme"`
	// Logo is a path or URL to a logo image, rendered in the navigation. Empty
	// means no logo is rendered, unless the theme supplies its own fallback.
	// Typically a file in static/, e.g. "/logo.svg".
	Logo string `toml:"logo"`
	// Favicon is a path or URL to a favicon. Empty means no icon link is
	// rendered, unless the theme supplies its own fallback. Typically a file in
	// static/, e.g. "/favicon.png".
	Favicon string `toml:"favicon"`
	// Accent is the CSS accent color (a hex value like "#4f7a4a") the theme uses
	// as a highlight. Empty means the theme's own accent applies, including any
	// per-scheme variant it defines; cress supplies no default of its own.
	Accent string `toml:"accent"`
	// AccentDark replaces Accent when the reader prefers a dark color scheme.
	// Empty means Accent applies to both. No single color clears WCAG AA contrast
	// against a light and a dark background at once, so an accent legible in one
	// scheme needs this to be legible in the other.
	AccentDark string `toml:"accent_dark"`
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
	Site Site        `toml:"site"`
	Nav  navDocument `toml:"nav"`
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
	}
	cfg.normalize()
	if err := validateAccents(path, cfg.Site); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validateAccents rejects a set accent that is not a hex color. The values are
// interpolated into the theme's stylesheet, so restricting their shape keeps
// arbitrary text out of the rendered CSS. An unset one is left unset rather
// than defaulted: the theme is the only place that knows its own backgrounds,
// and therefore the only place that can pick a legible accent for each scheme.
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

// normalize fills in defaults for fields left unset.
func (c *Config) normalize() {
	if strings.TrimSpace(c.Site.Theme) == "" {
		c.Site.Theme = DefaultTheme
	}
	c.Site.BaseURL = strings.TrimRight(c.Site.BaseURL, "/")
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
