package generator

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/output"
	"github.com/cozybadgerde/cress/internal/render"
)

// Result reports generated pages and nonfatal input diagnostics.
type Result struct {
	Pages    int
	Warnings []string
}

// Build prepares a site's ordinary pages and writes new or empty output.
func Build(ctx context.Context, directory string) (Result, error) {
	var result Result
	site, err := filepath.Abs(directory)
	if err != nil {
		return result, fmt.Errorf("resolve site directory: %w", err)
	}
	site, err = filepath.EvalSymlinks(site)
	if err != nil {
		return result, fmt.Errorf("resolve site directory: %w", err)
	}
	config, warnings, err := content.LoadConfig(ctx, site)
	result.Warnings = warnings
	if err != nil {
		return result, err
	}
	pages, warnings, err := content.Load(ctx, site, config)
	result.Warnings = append(result.Warnings, warnings...)
	if err != nil {
		return result, err
	}

	renderer, err := render.NewRenderer()
	if err != nil {
		return result, err
	}
	var files []output.File
	for _, prepared := range pages {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		rendered, err := renderer.Render(prepared.Data)
		if err != nil {
			return result, fmt.Errorf("%s: %w", prepared.Source, err)
		}
		files = append(files, output.File{
			Path: prepared.Destination, Source: prepared.Source, Data: rendered,
		})
	}
	if err := output.Write(ctx, site, files); err != nil {
		return result, err
	}
	result.Pages = len(files)
	return result, nil
}
