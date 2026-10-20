package content

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"
)

// Asset is an unchanged local file or generated publication file.
type Asset struct {
	Source      string
	Destination string
	Data        []byte
}

// Prepared holds visible pages and their publishable assets.
type Prepared struct {
	Pages  []Page
	Assets []Asset
}

type assetRoot struct {
	source      string
	destination string
}

// Load discovers local sources before resolving and rendering visible content.
func Load(ctx context.Context, site string, config Config) (prepared Prepared, warnings []string, loadErr error) {
	root, err := os.OpenRoot(site)
	if err != nil {
		return prepared, nil, fmt.Errorf("open site directory: %w", err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("close site directory: %w", err))
		}
	}()
	info, err := root.Lstat("content")
	if err != nil {
		return prepared, nil, fmt.Errorf("inspect content directory: %w", err)
	}
	if !info.IsDir() {
		return prepared, nil, fmt.Errorf("content must be a real directory, not a symlink or file")
	}
	owners := make(map[string]Page)
	var assetRoots []string
	files := make(map[string]reference)
	loadErr = fs.WalkDir(root.FS(), "content", func(name string, entry fs.DirEntry, walkErr error) error {
		if err := checkEntry(ctx, name, entry, walkErr); err != nil {
			return err
		}
		if entry.IsDir() {
			if strings.HasSuffix(name, ".assets") {
				assetRoots = append(assetRoots, name)
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".md") {
			return nil
		}
		source, err := root.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		item, diagnostics, err := prepareMetadata(strings.TrimPrefix(name, "content/"), source, config)
		warnings = append(warnings, diagnostics...)
		if err != nil {
			return err
		}
		owners[strings.TrimSuffix(name, ".md")+".assets"] = item
		if !item.Draft {
			prepared.Pages = append(prepared.Pages, item)
			files[name] = reference{url: item.Data.URL, page: true}
		}
		return nil
	})
	if loadErr != nil {
		return prepared, warnings, loadErr
	}
	for _, name := range assetRoots {
		owner, exists := owners[name]
		if !exists {
			return prepared, warnings, fmt.Errorf("%s: orphan asset directory has no owning Markdown page", name)
		}
		assets, err := loadAssets(ctx, root, assetRoot{source: name, destination: path.Dir(owner.Destination)})
		if err != nil {
			return prepared, warnings, err
		}
		if !owner.Draft {
			prepared.Assets = append(prepared.Assets, assets...)
		}
	}
	info, err = root.Lstat("static")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return prepared, warnings, fmt.Errorf("inspect static: %w", err)
	}
	if err == nil {
		if !info.IsDir() {
			return prepared, warnings, fmt.Errorf("static must be a real directory, not a symlink or file")
		}
		assets, err := loadAssets(ctx, root, assetRoot{source: "static", destination: "."})
		if err != nil {
			return prepared, warnings, err
		}
		prepared.Assets = append(prepared.Assets, assets...)
	}
	slices.SortFunc(prepared.Assets, func(first, second Asset) int {
		return strings.Compare(first.Source, second.Source)
	})
	for _, asset := range prepared.Assets {
		files[asset.Source] = reference{
			url: strings.TrimSuffix(config.Site.URL, "/") + (&url.URL{Path: "/" + asset.Destination}).EscapedPath(),
		}
	}
	for index := range prepared.Pages {
		if err := ctx.Err(); err != nil {
			return prepared, warnings, err
		}
		if err := renderContent(&prepared.Pages[index], config, files); err != nil {
			return prepared, warnings, err
		}
	}
	warnings = append(warnings, prepareLists(prepared.Pages)...)
	publication, diagnostics, err := preparePublication(prepared.Pages, config)
	warnings = append(warnings, diagnostics...)
	if err != nil {
		return prepared, warnings, err
	}
	prepared.Assets = append(prepared.Assets, publication...)
	return prepared, warnings, nil
}

func checkEntry(ctx context.Context, name string, entry fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return fmt.Errorf("read %s: %w", name, walkErr)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if entry.Type()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s: source symlinks are not supported", name)
	}
	if !entry.IsDir() && !entry.Type().IsRegular() {
		return fmt.Errorf("%s: source must be a regular file", name)
	}
	return nil
}

func loadAssets(ctx context.Context, root *os.Root, assetsRoot assetRoot) ([]Asset, error) {
	var assets []Asset
	err := fs.WalkDir(root.FS(), assetsRoot.source, func(name string, entry fs.DirEntry, walkErr error) error {
		if err := checkEntry(ctx, name, entry, walkErr); err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := root.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		assets = append(assets, Asset{
			Source: name, Destination: path.Join(assetsRoot.destination, strings.TrimPrefix(name, assetsRoot.source+"/")), Data: data,
		})
		return nil
	})
	return assets, err
}
