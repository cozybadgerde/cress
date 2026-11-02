package theme

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

type privateFileFS struct {
	fs.FS
	private  map[string]bool
	openErr  error
	closeErr error
}

func (source privateFileFS) Open(name string) (fs.File, error) {
	if source.private[name] && source.openErr != nil {
		return nil, source.openErr
	}
	file, err := source.FS.Open(name)
	if err != nil || !source.private[name] {
		return file, err
	}
	return privateFile{File: file, closeErr: source.closeErr}, nil
}

type privateFile struct {
	fs.File
	closeErr error
}

func (file privateFile) Read([]byte) (int, error) {
	return 0, errors.New("private file body must not be read")
}

func (file privateFile) Close() error {
	return errors.Join(file.File.Close(), file.closeErr)
}

const validManifest = `{"name":"demo","version":"1.2.3","cress":{"minimum":"1.0.0","major":1}}`

func TestManifestAndCompatibility(t *testing.T) {
	manifest, warnings, err := ParseManifest([]byte(validManifest))
	if err != nil || len(warnings) != 0 {
		t.Fatalf("manifest=%+v warnings=%v err=%v", manifest, warnings, err)
	}
	for _, test := range []struct {
		version string
		valid   bool
	}{
		{"dev", true}, {"v1.0.0", true}, {"1.2.0", true}, {"1.0.0+build", true},
		{"1.0.0-rc.1", false}, {"2.0.0", false}, {"0.9.0", false}, {"junk", false},
	} {
		warnings, err := CheckCompatibility(manifest, test.version)
		if (err == nil) != test.valid || (len(warnings) == 1) != (test.version == "dev") {
			t.Errorf("version=%s warnings=%v error=%v", test.version, warnings, err)
		}
	}
	manifest.Cress.Minimum = "1.0.1"
	if _, err := CheckCompatibility(manifest, "dev"); err == nil {
		t.Fatal("development runtime bypassed minimum")
	}
	for _, source := range []string{
		`{`, `null`, `{}`, strings.Replace(validManifest, `"major":1`, `"major":2`, 1),
		strings.Replace(validManifest, `"1.2.3"`, `"01.2.3"`, 1),
		strings.Replace(validManifest, `"1.2.3"`, `"1.2.3-01"`, 1),
		strings.Replace(validManifest, `"demo"`, `"../demo"`, 1),
	} {
		if _, _, err := ParseManifest([]byte(source)); err == nil {
			t.Errorf("accepted %s", source)
		}
	}
	_, warnings, err = ParseManifest([]byte(strings.Replace(validManifest, `"major":1`, `"major":1,"future":true`, 1)))
	if err != nil || len(warnings) != 1 {
		t.Fatalf("unknown field warnings=%v err=%v", warnings, err)
	}
}

func TestManifestUnknownCaseVariantsDoNotOverrideKnownFields(t *testing.T) {
	source := `{"name":"demo","Name":"other","version":"1.2.3","Version":null,` +
		`"cress":{"minimum":"1.0.0","Minimum":null,"major":1,"Major":2}}`
	manifest, warnings, err := ParseManifest([]byte(source))
	if err != nil || len(warnings) != 4 || manifest.Name != "demo" ||
		manifest.Version != "1.2.3" || manifest.Cress.Minimum != "1.0.0" || manifest.Cress.Major != 1 {
		t.Fatalf("unknown fields affected known values: %+v, %v, %v", manifest, warnings, err)
	}
}

func TestPackageHelpersAndBoundaries(t *testing.T) {
	files := fstest.MapFS{
		"theme.json":               {Data: []byte(validManifest)},
		"templates/page.gohtml":    {Data: []byte(`{{label "more"}}`)},
		"templates/landing.gohtml": {Data: []byte(`Landing`)},
		"templates/list.gohtml":    {Data: []byte(`List`)},
		"public/css/site.css":      {Data: []byte("unchanged")},
		"private.txt":              {Data: []byte("private")},
		"locales/en.toml":          {Data: []byte("more = '<More>'")},
	}

	pkg, _, err := loadPackage(context.Background(), files, Request{Name: "demo", Language: "en", Prefix: "/company/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(pkg.Assets) != 1 || string(pkg.Assets[0].Data) != "unchanged" ||
		pkg.Assets[0].Path != "themes/demo/css/site.css" {
		t.Fatalf("assets=%+v", pkg.Assets)
	}
	if value, err := pkg.AssetURL("css/site.css"); err != nil || value != "/company/themes/demo/css/site.css" {
		t.Fatalf("asset=%s err=%v", value, err)
	}
	if value, err := pkg.Label("more"); err != nil || value != "<More>" {
		t.Fatalf("label=%s err=%v", value, err)
	}
	for _, name := range []string{"../private.txt", "private.txt", "/css/site.css", `css\site.css`} {
		if _, err := pkg.AssetURL(name); err == nil {
			t.Errorf("private/invalid asset accepted: %s", name)
		}
	}
	if _, err := pkg.Label("missing"); err == nil {
		t.Fatal("missing label defaulted")
	}
	pkg, _, err = loadPackage(context.Background(), files, Request{Name: "demo", Language: "en-US", Prefix: "/"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pkg.Label("more"); err == nil {
		t.Fatal("regional locale fell back")
	}
	for _, value := range []string{`more = 1`, "[more]\ntext = 'nested'"} {
		files["locales/en-US.toml"] = &fstest.MapFile{Data: []byte(value)}
		if _, _, err := loadPackage(context.Background(), files, Request{Language: "en-US"}); err == nil {
			t.Errorf("accepted non-text labels: %s", value)
		}
	}
	delete(files, "locales/en-US.toml")
	delete(files, "templates/list.gohtml")
	if _, _, err := loadPackage(context.Background(), files, Request{}); err == nil {
		t.Fatal("incomplete theme accepted")
	}
}

func TestPackageValidatesPrivateFilesWithoutReadingBodies(t *testing.T) {
	files := fstest.MapFS{
		"theme.json":               {Data: []byte(validManifest)},
		"templates/page.gohtml":    {Data: []byte("Page")},
		"templates/landing.gohtml": {Data: []byte("Landing")},
		"templates/list.gohtml":    {Data: []byte("List")},
		"public/site.css":          {Data: []byte("unchanged")},
		"locales/en.toml":          {Data: []byte("more = 'More'")},
		"locales/de.toml":          {Data: []byte("unused")},
		"templates/design.txt":     {Data: []byte("private")},
		"originals/image.png":      {Data: []byte("private")},
	}
	private := map[string]bool{
		"locales/de.toml": true, "templates/design.txt": true, "originals/image.png": true,
	}
	for _, test := range []struct {
		name     string
		openErr  error
		closeErr error
	}{
		{name: "unused bodies are not read"},
		{name: "unreadable private file", openErr: fs.ErrPermission},
		{name: "private close failure", closeErr: errors.New("close failed")},
	} {
		t.Run(test.name, func(t *testing.T) {
			pkg, _, err := loadPackage(context.Background(), privateFileFS{
				FS: files, private: private, openErr: test.openErr, closeErr: test.closeErr,
			}, Request{Name: "demo", Language: "en", Prefix: "/"})
			wantErr := errors.Join(test.openErr, test.closeErr)
			if wantErr != nil {
				if err == nil || !strings.Contains(err.Error(), "locales/de.toml") ||
					!strings.Contains(err.Error(), wantErr.Error()) {
					t.Fatalf("private file failure lost: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(pkg.Templates) != 3 || pkg.Templates["page.gohtml"] != "Page" ||
				len(pkg.Assets) != 1 || string(pkg.Assets[0].Data) != "unchanged" {
				t.Fatalf("generation inputs changed: %+v", pkg)
			}
			if label, err := pkg.Label("more"); err != nil || label != "More" {
				t.Fatalf("selected locale changed: %q, %v", label, err)
			}
		})
	}
}
