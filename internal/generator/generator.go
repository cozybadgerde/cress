package generator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/output"
	"github.com/cozybadgerde/cress/internal/render"
	"github.com/cozybadgerde/cress/internal/theme"
)

// Result reports generated pages and nonfatal input diagnostics.
type Result struct {
	Pages    int
	Warnings []string
}

// Build prepares and safely replaces a site's complete generated output.
func Build(ctx context.Context, directory string) (Result, error) {
	return BuildVersion(ctx, directory, "dev")
}

// BuildVersion validates named themes against the executable's runtime version.
func BuildVersion(ctx context.Context, directory, version string) (result Result, buildErr error) {
	site, err := filepath.Abs(directory)
	if err != nil {
		return result, fmt.Errorf("resolve site directory: %w", err)
	}
	site, err = filepath.EvalSymlinks(site)
	if err != nil {
		return result, fmt.Errorf("resolve site directory: %w", err)
	}
	writer, err := output.Acquire(ctx, site)
	if err != nil {
		return result, err
	}
	defer func() { buildErr = errors.Join(buildErr, writer.Close()) }()
	config, warnings, err := content.LoadConfig(ctx, site)
	result.Warnings = warnings
	if err != nil {
		return result, err
	}
	prepared, warnings, err := content.Load(ctx, site, config)
	result.Warnings = append(result.Warnings, warnings...)
	if err != nil {
		return result, err
	}

	pkg, diagnostics, err := theme.Load(ctx, theme.Request{
		Site: site, Name: config.Theme, Language: config.Site.Language, Prefix: config.Site.URL, Version: version,
	})
	result.Warnings = append(result.Warnings, diagnostics...)
	if err != nil {
		return result, err
	}
	renderer, err := render.NewPackageRenderer(pkg)
	if err != nil {
		return result, err
	}
	var files []output.File
	for _, item := range prepared.Pages {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		rendered, err := renderer.Render(item.Data)
		if err != nil {
			return result, fmt.Errorf("%s: %w", item.Source, err)
		}
		files = append(files, output.File{
			Path: item.Destination, Source: item.Source, Data: rendered,
		})
	}
	for _, asset := range prepared.Assets {
		files = append(files, output.File{Path: asset.Destination, Source: asset.Source, Data: asset.Data})
	}
	for _, asset := range pkg.Assets {
		files = append(files, output.File{Path: asset.Path, Source: asset.Source, Data: asset.Data})
	}
	if err := writer.Write(ctx, files); err != nil {
		return result, err
	}
	result.Pages = len(prepared.Pages)
	return result, nil
}
