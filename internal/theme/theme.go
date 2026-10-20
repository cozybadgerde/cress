package theme

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/pelletier/go-toml/v2"
)

//go:embed skeleton
var embedded embed.FS

// Request selects a package and prepares its language and deployment URLs.
type Request struct {
	Site     string
	Name     string
	Language string
	Prefix   string
	Version  string
}

// Asset is a designated public theme file.
type Asset struct {
	Source string
	Path   string
	Data   []byte
}

// Package contains snapshot template sources and prepared presentation helpers.
type Package struct {
	Manifest  Manifest
	Templates map[string]string
	Assets    []Asset
	labels    map[string]string
	language  string
	assetURLs map[string]string
}

// Skeleton loads the unconfigured development skeleton, not a named built-in.
func Skeleton() (Package, error) {
	source, err := fs.Sub(embedded, "skeleton")
	if err != nil {
		return Package{}, fmt.Errorf("open embedded skeleton: %w", err)
	}
	pkg, _, err := loadPackage(context.Background(), source, Request{Language: "en", Prefix: "/"})
	pkg.Assets = nil
	return pkg, err
}

// Load resolves a complete local package without merging or invalid-source fallback.
func Load(ctx context.Context, request Request) (pkg Package, warnings []string, loadErr error) {
	if request.Name == "" {
		pkg, err := Skeleton()
		return pkg, nil, err
	}
	if !ValidName(request.Name) {
		return pkg, nil, fmt.Errorf("invalid theme name %q", request.Name)
	}
	root, err := os.OpenRoot(request.Site)
	if err != nil {
		return pkg, nil, fmt.Errorf("open site for theme: %w", err)
	}
	defer func() { loadErr = errors.Join(loadErr, root.Close()) }()
	for _, name := range []string{"themes", path.Join("themes", request.Name)} {
		info, err := root.Lstat(name)
		if errors.Is(err, fs.ErrNotExist) {
			return pkg, nil, fmt.Errorf("theme %q is absent locally and no built-in theme has that name", request.Name)
		}
		if err != nil {
			return pkg, nil, fmt.Errorf("inspect %s: %w", name, err)
		}
		if !info.IsDir() {
			return pkg, nil, fmt.Errorf("%s must be a real directory, not a symlink or file", name)
		}
	}
	source, err := fs.Sub(root.FS(), path.Join("themes", request.Name))
	if err != nil {
		return pkg, nil, fmt.Errorf("open theme %s: %w", request.Name, err)
	}
	pkg, warnings, err = loadPackage(ctx, source, request)
	if err != nil {
		return pkg, warnings, fmt.Errorf("theme %s: %w", request.Name, err)
	}
	diagnostics, err := CheckCompatibility(pkg.Manifest, request.Version)
	warnings = append(warnings, diagnostics...)
	return pkg, warnings, err
}

func loadPackage(ctx context.Context, source fs.FS, request Request) (Package, []string, error) {
	pkg := Package{
		Templates: make(map[string]string), assetURLs: make(map[string]string),
		language: request.Language, labels: make(map[string]string),
	}
	files := make(map[string][]byte)
	if err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("read %s: %w", name, walkErr)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if !entry.IsDir() && !entry.Type().IsRegular() {
			return fmt.Errorf("%s must be a regular file, not a symlink or special file", name)
		}
		if entry.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		files[name] = data
		return nil
	}); err != nil {
		return pkg, nil, err
	}
	manifest, warnings, err := ParseManifest(files["theme.json"])
	if err != nil {
		return pkg, warnings, err
	}
	if request.Name != "" && manifest.Name != request.Name {
		return pkg, warnings, fmt.Errorf("theme.json: name %q does not match selected name %q", manifest.Name, request.Name)
	}
	pkg.Manifest = manifest
	for _, name := range sortedFiles(files) {
		data := files[name]
		if strings.HasPrefix(name, "templates/") && strings.HasSuffix(name, ".gohtml") {
			if !utf8.Valid(data) {
				return pkg, warnings, fmt.Errorf("%s must be UTF-8", name)
			}
			pkg.Templates[strings.TrimPrefix(name, "templates/")] = string(data)
		}
		if strings.HasPrefix(name, "public/") {
			relative := strings.TrimPrefix(name, "public/")
			destination := path.Join("themes", manifest.Name, relative)
			pkg.assetURLs[relative] = strings.TrimSuffix(request.Prefix, "/") + (&url.URL{Path: "/" + destination}).EscapedPath()
			pkg.Assets = append(pkg.Assets, Asset{
				Source: "theme " + manifest.Name + "/" + name, Path: destination, Data: data,
			})
		}
	}
	for _, name := range []string{"page.gohtml", "landing.gohtml", "list.gohtml"} {
		if _, exists := pkg.Templates[name]; !exists {
			return pkg, warnings, fmt.Errorf("missing required templates/%s", name)
		}
	}
	if data, exists := files["locales/"+request.Language+".toml"]; exists {
		if !utf8.Valid(data) {
			return pkg, warnings, fmt.Errorf("locale %s must be UTF-8", request.Language)
		}
		if err := toml.Unmarshal(data, &pkg.labels); err != nil {
			return pkg, warnings, fmt.Errorf("parse locale %s as plain-text label map: %w", request.Language, err)
		}
	}
	return pkg, warnings, nil
}

// AssetURL returns an existing public asset URL without exposing private sources.
func (pkg Package) AssetURL(name string) (string, error) {
	if !fs.ValidPath(name) || name == "." || strings.Contains(name, `\`) {
		return "", fmt.Errorf("theme asset path %q is invalid", name)
	}
	value, exists := pkg.assetURLs[name]
	if !exists {
		return "", fmt.Errorf("theme %s: public asset %q does not exist", pkg.Manifest.Name, name)
	}
	return value, nil
}

// Label returns a requested plain-text label with exact-language matching.
func (pkg Package) Label(key string) (string, error) {
	value, exists := pkg.labels[key]
	if !exists {
		return "", fmt.Errorf("theme %s: label %q is missing for exact language %q", pkg.Manifest.Name, key, pkg.language)
	}
	return value, nil
}
