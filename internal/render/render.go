package render

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/cozybadgerde/cress/internal/page"
)

//go:embed skeleton/*.gohtml
var skeleton embed.FS

// Renderer composes prepared pages with the embedded development skeleton.
type Renderer struct {
	template *template.Template
}

// NewRenderer loads the embedded development templates.
func NewRenderer() (*Renderer, error) {
	parsed, err := template.ParseFS(skeleton, "skeleton/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("parse skeleton template: %w", err)
	}
	return &Renderer{template: parsed}, nil
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
	var output bytes.Buffer
	if err := renderer.template.ExecuteTemplate(&output, name, prepared); err != nil {
		return nil, fmt.Errorf("render skeleton template: %w", err)
	}
	return output.Bytes(), nil
}
