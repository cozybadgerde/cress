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

// NewRenderer loads the embedded ordinary-page template.
func NewRenderer() (*Renderer, error) {
	parsed, err := template.ParseFS(skeleton, "skeleton/page.gohtml")
	if err != nil {
		return nil, fmt.Errorf("parse skeleton template: %w", err)
	}
	return &Renderer{template: parsed}, nil
}

// Render produces a complete page without discovering or modifying content.
func (renderer *Renderer) Render(prepared page.Page) ([]byte, error) {
	var output bytes.Buffer
	if err := renderer.template.ExecuteTemplate(&output, "page.gohtml", prepared); err != nil {
		return nil, fmt.Errorf("render skeleton template: %w", err)
	}
	return output.Bytes(), nil
}
