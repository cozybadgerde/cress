package scaffold_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/scaffold"
)

func TestInitializedSiteBuildsWithBothThemes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, name := range []string{"poppy", "clover"} {
		t.Run(name, func(t *testing.T) {
			site := filepath.Join(t.TempDir(), "site")
			if err := scaffold.Init(context.Background(), site, name); err != nil {
				t.Fatal(err)
			}
			result, err := generator.Build(context.Background(), site)
			if err != nil || result.Pages != 7 || len(result.Warnings) != 1 {
				t.Fatalf("initialized build: %+v, %v", result, err)
			}
			if _, err := generator.Build(context.Background(), site); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"index.html", "services/index.html", "journal/list/index.html", "journal/welcome/index.html", "tour/index.html"} {
				data, err := os.ReadFile(filepath.Join(site, "public", file))
				if err != nil || !strings.Contains(string(data), `themes/`+name+`/site.css`) ||
					!strings.Contains(string(data), `id="main"`) {
					t.Fatalf("%s: %s, %v", file, data, err)
				}
			}
			tour, err := os.ReadFile(filepath.Join(site, "public/tour/index.html"))
			if err != nil {
				t.Fatal(err)
			}
			for _, example := range []string{
				"<em>emphasis</em>", "<strong>strong text</strong>",
				"<h2>", "<h3>", "<h4>", "<h5>", "<h6>",
				"<ul>", "<ol>", "<blockquote>", "<hr", "<br", "<table>",
				`href="/services/"`, `href="/contact/"`, `src="/seedling.svg"`,
				"<figcaption>", `class="captioned-image"`, `<span class="kn">import</span>`,
				"&lt;p&gt;This is code", "<pre><code>cress site build",
				`<abbr title="Frequently asked questions">FAQ</abbr>`,
				"<p>This paragraph is an authored HTML block.</p>",
				`datetime="2026-10-23"`,
			} {
				if !strings.Contains(string(tour), example) {
					t.Errorf("tour missing rendered example %q", example)
				}
			}
			data, err := os.ReadFile(filepath.Join(site, ".gitignore"))
			if err != nil || string(data) != "/public/\n/.cress/\n" {
				t.Fatalf("ignore rules: %s, %v", data, err)
			}
			before, err := os.ReadFile(filepath.Join(site, "cress.toml"))
			if err != nil {
				t.Fatal(err)
			}
			if err := scaffold.Init(context.Background(), site, name); err == nil {
				t.Fatal("nonempty site overwritten")
			}
			after, err := os.ReadFile(filepath.Join(site, "cress.toml"))
			if err != nil || string(after) != string(before) {
				t.Fatal("existing site changed")
			}
		})
	}
}

func TestInitializedConfigOptionalExamples(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, name := range []string{"poppy", "clover"} {
		t.Run(name, func(t *testing.T) {
			site := filepath.Join(t.TempDir(), "site")
			if err := scaffold.Init(context.Background(), site, name); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(site, "cress.toml"))
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(string(data), "\n")
			for index, line := range lines {
				example, commented := strings.CutPrefix(line, "# ")
				key, _, assignment := strings.Cut(example, " = ")
				if commented && (strings.HasPrefix(example, "[") ||
					(assignment && !strings.ContainsAny(key, " \t"))) {
					lines[index] = example
				}
			}
			config, warnings, err := content.ParseConfig([]byte(strings.Join(lines, "\n")))
			if err != nil || len(warnings) != 0 {
				t.Fatalf("optional examples: warnings %v, error %v", warnings, err)
			}
			if config.Theme != name || !config.Highlight ||
				config.Host != "https://example.org" ||
				config.Site.Copyright != "Copyright Your business" ||
				config.Image == nil || config.Image.URL != "/images/team.jpg" ||
				len(config.Image.Caption) == 0 ||
				config.Site.Brand.Color != "#346b38" ||
				config.Site.Brand.Favicon != "/images/favicon.svg" ||
				config.Site.Brand.Logo == nil || config.Site.Brand.Logo.URL != "/images/logo.svg" {
				t.Fatalf("optional examples not applied: %+v", config)
			}
		})
	}
}

func TestInitRefusesUnsafeTargetsAndCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, kind := range []string{"file", "symlink", "hidden file", "invalid theme", "cancelled"} {
		t.Run(kind, func(t *testing.T) {
			base := t.TempDir()
			target := filepath.Join(base, "target")
			protected := t.TempDir()
			switch kind {
			case "file":
				if err := os.WriteFile(target, []byte("keep"), 0o644); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Symlink(protected, target); err != nil {
					t.Fatal(err)
				}
			case "hidden file":
				if err := os.Mkdir(target, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(target, ".keep"), []byte("keep"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			name := "poppy"
			if kind == "invalid theme" {
				name = "../poppy"
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if kind == "cancelled" {
				cancel()
			}
			err := scaffold.Init(ctx, target, name)
			if err == nil || (kind == "cancelled" && !errors.Is(err, context.Canceled)) {
				t.Fatalf("unsafe initialization accepted: %v", err)
			}
			entries, err := os.ReadDir(protected)
			if err != nil || len(entries) != 0 {
				t.Fatal("symlink target modified")
			}
		})
	}
}
