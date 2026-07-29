// Package theme resolves and loads a cress theme: the html/template set that
// wraps rendered page HTML, plus the static assets shipped alongside it. The
// default theme ("cress") is embedded in the binary; sites may override it with
// a directory under themes/.
package theme

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cozybadgerde/cress/internal/config"
)

//go:embed all:builtin
var builtinFS embed.FS

const (
	// entryTemplate is the template every theme must define; it renders one page.
	entryTemplate = "page.html"
	templatesGlob = "templates/*.html"
	staticDir     = "static"
	builtinRoot   = "builtin"
)

// Theme is a loaded theme ready to render pages and expose its assets.
type Theme struct {
	name   string
	tmpl   *template.Template
	static fs.FS // nil when the theme ships no static assets
}

// Name reports the resolved theme name.
func (t *Theme) Name() string { return t.name }

// Resolve loads the theme named name for the site rooted at siteRoot. A theme
// directory at siteRoot/themesDir/name takes precedence; otherwise the built-in
// default theme (config.DefaultTheme) is used. Any other name with no matching
// directory is an error.
func Resolve(siteRoot, themesDir, name string) (*Theme, error) {
	diskPath := filepath.Join(siteRoot, themesDir, name)
	if info, err := os.Stat(diskPath); err == nil && info.IsDir() {
		return load(name, os.DirFS(diskPath))
	}
	if name == config.DefaultTheme {
		sub, err := fs.Sub(builtinFS, filepath.Join(builtinRoot, config.DefaultTheme))
		if err != nil {
			return nil, fmt.Errorf("loading built-in theme: %w", err)
		}
		return load(name, sub)
	}
	return nil, fmt.Errorf("theme %q not found under %s", name, filepath.Join(siteRoot, themesDir))
}

// load parses a theme from fsys, which must contain a templates/ directory (with
// entryTemplate) and may contain a static/ directory.
func load(name string, fsys fs.FS) (*Theme, error) {
	tmpl, err := template.ParseFS(fsys, templatesGlob)
	if err != nil {
		return nil, fmt.Errorf("theme %q: parsing templates: %w", name, err)
	}
	if tmpl.Lookup(entryTemplate) == nil {
		return nil, fmt.Errorf("theme %q: missing %s template", name, entryTemplate)
	}

	t := &Theme{name: name, tmpl: tmpl}
	if info, err := fs.Stat(fsys, staticDir); err == nil && info.IsDir() {
		sub, err := fs.Sub(fsys, staticDir)
		if err != nil {
			return nil, fmt.Errorf("theme %q: reading static: %w", name, err)
		}
		t.static = sub
	}
	return t, nil
}

// Render executes the theme's entry template with data, writing HTML to w.
func (t *Theme) Render(w io.Writer, data any) error {
	return t.RenderTemplate(w, entryTemplate, data)
}

// RenderTemplate executes the theme's template named name with data, writing
// HTML to w. Use it for the templates a theme may define but need not, pairing
// it with HasTemplate to choose a fallback.
func (t *Theme) RenderTemplate(w io.Writer, name string, data any) error {
	if err := t.tmpl.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("theme %q: rendering %s: %w", t.name, name, err)
	}
	return nil
}

// HasTemplate reports whether the theme defines a template named name. Only
// entryTemplate is required, so a caller wanting any other template asks first
// and falls back when the answer is no.
func (t *Theme) HasTemplate(name string) bool { return t.tmpl.Lookup(name) != nil }

// StaticFS returns the theme's static asset tree, or nil when it ships none.
// The builder copies these files verbatim into the output root.
func (t *Theme) StaticFS() fs.FS { return t.static }
