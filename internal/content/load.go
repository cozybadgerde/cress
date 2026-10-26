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

// inventory is the filesystem discovery result; files maps every source to its
// reference target so rendering can resolve links before any page is rendered.
type inventory struct {
	pages    []Page
	assets   []Asset
	files    map[string]reference
	warnings []string
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
	found, err := discover(ctx, root, config)
	prepared.Pages, prepared.Assets, warnings = found.pages, found.assets, found.warnings
	if err != nil {
		return prepared, warnings, err
	}
	for index := range prepared.Pages {
		if err := ctx.Err(); err != nil {
			return prepared, warnings, err
		}
		if err := renderContent(&prepared.Pages[index], config, found.files); err != nil {
			return prepared, warnings, err
		}
	}
	warnings = append(warnings, prepareLists(prepared.Pages, config)...)
	publication, diagnostics, err := preparePublication(prepared.Pages, config)
	warnings = append(warnings, diagnostics...)
	if err != nil {
		return prepared, warnings, err
	}
	prepared.Assets = append(prepared.Assets, publication...)
	return prepared, warnings, nil
}

// discover reads pages and assets from the confined site root. On failure it
// still returns what was found, so warnings stay visible.
func discover(ctx context.Context, root *os.Root, config Config) (inventory, error) {
	found := inventory{files: make(map[string]reference)}
	owners, assetRoots, err := discoverPages(ctx, root, config, &found)
	if err != nil {
		return found, err
	}
	if err := loadPageAssets(ctx, root, config, owners, assetRoots, &found); err != nil {
		return found, err
	}
	if err := loadStatic(ctx, root, &found); err != nil {
		return found, err
	}
	slices.SortFunc(found.assets, func(first, second Asset) int {
		return strings.Compare(first.Source, second.Source)
	})
	for _, asset := range found.assets {
		found.files[asset.Source] = reference{
			url: strings.TrimSuffix(config.Site.URL, "/") + (&url.URL{Path: "/" + asset.Destination}).EscapedPath(),
		}
	}
	return found, nil
}

// discoverPages also returns every page, drafts included, by its asset
// directory so draft assets and metadata are still validated.
func discoverPages(ctx context.Context, root *os.Root, config Config, found *inventory) (map[string]Page, []string, error) {
	owners := make(map[string]Page)
	var assetRoots []string
	err := fs.WalkDir(root.FS(), "content", func(name string, entry fs.DirEntry, walkErr error) error {
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
		found.warnings = append(found.warnings, diagnostics...)
		if err != nil {
			return err
		}
		owners[strings.TrimSuffix(name, ".md")+".assets"] = item
		if !item.Draft || config.IncludeDrafts {
			found.pages = append(found.pages, item)
			found.files[name] = reference{url: item.Data.URL, page: true}
		}
		return nil
	})
	return owners, assetRoots, err
}

func loadPageAssets(ctx context.Context, root *os.Root, config Config, owners map[string]Page, assetRoots []string, found *inventory) error {
	for _, name := range assetRoots {
		owner, exists := owners[name]
		if !exists {
			return fmt.Errorf("%s: orphan asset directory has no owning Markdown page", name)
		}
		assets, err := loadAssets(ctx, root, assetRoot{source: name, destination: path.Dir(owner.Destination)})
		if err != nil {
			return err
		}
		if !owner.Draft || config.IncludeDrafts {
			found.assets = append(found.assets, assets...)
		}
	}
	return nil
}

func loadStatic(ctx context.Context, root *os.Root, found *inventory) error {
	info, err := root.Lstat("static")
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect static: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("static must be a real directory, not a symlink or file")
	}
	assets, err := loadAssets(ctx, root, assetRoot{source: "static", destination: "."})
	if err != nil {
		return err
	}
	found.assets = append(found.assets, assets...)
	return nil
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
