// Package config loads cress's TOML site configuration (cress.toml): the site
// metadata (title, base URL, theme) and the explicit, flat navigation list.
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

// DefaultAccent is the accent color used when the config leaves it unset. It is
// the green from the default logo.
const DefaultAccent = "#9cb43b"

// navTable is the TOML table holding the navigation (label = content path).
const navTable = "nav"

// ErrNotFound is returned by Load when the config file does not exist. Callers
// distinguish it (with errors.Is) to point the user at `cress init`.
var ErrNotFound = errors.New("config not found")

// Config is a cress site's top-level configuration.
type Config struct {
	Site Site
	// Nav is the site navigation in document order, built from the [nav] table.
	Nav []NavItem
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
	// falls back to the theme's default logo. Typically a file in static/, e.g.
	// "/logo.svg".
	Logo string `toml:"logo"`
	// Favicon is a path or URL to a favicon. Empty falls back to the theme's
	// default favicon. Typically a file in static/, e.g. "/favicon.png".
	Favicon string `toml:"favicon"`
	// Accent is the CSS accent color (a hex value like "#9cb43b") the theme uses
	// as a soft highlight. Empty resolves to DefaultAccent.
	Accent string `toml:"accent"`
}

// NavItem is one entry in the site navigation. Title is the link label; Path
// points at a content file relative to the content directory (e.g. "about.md"),
// which the builder resolves to the page's output URL.
type NavItem struct {
	Title string
	Path  string
}

// document is the on-disk shape decoded from cress.toml. Navigation is a table
// of label -> content path; TOML tables are unordered, so Load recovers the
// authoring order from the parse metadata.
type document struct {
	Site Site              `toml:"site"`
	Nav  map[string]string `toml:"nav"`
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

	cfg := &Config{Site: doc.Site, Nav: orderedNav(md, doc.Nav)}
	cfg.normalize()
	if !isHexColor(cfg.Site.Accent) {
		return nil, fmt.Errorf("config %s: invalid accent color %q (want a hex value like %q)", path, cfg.Site.Accent, DefaultAccent)
	}
	return cfg, nil
}

// orderedNav rebuilds the navigation in the order the entries appear in the
// source, using the parse metadata (a decoded map alone loses that order).
func orderedNav(md toml.MetaData, entries map[string]string) []NavItem {
	if len(entries) == 0 {
		return nil
	}
	nav := make([]NavItem, 0, len(entries))
	seen := make(map[string]bool, len(entries))
	for _, key := range md.Keys() {
		if len(key) != 2 || key[0] != navTable {
			continue
		}
		label := key[1]
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
	if strings.TrimSpace(c.Site.Accent) == "" {
		c.Site.Accent = DefaultAccent
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
