package generator_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/scaffold"
	"github.com/cozybadgerde/cress/internal/theme"
)

func TestAnemoneImageFreePagesAndLocalizedPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := t.TempDir()
	if err := scaffold.Init(context.Background(), site, "anemone"); err != nil {
		t.Fatal(err)
	}
	config := "[site]\nlanguage = 'de'\npath = '/company/'\n[theme]\nname = 'anemone'\n"
	if err := os.WriteFile(filepath.Join(site, "cress.toml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(site, "content/index.md"), []byte(
		"+++\ntitle = '<Quiet work>'\ntype = 'landing'\n+++\nAn introduction without a photo.\n\n[Services](services.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(site, "content/empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(site, "content/empty/list.md"), []byte(
		"+++\ntitle = 'Empty collection'\ntype = 'list'\n+++\nAn introduction without entries.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := generator.Build(context.Background(), site); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"index.html", "about/index.html", "empty/list/index.html", "journal/welcome/index.html"} {
		data, err := os.ReadFile(filepath.Join(site, "public", file))
		if err != nil {
			t.Fatal(err)
		}
		html := string(data)
		for _, want := range []string{
			`<html lang="de">`, `href="/company/themes/anemone/site.css"`,
			`href="#main">Zum Inhalt</a>`,
		} {
			if !strings.Contains(html, want) {
				t.Errorf("%s missing %s", file, want)
			}
		}
		notFound, err := os.ReadFile(filepath.Join(site, "public", "404.html"))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{
			`<html lang="de">`, `<h1>Seite nicht gefunden</h1>`,
			"Die angeforderte Seite wurde nicht gefunden.",
			`<a href="/company/">Startseite</a>`,
		} {
			if !strings.Contains(string(notFound), want) {
				t.Errorf("localized 404 missing %q: %s", want, notFound)
			}
		}
		for _, absent := range []string{`class="site-header"`, `class="site-footer"`, `<img`, `class="entries"`, `with-image`} {
			if strings.Contains(html, absent) {
				t.Errorf("%s rendered empty optional group %s", file, absent)
			}
		}
		switch file {
		case "index.html":
			if !strings.Contains(html, `<h1 id="page-title">&lt;Quiet work&gt;</h1>`) ||
				!strings.Contains(html, `href="/company/services/">Services</a>`) {
				t.Errorf("landing lost escaped title or resolved action: %s", html)
			}
		case "empty/list/index.html":
			if !strings.Contains(html, "<p>An introduction without entries.</p>") {
				t.Errorf("empty list lost introduction: %s", html)
			}
		case "journal/welcome/index.html":
			if !strings.Contains(html, `Veröffentlicht <time datetime=`) {
				t.Errorf("article lost localized date: %s", html)
			}
		}
	}
}

func TestBuiltinIdentityHomeLink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, name := range []string{"poppy", "clover", "anemone"} {
		for _, fixture := range []struct {
			name, metadata, want string
		}{
			{"combined", "title = '<Studio>'\n[brand.logo]\nsrc = '/images/logo.svg'\nalt = '<Studio logo>'",
				`<a class="identity" href="/company/"><img class="logo" src="/company/images/logo.svg" alt=""><span class="site-title">&lt;Studio&gt;</span></a>`},
			{"title only", "title = '<Studio>'",
				`<a class="identity" href="/company/"><span class="site-title">&lt;Studio&gt;</span></a>`},
			{"logo only", "[brand.logo]\nsrc = '/images/logo.svg'\nalt = '<Studio logo>'",
				`<a class="identity" href="/company/"><img class="logo" src="/company/images/logo.svg" alt="&lt;Studio logo&gt;"></a>`},
			{"decorative logo only", "[brand.logo]\nsrc = '/images/logo.svg'\nalt = ''",
				`<a class="identity" href="/company/" aria-label="Home"><img class="logo" src="/company/images/logo.svg" alt=""></a>`},
			{"localized decorative logo", "language = 'de'\n[brand.logo]\nsrc = '/images/logo.svg'\nalt = ''",
				`<a class="identity" href="/company/" aria-label="Startseite"><img class="logo" src="/company/images/logo.svg" alt=""></a>`},
			{"navigation only", "[navigation]\nmain = [{ label = 'About', url = '/about/' }]", ""},
		} {
			t.Run(name+"/"+fixture.name, func(t *testing.T) {
				site := t.TempDir()
				if err := scaffold.Init(context.Background(), site, name); err != nil {
					t.Fatal(err)
				}
				config := "[theme]\nname = '" + name + "'\n[site]\npath = '/company/'\n" + fixture.metadata
				if err := os.WriteFile(filepath.Join(site, "cress.toml"), []byte(config), 0o644); err != nil {
					t.Fatal(err)
				}
				if _, err := generator.Build(context.Background(), site); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(filepath.Join(site, "public/index.html"))
				if err != nil {
					t.Fatal(err)
				}
				html := string(data)
				start := strings.Index(html, `<a class="identity"`)
				if fixture.want == "" {
					if start >= 0 {
						t.Fatal("navigation-only header has an empty identity link")
					}
					return
				}
				if start < 0 {
					t.Fatal("missing identity home link")
				}
				link := html[start : start+strings.Index(html[start:], "</a>")+4]
				link = strings.Join(strings.Fields(link), " ")
				link = strings.ReplaceAll(link, "> <", "><")
				if link != fixture.want || strings.Count(html, `class="identity"`) != 1 {
					t.Errorf("identity link = %s; want %s", link, fixture.want)
				}
			})
		}
	}
}

func TestBuiltinCompletePageStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, name := range []string{"poppy", "clover", "anemone"} {
		site := t.TempDir()
		if err := scaffold.Init(context.Background(), site, name); err != nil {
			t.Fatal(err)
		}
		config, err := os.ReadFile(filepath.Join(site, "cress.toml"))
		if err != nil {
			t.Fatal(err)
		}
		config = []byte(strings.Replace(string(config), "[brand]\n", "[brand]\ncolor = '#a8231a'\n", 1))
		if err := os.WriteFile(filepath.Join(site, "cress.toml"), config, 0o644); err != nil {
			t.Fatal(err)
		}
		for _, fixture := range []struct{ file, metadata, image string }{
			{"index.md", "type = 'landing'\n", "index.assets/placeholder.svg"},
			{"journal/list.md", "type = 'list'\n[list]\norder = 'title'\n", "../index.assets/placeholder.svg"},
			{"journal/welcome.md", "", "../index.assets/placeholder.svg"},
		} {
			source := "+++\ntitle = 'Journal image'\n" +
				"image = { src = '" + fixture.image + "', alt = 'Placeholder', caption = 'Image credit' }\n" +
				fixture.metadata + "+++\n"
			if err := os.WriteFile(filepath.Join(site, "content", fixture.file), []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := generator.Build(context.Background(), site); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"index.html", "services/index.html", "about/index.html", "contact/index.html", "journal/list/index.html", "journal/welcome/index.html"} {
			data, err := os.ReadFile(filepath.Join(site, "public", file))
			if err != nil {
				t.Fatal(err)
			}
			html := string(data)
			for _, invariant := range []string{`<html lang="en">`, `href="#main"`, `id="main" tabindex="-1"`, `aria-label="Main"`, `aria-label="Footer"`, `<style>:root { --link: #a8231a; }</style>`} {
				if file == "index.html" && invariant == `id="main" tabindex="-1"` {
					invariant = `id="main" class="landing" tabindex="-1"`
				}
				if !strings.Contains(html, invariant) {
					t.Errorf("%s/%s missing %s", name, file, invariant)
				}
			}
			if strings.Count(html, "<h1") != 1 || strings.Count(html, "<main ") != 1 ||
				strings.Contains(html, "<script") {
				t.Errorf("%s/%s invalid landmarks/headings or unexpected scripts", name, file)
			}
			if file == "index.html" || strings.HasPrefix(file, "journal/") {
				lazy := strings.Count(html, `loading="lazy" decoding="async"`)
				wantLazy := 0
				if file == "journal/list/index.html" {
					wantLazy = 1
				}
				if lazy != wantLazy || strings.Count(html, `alt="Placeholder">`) != 1 ||
					strings.Count(html, "<figcaption>Image credit</figcaption>") != wantLazy+1 {
					t.Errorf("%s/%s: incorrect image loading or lost caption: %s", name, file, html)
				}
				if wantLazy == 1 && strings.Index(html, `alt="Placeholder">`) > strings.Index(html, `loading="lazy"`) {
					t.Errorf("%s/%s: list page image must load eagerly before lazy entry images", name, file)
				}
			}
		}
		notFound, err := os.ReadFile(filepath.Join(site, "public", "404.html"))
		if err != nil || !strings.Contains(string(notFound), `<html lang="en">`) ||
			!strings.Contains(string(notFound), "<h1>Page not found</h1>") {
			t.Errorf("%s missing its localized 404 page: %s, %v", name, notFound, err)
		}
		if err := theme.Init(context.Background(), site, name); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(site, "themes", name, "templates/page.gohtml")
		if err := os.WriteFile(target, []byte("Local package: {{.Title}}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := generator.Build(context.Background(), site); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(site, "public/about/index.html"))
		if err != nil || string(data) != "Local package: About" {
			t.Fatalf("local package did not override built-in: %s, %v", data, err)
		}
		if err := os.WriteFile(target, []byte("{{missing}}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := generator.Build(context.Background(), site); err == nil {
			t.Fatal("invalid local package fell back")
		}
	}
}

func TestBuiltinLinkPreviewsAndProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, name := range []string{"poppy", "clover", "anemone"} {
		site := t.TempDir()
		if err := scaffold.Init(context.Background(), site, name); err != nil {
			t.Fatal(err)
		}
		config := "[site]\ntitle = 'Shop & Co'\nhost = 'https://example.org'\npath = '/company/'\n[theme]\nname = '" + name + "'\n" +
			"[profile]\ntype = 'Bakery'\n"
		if err := os.WriteFile(filepath.Join(site, "cress.toml"), []byte(config), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(site, "content/about.md"), []byte(
			"+++\ntitle = 'About <us>'\ndescription = 'Who we are'\n"+
				"image = { src = 'index.assets/placeholder.svg', alt = 'Our team' }\n+++\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := generator.Build(context.Background(), site); err != nil {
			t.Fatal(err)
		}
		about, err := os.ReadFile(filepath.Join(site, "public/about/index.html"))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{
			`<meta property="og:type" content="website">`,
			`<meta property="og:url" content="https://example.org/company/about/">`,
			`<meta property="og:title" content="About &lt;us&gt;">`,
			`<meta property="og:site_name" content="Shop &amp; Co">`,
			`<meta property="og:description" content="Who we are">`,
			`<meta property="og:image" content="https://example.org/company/placeholder.svg">`,
			`<meta property="og:image:alt" content="Our team">`,
			`<meta name="twitter:card" content="summary_large_image">`,
		} {
			if !strings.Contains(string(about), want) {
				t.Errorf("%s missing %s", name, want)
			}
		}
		home, err := os.ReadFile(filepath.Join(site, "public/index.html"))
		if err != nil || !strings.Contains(string(home), `<script type="application/ld+json">{"@context":"https://schema.org","@type":"Bakery","name":"Shop \u0026 Co","url":"https://example.org/company/"}</script>`) ||
			strings.Contains(string(about), "ld+json") {
			t.Errorf("%s structured data missing on home or present elsewhere: %s, %v", name, home, err)
		}
		notFound, err := os.ReadFile(filepath.Join(site, "public/404.html"))
		if err != nil || strings.Contains(string(notFound), "og:") {
			t.Errorf("%s 404 has link preview or failed: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(site, "cress.toml"), []byte("[theme]\nname = '"+name+"'\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		result, err := generator.Build(context.Background(), site)
		if err != nil {
			t.Fatal(err)
		}
		about, err = os.ReadFile(filepath.Join(site, "public/about/index.html"))
		if err != nil || strings.Contains(string(about), "og:") || strings.Contains(string(about), "twitter:") ||
			!strings.Contains(strings.Join(result.Warnings, "\n"), "link previews") {
			t.Errorf("%s without host: preview tags or missing warning %v: %v", name, result.Warnings, err)
		}
	}
}
