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
	Prefix   string
	Token    string
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
	return generate(ctx, buildRequest{site: site, version: version, writer: writer})
}

// PreviewRequest fixes the target and draft visibility for one session.
type PreviewRequest struct {
	Version string
	Drafts  bool
	Session *output.Session
}

// BuildPreview uses the normal pipeline without touching deployable output.
func BuildPreview(ctx context.Context, request PreviewRequest) (result Result, buildErr error) {
	token, err := request.Session.Attempt(ctx, func(writer *output.Writer) error {
		var err error
		result, err = generate(ctx, buildRequest{
			site: request.Session.Site(), version: request.Version, drafts: request.Drafts, writer: writer,
		})
		return err
	})
	result.Token = token
	return result, err
}

type buildRequest struct {
	site    string
	version string
	drafts  bool
	writer  *output.Writer
}

func generate(ctx context.Context, request buildRequest) (result Result, buildErr error) {
	site := request.site
	config, warnings, err := content.LoadConfig(ctx, site)
	result.Warnings = warnings
	if err != nil {
		return result, err
	}
	config.IncludeDrafts = request.drafts
	result.Prefix = config.Site.URL
	prepared, warnings, err := content.Load(ctx, site, config)
	result.Warnings = append(result.Warnings, warnings...)
	if err != nil {
		return result, err
	}

	pkg, diagnostics, err := theme.Load(ctx, theme.Request{
		Site: site, Name: config.Theme, Language: config.Site.Language, Prefix: config.Site.URL, Version: request.version,
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
	notFound, err := renderer.RenderNotFound(config.Site)
	if err != nil {
		return result, fmt.Errorf("render not-found page: %w", err)
	}
	files = append(files, output.File{
		Path: "404.html", Source: "generated 404.html", Data: notFound,
	})
	for _, asset := range prepared.Assets {
		files = append(files, output.File{Path: asset.Destination, Source: asset.Source, Data: asset.Data})
	}
	for _, asset := range pkg.Assets {
		files = append(files, output.File{Path: asset.Path, Source: asset.Source, Data: asset.Data})
	}
	if err := request.writer.Write(ctx, files); err != nil {
		return result, err
	}
	result.Pages = len(prepared.Pages) + 1
	return result, nil
}
