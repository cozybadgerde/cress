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
	"path"
	"path/filepath"
	"strings"

	"github.com/cozybadgerde/cress/internal/config"
)

//go:embed all:builtin
var builtinFS embed.FS

const (
	// entryTemplate is the template every theme must define; it renders one page.
	entryTemplate = "page.html"
	// layoutsGlob matches the templates a page can be rendered through. It is
	// one level deep on purpose: a layout is addressed by its file name, and
	// html/template keys a parsed file by its base name, so allowing nested
	// layouts would put two files sharing a base name into one namespace and let
	// whichever parsed last win without saying so.
	layoutsGlob = "templates/*.html"
	// partialsDir holds the fragments layouts compose from. A partial is reached
	// by its {{ define }} name rather than by its path, which is what leaves the
	// tree below this free: a theme may group its partials into subdirectories
	// however it likes without any of it meaning anything to the loader.
	partialsDir = "templates/partials"
	staticDir   = "static"
	builtinRoot = "builtin"
	htmlExt     = ".html"
)

// Theme is a loaded theme ready to render pages and expose its assets.
type Theme struct {
	name string
	tmpl *template.Template
	// layouts names the templates a page may be rendered through. Holding it
	// apart from tmpl is what separates the two halves of a theme: everything
	// parses into one set so a layout can call any partial, while only the names
	// in here can be asked for by a page.
	layouts map[string]bool
	static  fs.FS // nil when the theme ships no static assets
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
// entryTemplate), may contain templates/partials/, and may contain a static/
// directory.
func load(name string, fsys fs.FS) (*Theme, error) {
	tmpl, err := template.ParseFS(fsys, layoutsGlob)
	if err != nil {
		return nil, fmt.Errorf("theme %q: parsing templates: %w", name, err)
	}
	layouts, err := layoutNames(fsys)
	if err != nil {
		return nil, fmt.Errorf("theme %q: %w", name, err)
	}
	if !layouts[entryTemplate] {
		return nil, fmt.Errorf("theme %q: missing %s template", name, entryTemplate)
	}
	if err := parsePartials(tmpl, fsys); err != nil {
		return nil, fmt.Errorf("theme %q: %w", name, err)
	}

	t := &Theme{name: name, tmpl: tmpl, layouts: layouts}
	if info, err := fs.Stat(fsys, staticDir); err == nil && info.IsDir() {
		sub, err := fs.Sub(fsys, staticDir)
		if err != nil {
			return nil, fmt.Errorf("theme %q: reading static: %w", name, err)
		}
		t.static = sub
	}
	return t, nil
}

// layoutNames lists the templates a page may be rendered through, keyed by the
// name a page asks for.
//
// It reads the file names rather than the parsed set, and the difference is the
// point: a {{ define }} inside a layout is a fragment that layout composes with,
// not a layout a page may name. Deriving this from the parse would put both in
// one namespace, which is the arrangement this directory layout exists to undo.
func layoutNames(fsys fs.FS) (map[string]bool, error) {
	files, err := fs.Glob(fsys, layoutsGlob)
	if err != nil {
		return nil, fmt.Errorf("listing layouts: %w", err)
	}
	layouts := make(map[string]bool, len(files))
	for _, file := range files {
		layouts[path.Base(file)] = true
	}
	return layouts, nil
}

// parsePartials adds every template under partialsDir to tmpl.
//
// The tree is walked rather than globbed because a partial's path means nothing:
// it is reached by its {{ define }} name, so a theme may nest its partials as
// deeply as it likes and the loader neither knows nor cares.
//
// The directory is optional, and template.ParseFS treats a pattern matching no
// files as an error, so the walk is entered only once the directory is known to
// be there. A theme with no partials loads exactly as it did before it could
// have had any.
func parsePartials(tmpl *template.Template, fsys fs.FS) error {
	info, err := fs.Stat(fsys, partialsDir)
	if err != nil || !info.IsDir() {
		return nil
	}
	return fs.WalkDir(fsys, partialsDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.EqualFold(path.Ext(p), htmlExt) {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return fmt.Errorf("reading %s: %w", p, err)
		}
		// Registering the file under its full path keeps it clear of the layout
		// namespace, which holds base names only. What matters is the {{ define }}
		// blocks inside it: those land in the shared set under the names the theme
		// chose, which is how a layout reaches them.
		if _, err := tmpl.New(p).Parse(string(data)); err != nil {
			return fmt.Errorf("parsing %s: %w", p, err)
		}
		return nil
	})
}

// Render executes the theme's entry template with data, writing HTML to w.
func (t *Theme) Render(w io.Writer, data PageData) error {
	return t.RenderTemplate(w, entryTemplate, data)
}

// RenderTemplate executes the theme's template named name with data, writing
// HTML to w. Use it for the layouts a theme may define but need not, pairing it
// with HasLayout to choose a fallback.
//
// Both render entry points take PageData rather than any: every template in the
// contract receives the same shape, and naming it here is what keeps a caller
// from inventing a second one.
func (t *Theme) RenderTemplate(w io.Writer, name string, data PageData) error {
	if err := t.tmpl.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("theme %q: rendering %s: %w", t.name, name, err)
	}
	return nil
}

// HasLayout reports whether the theme defines a layout named name. Only
// entryTemplate is required, so a caller wanting any other layout asks first and
// falls back when the answer is no.
//
// A partial is deliberately not a layout, however well it might render on its
// own. It exists to be composed by a template that knows where it belongs, so a
// page naming one is told no, exactly like a page naming something the theme
// does not have at all.
func (t *Theme) HasLayout(name string) bool { return t.layouts[name] }

// StaticFS returns the theme's static asset tree, or nil when it ships none.
// The builder copies these files verbatim into the output root.
func (t *Theme) StaticFS() fs.FS { return t.static }
