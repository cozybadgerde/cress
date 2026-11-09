package render

import (
	"bytes"
	"fmt"
	"html/template"
	"maps"
	"slices"
	"strings"

	"github.com/cozybadgerde/cress/internal/page"
	"github.com/cozybadgerde/cress/internal/theme"
)

// Renderer composes prepared pages with a complete theme package.
type Renderer struct {
	template      *template.Template
	notFoundTitle string
}

// NewRenderer loads the embedded development templates.
func NewRenderer() (*Renderer, error) {
	pkg, err := theme.Skeleton()
	if err != nil {
		return nil, err
	}
	return NewPackageRenderer(pkg)
}

// NewPackageRenderer parses a complete theme snapshot with presentation-only helpers.
func NewPackageRenderer(pkg theme.Package) (*Renderer, error) {
	parsed := template.New("theme").Option("missingkey=error").Funcs(template.FuncMap{
		"asset": pkg.AssetURL, "label": pkg.Label,
	})
	for _, name := range slices.Sorted(maps.Keys(pkg.Templates)) {
		if _, err := parsed.New(name).Parse(pkg.Templates[name]); err != nil {
			return nil, fmt.Errorf("parse theme %s template %s: %w", pkg.Manifest.Name, name, err)
		}
	}
	for _, name := range []string{"page.gohtml", "landing.gohtml", "list.gohtml", "404.gohtml"} {
		entry := parsed.Lookup(name)
		if entry == nil || entry.Tree == nil || strings.TrimSpace(entry.Tree.Root.String()) == "" {
			return nil, fmt.Errorf("theme %s: required template %s is missing or empty", pkg.Manifest.Name, name)
		}
	}
	notFoundTitle, err := pkg.Label("not_found_title")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(notFoundTitle) == "" {
		return nil, fmt.Errorf("theme %s: label %q must not be empty", pkg.Manifest.Name, "not_found_title")
	}
	notFoundMessage, err := pkg.Label("not_found_message")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(notFoundMessage) == "" {
		return nil, fmt.Errorf("theme %s: label %q must not be empty", pkg.Manifest.Name, "not_found_message")
	}
	return &Renderer{template: parsed, notFoundTitle: notFoundTitle}, nil
}

// Render produces a complete page without discovering or modifying content.
func (renderer *Renderer) Render(prepared page.Page) ([]byte, error) {
	var name string
	switch prepared.Type {
	case page.Ordinary:
		name = "page.gohtml"
	case page.Landing:
		name = "landing.gohtml"
	case page.List:
		name = "list.gohtml"
	default:
		return nil, fmt.Errorf("unsupported prepared page type %q", prepared.Type)
	}
	return renderer.render(name, prepared)
}

// RenderNotFound produces the theme's localized not-found page.
func (renderer *Renderer) RenderNotFound(site page.Site) ([]byte, error) {
	return renderer.render("404.gohtml", page.Page{
		Title: renderer.notFoundTitle, URL: site.URL + "404.html", Site: site,
	})
}

func (renderer *Renderer) render(name string, prepared page.Page) ([]byte, error) {
	var output bytes.Buffer
	if err := renderer.template.ExecuteTemplate(&output, name, prepared); err != nil {
		return nil, fmt.Errorf("render theme template %s: %w", name, err)
	}
	return output.Bytes(), nil
}
