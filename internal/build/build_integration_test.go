package build_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cozybadgerde/cress/internal/build"
	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/scaffold"
	"github.com/cozybadgerde/cress/internal/theme"
	"github.com/cozybadgerde/cress/internal/version"
)

// scaffoldPages is how many pages `cress init` produces: index, about, imprint,
// privacy, and the three under guides/. The tests below build a scaffolded site
// and count what comes out, so adding or removing a starter page moves every
// expectation here at once.
const scaffoldPages = 7

func TestBuild_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if res.Pages != scaffoldPages {
		t.Errorf("rendered %d pages, want %d (the whole scaffold)", res.Pages, scaffoldPages)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", res.Warnings)
	}

	out := filepath.Join(root, config.OutputDir)
	// style.css comes from the theme; favicon.png and logo.svg from the site's
	// static/ tree that `cress init` scaffolds.
	for _, name := range []string{"index.html", "about.html", "style.css", "favicon.png", "logo.svg"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("missing output %s: %v", name, err)
		}
	}

	assertMarkers(t, "index.html", readFile(t, filepath.Join(out, "index.html")), []marker{
		// The home page is the site, so its title is the site's name alone rather
		// than naming it twice.
		{"<title>My cozy site</title>", "the home page title"},
		{`<link rel="icon" href="/favicon.png"`, "the default favicon link"},
		{`class="site-logo"`, "the nav logo"},
		{`href="/about.html"`, "the nav link to about"},
		{`class="footer-nav"`, "the footer nav"},
		{`href="/imprint.html"`, "the footer nav link to imprint"},
		{"Fresh little sites, fast.", "the rendered body content"},
	})

	assertMarkers(t, "about.html", readFile(t, filepath.Join(out, "about.html")), []marker{
		{`aria-current="page"`, "its own nav entry marked active"},
		{`<code class="language-go">`, "the fenced code block's language class"},
	})
}

func TestBuild_accentOverrides_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	const dark = "@media (prefers-color-scheme: dark)"
	for _, tc := range []struct {
		name       string
		site       string
		want, omit []string
	}{
		// The scaffolded site sets no accent, so the theme's own per-scheme pair
		// in style.css is left to apply. Emitting a :root block here is the bug
		// that made both of the theme's values unreachable.
		{name: "unset", site: "", omit: []string{"--accent:", dark}},
		{name: "light only", site: "accent = \"#3e2723\"\n", want: []string{"--accent: #3e2723"}, omit: []string{dark}},
		{
			name: "both",
			site: "accent = \"#3e2723\"\naccent_dark = \"#d7b8a3\"\n",
			want: []string{"--accent: #3e2723", dark, "--accent: #d7b8a3"},
		},
		{name: "dark only", site: "accent_dark = \"#d7b8a3\"\n", want: []string{dark, "--accent: #d7b8a3"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if _, err := scaffold.Create(root, false); err != nil {
				t.Fatalf("scaffold: %v", err)
			}
			writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\n"+tc.site)

			if _, err := build.Build(build.Options{Root: root}); err != nil {
				t.Fatalf("build: %v", err)
			}
			doc := readFile(t, filepath.Join(root, config.OutputDir, "index.html"))

			for _, want := range tc.want {
				if !strings.Contains(doc, want) {
					t.Errorf("rendered page is missing %q", want)
				}
			}
			for _, omit := range tc.omit {
				if strings.Contains(doc, omit) {
					t.Errorf("rendered page should not contain %q", omit)
				}
			}
		})
	}
}

// The logo renders as an <img> inside a <picture> whether or not logo_dark is
// set; only the dark-scheme <source> is conditional. One shape keeps the
// template free of a second branch, and it costs the layout nothing because the
// theme CSS gives .site-title picture display: contents.
func TestBuild_logoVariants_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	const picture = "<picture>"
	for _, tc := range []struct {
		name       string
		site       string
		want, omit []string
	}{
		{name: "no logo", site: "", omit: []string{"site-logo", picture}},
		{
			name: "logo only",
			site: "logo = \"/logo.svg\"\n",
			want: []string{picture, `<img class="site-logo" src="/logo.svg"`},
			omit: []string{"<source"},
		},
		{
			name: "both",
			site: "logo = \"/logo.svg\"\nlogo_dark = \"/logo-dark.svg\"\n",
			want: []string{
				picture,
				`<source media="(prefers-color-scheme: dark)" srcset="/logo-dark.svg"`,
				`<img class="site-logo" src="/logo.svg"`,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if _, err := scaffold.Create(root, false); err != nil {
				t.Fatalf("scaffold: %v", err)
			}
			writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\n"+tc.site)

			if _, err := build.Build(build.Options{Root: root}); err != nil {
				t.Fatalf("build: %v", err)
			}
			doc := readFile(t, filepath.Join(root, config.OutputDir, "index.html"))

			for _, want := range tc.want {
				if !strings.Contains(doc, want) {
					t.Errorf("rendered page is missing %q", want)
				}
			}
			for _, omit := range tc.omit {
				if strings.Contains(doc, omit) {
					t.Errorf("rendered page should not contain %q", omit)
				}
			}
		})
	}
}

func TestBuild_draftsAndStatic_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	writeSiteFile(t, filepath.Join(root, "content", "secret.md"), "---\ntitle: Secret\ndraft: true\n---\nhidden\n")
	writeSiteFile(t, filepath.Join(root, "static", "robots.txt"), "User-agent: *\n")

	// Default build: drafts excluded, site static copied.
	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if res.Pages != scaffoldPages {
		t.Errorf("rendered %d pages, want %d (draft excluded)", res.Pages, scaffoldPages)
	}
	out := filepath.Join(root, config.OutputDir)
	if _, err := os.Stat(filepath.Join(out, "secret.html")); !os.IsNotExist(err) {
		t.Errorf("draft page should not be rendered (stat err = %v)", err)
	}
	if _, err := os.Stat(filepath.Join(out, "robots.txt")); err != nil {
		t.Errorf("site static file should be copied: %v", err)
	}

	// With drafts enabled the page appears.
	res, err = build.Build(build.Options{Root: root, Drafts: true})
	if err != nil {
		t.Fatalf("build --drafts: %v", err)
	}
	if res.Pages != scaffoldPages+1 {
		t.Errorf("rendered %d pages with drafts, want %d", res.Pages, scaffoldPages+1)
	}
	if _, err := os.Stat(filepath.Join(out, "secret.html")); err != nil {
		t.Errorf("draft page should render with --drafts: %v", err)
	}
}

func TestBuild_neverDeletesOutput_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	if _, err := build.Build(build.Options{Root: root}); err != nil {
		t.Fatalf("first build: %v", err)
	}

	// Two ways a file ends up in the output tree without this build writing it:
	// the user put it there, and a page that used to produce it is now gone.
	out := filepath.Join(root, config.OutputDir)
	writeSiteFile(t, filepath.Join(out, "keepme.txt"), "not cress's\n")
	if err := os.Remove(filepath.Join(root, "content", "about.md")); err != nil {
		t.Fatalf("removing about.md: %v", err)
	}

	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if res.Pages != scaffoldPages-1 {
		t.Errorf("rendered %d pages, want %d after removing about.md", res.Pages, scaffoldPages-1)
	}

	for _, name := range []string{"keepme.txt", "about.html"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("%s must survive a rebuild; a build never deletes: %v", name, err)
		}
	}
	if got := readFile(t, filepath.Join(out, "keepme.txt")); got != "not cress's\n" {
		t.Errorf("keepme.txt = %q, want it byte-for-byte untouched", got)
	}
	assertWarns(t, res.Warnings, "keepme.txt", "about.html")
}

func TestBuild_ignoresDotEntriesInOutput_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	if _, err := build.Build(build.Options{Root: root}); err != nil {
		t.Fatalf("first build: %v", err)
	}

	// What a host or a gh-pages worktree leaves in the output tree. These belong
	// to the user, so warning about them on every single build would train the
	// warnings to be ignored.
	out := filepath.Join(root, config.OutputDir)
	writeSiteFile(t, filepath.Join(out, ".nojekyll"), "")
	writeSiteFile(t, filepath.Join(out, ".git", "config"), "[core]\n")

	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("dot-entries must not warn, got: %v", res.Warnings)
	}
}

func TestBuild_collapsesManyStaleWarnings_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	if _, err := build.Build(build.Options{Root: root}); err != nil {
		t.Fatalf("first build: %v", err)
	}

	// One past the point where individual warnings collapse into a count, so a
	// bulk rename cannot bury the rest of the build's output.
	const stale = 11
	out := filepath.Join(root, config.OutputDir)
	for i := 0; i < stale; i++ {
		writeSiteFile(t, filepath.Join(out, fmt.Sprintf("old-%d.html", i)), "old\n")
	}

	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("want a single summary warning, got %d: %v", len(res.Warnings), res.Warnings)
	}
	if !strings.Contains(res.Warnings[0], "11 file(s)") {
		t.Errorf("summary warning = %q, want it to count all 11 files", res.Warnings[0])
	}
}

func TestBuild_footerAndCopyright_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// The credit's link, not its tagline: the scaffolded index.md says "Fresh
	// little sites, fast." in its own body, so matching that would pass whether
	// or not the footer rendered.
	const credit = `<a href="https://github.com/cozybadgerde/cress">Cress</a>`
	thisYear := strconv.Itoa(time.Now().Year())

	for _, tc := range []struct {
		name       string
		site       string
		want, omit []string
	}{
		{
			name: "unset keeps the credit and omits the copyright",
			omit: []string{"footer-copyright"},
			want: []string{credit},
		},
		{
			name: "a footer message replaces the credit",
			site: "footer = \"A small corner of the internet.\"\n",
			want: []string{"A small corner of the internet."},
			omit: []string{credit},
		},
		{
			name: "the year token expands",
			site: "copyright = \"(c) {year} Cozy Badger\"\n",
			want: []string{`class="footer-copyright"`, "(c) " + thisYear + " Cozy Badger"},
		},
		{
			name: "both render together",
			site: "footer = \"Handmade.\"\ncopyright = \"(c) {year} Cozy Badger\"\n",
			want: []string{"Handmade.", "(c) " + thisYear + " Cozy Badger"},
			omit: []string{credit},
		},
		// The footer is documented as plain text. Escaping is what makes that
		// true, so a site config can never inject markup into every page.
		{
			name: "markup in the footer is escaped, not rendered",
			site: "footer = \"<b>bold</b> & <script>alert(1)</script>\"\n",
			want: []string{"&lt;b&gt;bold&lt;/b&gt; &amp; &lt;script&gt;"},
			omit: []string{"<b>bold</b>", "<script>alert(1)</script>"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if _, err := scaffold.Create(root, false); err != nil {
				t.Fatalf("scaffold: %v", err)
			}
			writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\n"+tc.site)

			if _, err := build.Build(build.Options{Root: root}); err != nil {
				t.Fatalf("build: %v", err)
			}
			doc := readFile(t, filepath.Join(root, config.OutputDir, "index.html"))

			for _, want := range tc.want {
				if !strings.Contains(doc, want) {
					t.Errorf("rendered page is missing %q", want)
				}
			}
			for _, omit := range tc.omit {
				if strings.Contains(doc, omit) {
					t.Errorf("rendered page should not contain %q", omit)
				}
			}
		})
	}
}

// A project page on GitHub or GitLab serves the site from a subdirectory, so
// every URL cress emits has to be rooted under the path in base_url: the nav,
// the theme's own assets, the branding, and the links inside content.
func TestBuild_subpath_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	writeSiteFile(t, filepath.Join(root, "cress.toml"), `
[site]
title = "S"
base_url = "https://user.github.io/cress"
logo = "/logo.svg"
favicon = "/favicon.png"

[nav.main]
Home = "index.md"
About = "about.md"

[nav.footer]
Imprint = "imprint.md"
`)

	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if res.BasePath != "/cress" {
		t.Errorf("Result.BasePath = %q, want %q", res.BasePath, "/cress")
	}

	out := filepath.Join(root, config.OutputDir)
	// Output paths are unchanged: only the URLs written into the HTML move.
	if _, err := os.Stat(filepath.Join(out, "about.html")); err != nil {
		t.Errorf("output layout should not change under a base path: %v", err)
	}

	assertMarkers(t, "index.html", readFile(t, filepath.Join(out, "index.html")), []marker{
		{`href="/cress/style.css"`, "the theme's stylesheet under the base path"},
		{`href="/cress/about.html"`, "a nav link under the base path"},
		{`href="/cress/imprint.html"`, "a footer nav link under the base path"},
		{`href="/cress/"`, "the home link under the base path"},
		{`src="/cress/logo.svg"`, "the logo under the base path"},
		{`href="/cress/favicon.png"`, "the favicon under the base path"},
	})

	// A link an author wrote against the site root is rewritten too, which is the
	// half of this that the builder cannot see.
	assertMarkers(t, "guides/index.html", readFile(t, filepath.Join(out, "guides", "index.html")), []marker{
		{`href="/cress/guides/styleguide.html"`, "a content link under the base path"},
	})
	assertMarkers(t, "guides/writing.html", readFile(t, filepath.Join(out, "guides", "writing.html")), []marker{
		{`src="/cress/example_busy.webp"`, "a content image under the base path"},
	})

	// The synthesized 404 carries a link home, which is only useful if it points
	// inside the site rather than at the domain root.
	assertMarkers(t, build.NotFoundFile, readFile(t, filepath.Join(out, build.NotFoundFile)), []marker{
		{`href="/cress/"`, "the back-home link under the base path"},
	})

	// The active-page marker still resolves: it compares the page's URL against
	// the nav's, and both moved.
	assertMarkers(t, "about.html", readFile(t, filepath.Join(out, "about.html")), []marker{
		{`href="/cress/about.html" aria-current="page"`, "its own nav entry still marked active"},
	})
}

// The root case has to keep emitting exactly what it did before, since that is
// every site that does not set a path.
func TestBuild_domainRootIsUnprefixed_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if res.BasePath != "" {
		t.Errorf("Result.BasePath = %q, want empty for a site at a domain root", res.BasePath)
	}

	doc := readFile(t, filepath.Join(root, config.OutputDir, "index.html"))
	assertMarkers(t, "index.html", doc, []marker{
		{`href="/style.css"`, "the theme's stylesheet at the root"},
		{`href="/about.html"`, "a nav link at the root"},
		{`href="/"`, "the home link at the root"},
		{`src="/logo.svg"`, "the logo at the root"},
	})
	if strings.Contains(doc, "//") && strings.Contains(doc, `href="//`) {
		t.Errorf("an empty base path should not double any slash:\n%s", doc)
	}
}

// The metadata a page carries comes from three places: the page's own front
// matter, the site config behind it, and what the builder derives.
func TestBuild_metadata_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	writeSiteFile(t, filepath.Join(root, "cress.toml"), `
[site]
title = "My cozy site"
description = "A site-wide tagline."
base_url = "https://example.com"
language = "de"
`)
	// A page that overrides both keys, and one that inherits them.
	writeSiteFile(t, filepath.Join(root, "content", "about.md"),
		"---\ntitle: About\ndescription: What this page is about.\nlanguage: fr\n---\n\n# About\n")

	if _, err := build.Build(build.Options{Root: root}); err != nil {
		t.Fatalf("build: %v", err)
	}
	out := filepath.Join(root, config.OutputDir)

	assertMarkers(t, "about.html", readFile(t, filepath.Join(out, "about.html")), []marker{
		{`<html lang="fr">`, "the page's own language"},
		{`<meta name="description" content="What this page is about." />`, "the page's own description"},
		{`<link rel="canonical" href="https://example.com/about.html" />`, "the canonical link"},
		{"<title>About · My cozy site</title>", "the composed title"},
	})

	assertMarkers(t, "imprint.html", readFile(t, filepath.Join(out, "imprint.html")), []marker{
		{`<html lang="de">`, "the site language, inherited"},
		{`<meta name="description" content="A site-wide tagline." />`, "the site description, inherited"},
		{`<link rel="canonical" href="https://example.com/imprint.html" />`, "the canonical link"},
	})

	// The home page is the site, so its title does not name the site twice.
	assertMarkers(t, "index.html", readFile(t, filepath.Join(out, "index.html")), []marker{
		{"<title>My cozy site</title>", "the home title without a suffix"},
		{`<link rel="canonical" href="https://example.com/" />`, "the canonical link for the root"},
	})

	// A 404 stands in for every address that does not exist, so it names none of
	// them as canonical, and asks to stay out of search results.
	notFound := readFile(t, filepath.Join(out, build.NotFoundFile))
	assertMarkers(t, build.NotFoundFile, notFound, []marker{
		{`<meta name="robots" content="noindex" />`, "the noindex directive"},
		{`<html lang="de">`, "the site language"},
	})
	if strings.Contains(notFound, "canonical") {
		t.Errorf("the 404 page should carry no canonical link:\n%s", notFound)
	}
}

// Without a base_url there is no absolute form for a page, so the tag that needs
// one is left out rather than emitted empty.
func TestBuild_noCanonicalWithoutBaseURL_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\ndescription = \"D\"\n")

	if _, err := build.Build(build.Options{Root: root}); err != nil {
		t.Fatalf("build: %v", err)
	}

	doc := readFile(t, filepath.Join(root, config.OutputDir, "index.html"))
	if strings.Contains(doc, "canonical") {
		t.Errorf("no base_url means no canonical link:\n%s", doc)
	}
	assertMarkers(t, "index.html", doc, []marker{
		{`<meta name="description" content="D" />`, "the description, which needs no base_url"},
		{`<html lang="en">`, "the default language"},
	})
}

// A site whose links are rooted under a subdirectory has to describe itself with
// URLs that carry the same prefix, exactly once.
func TestBuild_canonicalUnderSubpath_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	writeSiteFile(t, filepath.Join(root, "cress.toml"),
		"[site]\ntitle = \"S\"\nbase_url = \"https://user.github.io/cress\"\n")

	if _, err := build.Build(build.Options{Root: root}); err != nil {
		t.Fatalf("build: %v", err)
	}
	out := filepath.Join(root, config.OutputDir)

	assertMarkers(t, "about.html", readFile(t, filepath.Join(out, "about.html")), []marker{
		{`<link rel="canonical" href="https://user.github.io/cress/about.html" />`, "the canonical link under the base path"},
	})
	assertMarkers(t, "index.html", readFile(t, filepath.Join(out, "index.html")), []marker{
		{`<link rel="canonical" href="https://user.github.io/cress/" />`, "the canonical link for the rooted home page"},
	})
}

// A site need not have an index of its own, and when it does not, no other page
// inherits the home page's treatment by accident.
func TestBuild_noHomePage_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\n")
	writeSiteFile(t, filepath.Join(root, "content", "about.md"), "---\ntitle: About\n---\n\nprose\n")

	if _, err := build.Build(build.Options{Root: root}); err != nil {
		t.Fatalf("build: %v", err)
	}

	assertMarkers(t, "about.html", readFile(t, filepath.Join(root, config.OutputDir, "about.html")), []marker{
		{"<title>About · S</title>", "a full title, since this page is not the home page"},
	})
}

func TestBuild_notFound_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("synthesized when neither content nor theme supplies one", func(t *testing.T) {
		root := t.TempDir()
		if _, err := scaffold.Create(root, false); err != nil {
			t.Fatalf("scaffold: %v", err)
		}

		res, err := build.Build(build.Options{Root: root})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		// A synthesized 404 is not a page the author wrote, so it must not move
		// the count; and it must never be reported as a file cress did not write.
		if res.Pages != scaffoldPages {
			t.Errorf("rendered %d pages, want %d: a synthesized 404 is not counted", res.Pages, scaffoldPages)
		}
		if len(res.Warnings) != 0 {
			t.Errorf("unexpected warnings: %v", res.Warnings)
		}

		doc := readFile(t, filepath.Join(root, config.OutputDir, build.NotFoundFile))
		assertMarkers(t, build.NotFoundFile, doc, []marker{
			{"<title>Page not found · My cozy site</title>", "the synthesized title"},
			{"Page not found</h1>", "the synthesized heading"},
			{`href="/about.html"`, "the theme's nav, so the 404 is not an orphan"},
			{`class="site-footer"`, "the theme's footer"},
			{`href="/"`, "a way back to the home page"},
		})
	})

	t.Run("content/404.md wins over the synthesized page", func(t *testing.T) {
		root := t.TempDir()
		if _, err := scaffold.Create(root, false); err != nil {
			t.Fatalf("scaffold: %v", err)
		}
		writeSiteFile(t, filepath.Join(root, "content", "404.md"),
			"---\ntitle: Lost\n---\n\n# Lost\n\nMINE_NOT_CRESSES\n")

		res, err := build.Build(build.Options{Root: root})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		// Authored this time, so it counts like any other page.
		if res.Pages != scaffoldPages+1 {
			t.Errorf("rendered %d pages, want %d with an authored 404", res.Pages, scaffoldPages+1)
		}

		doc := readFile(t, filepath.Join(root, config.OutputDir, build.NotFoundFile))
		if !strings.Contains(doc, "MINE_NOT_CRESSES") {
			t.Error("content/404.md must win; got the synthesized page instead")
		}
		if strings.Contains(doc, "There is nothing at this address") {
			t.Error("the synthesized body leaked into an authored 404")
		}
	})

	t.Run("a theme's 404 template wins over the synthesized page", func(t *testing.T) {
		root := t.TempDir()
		if _, err := scaffold.Create(root, false); err != nil {
			t.Fatalf("scaffold: %v", err)
		}
		templates := filepath.Join(root, "themes", "mine", "templates")
		writeSiteFile(t, filepath.Join(templates, "page.html"), `PAGE:{{ .Page.Title }}`)
		writeSiteFile(t, filepath.Join(templates, build.NotFoundFile),
			`THEME404:{{ .Page.Title }}:{{ .Site.Title }}`)
		writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\ntheme = \"mine\"\n")

		if _, err := build.Build(build.Options{Root: root}); err != nil {
			t.Fatalf("build: %v", err)
		}

		// The theme's 404 template receives the same data a page does.
		if got := readFile(t, filepath.Join(root, config.OutputDir, build.NotFoundFile)); got != "THEME404:Page not found:S" {
			t.Errorf("404.html = %q, want the theme's own 404 template to render it", got)
		}
	})

	t.Run("a theme without a 404 template still gets one", func(t *testing.T) {
		root := t.TempDir()
		if _, err := scaffold.Create(root, false); err != nil {
			t.Fatalf("scaffold: %v", err)
		}
		templates := filepath.Join(root, "themes", "bare", "templates")
		writeSiteFile(t, filepath.Join(templates, "page.html"), `BARE:{{ .Page.Title }}`)
		writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\ntheme = \"bare\"\n")

		if _, err := build.Build(build.Options{Root: root}); err != nil {
			t.Fatalf("build: %v", err)
		}

		// The point of tier 3: a theme written before cress had a 404 gets a
		// styled one through page.html rather than none at all.
		if got := readFile(t, filepath.Join(root, config.OutputDir, build.NotFoundFile)); got != "BARE:Page not found" {
			t.Errorf("404.html = %q, want it rendered through the theme's page.html", got)
		}
	})
}

func TestBuild_layout_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// A theme offering a layout beyond the required entry template. Each renders
	// a distinct marker, so which one ran is unambiguous.
	newSite := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		templates := filepath.Join(root, "themes", "mine", "templates")
		writeSiteFile(t, filepath.Join(templates, "page.html"), `PAGE:{{ .Page.Title }}`)
		writeSiteFile(t, filepath.Join(templates, "wide.html"), `WIDE:{{ .Page.Title }}`)
		writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\ntheme = \"mine\"\n")
		return root
	}

	t.Run("a page renders through the layout it names", func(t *testing.T) {
		root := newSite(t)
		writeSiteFile(t, filepath.Join(root, "content", "wide.md"), "---\ntitle: Wide\nlayout: wide\n---\n\nbody\n")

		res, err := build.Build(build.Options{Root: root})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		if len(res.Warnings) != 0 {
			t.Errorf("a layout the theme defines must not warn: %v", res.Warnings)
		}
		if got := readFile(t, filepath.Join(root, config.OutputDir, "wide.html")); got != "WIDE:Wide" {
			t.Errorf("wide.html = %q, want the named layout to render it", got)
		}
	})

	t.Run("a page naming no layout keeps the entry template", func(t *testing.T) {
		root := newSite(t)
		writeSiteFile(t, filepath.Join(root, "content", "plain.md"), "---\ntitle: Plain\n---\n\nbody\n")

		res, err := build.Build(build.Options{Root: root})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		if len(res.Warnings) != 0 {
			t.Errorf("unexpected warnings: %v", res.Warnings)
		}
		if got := readFile(t, filepath.Join(root, config.OutputDir, "plain.html")); got != "PAGE:Plain" {
			t.Errorf("plain.html = %q, want the entry template to render it", got)
		}
	})

	// Naming a layout no theme defines is the case a theme switch produces, so it
	// must still yield a page. It falls back and says so, the way an unresolvable
	// nav entry does, rather than failing the build or going quiet.
	t.Run("an unknown layout falls back and warns", func(t *testing.T) {
		root := newSite(t)
		writeSiteFile(t, filepath.Join(root, "content", "odd.md"), "---\ntitle: Odd\nlayout: nope\n---\n\nbody\n")

		res, err := build.Build(build.Options{Root: root})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		if got := readFile(t, filepath.Join(root, config.OutputDir, "odd.html")); got != "PAGE:Odd" {
			t.Errorf("odd.html = %q, want the entry template as the fallback", got)
		}
		if len(res.Warnings) != 1 {
			t.Fatalf("warnings = %v, want exactly one naming the missing layout", res.Warnings)
		}
		for _, want := range []string{"odd.md", `"nope"`, `"mine"`} {
			if !strings.Contains(res.Warnings[0], want) {
				t.Errorf("warning %q does not mention %s", res.Warnings[0], want)
			}
		}
	})

	// A layout name is a template lookup, never a file path, so a name that looks
	// like one reaches nothing outside the theme's own templates.
	t.Run("a traversing layout name is an ordinary miss", func(t *testing.T) {
		root := newSite(t)
		writeSiteFile(t, filepath.Join(root, "content", "evil.md"),
			"---\ntitle: Evil\nlayout: ../../../../etc/passwd\n---\n\nbody\n")

		res, err := build.Build(build.Options{Root: root})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		if got := readFile(t, filepath.Join(root, config.OutputDir, "evil.html")); got != "PAGE:Evil" {
			t.Errorf("evil.html = %q, want the entry template as the fallback", got)
		}
		if len(res.Warnings) != 1 {
			t.Errorf("warnings = %v, want exactly one", res.Warnings)
		}
	})

	// A partial is not a layout. It renders a fragment, so a page reaching one
	// would emit a document fragment as a whole page, and the name it would be
	// reached by is exactly the sort a theme author picks for a partial.
	t.Run("a partial cannot be reached as a layout", func(t *testing.T) {
		root := newSite(t)
		partials := filepath.Join(root, "themes", "mine", "templates", "partials")
		writeSiteFile(t, filepath.Join(partials, "banner.html"), `{{ define "banner" }}BANNER{{ end }}`)
		writeSiteFile(t, filepath.Join(root, "content", "try.md"), "---\ntitle: Try\nlayout: banner\n---\n\nbody\n")

		res, err := build.Build(build.Options{Root: root})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		if got := readFile(t, filepath.Join(root, config.OutputDir, "try.html")); got != "PAGE:Try" {
			t.Errorf("try.html = %q, want the entry template rather than the partial", got)
		}
		if len(res.Warnings) != 1 {
			t.Fatalf("warnings = %v, want exactly one naming the partial as missing", res.Warnings)
		}
	})
}

// The built-in theme ships the layout its scaffold names, so `cress init`
// followed by `cress build` demonstrates the feature without a warning.
func TestBuild_builtinLandingLayout_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("the scaffold must build clean: %v", res.Warnings)
	}
	out := filepath.Join(root, config.OutputDir)

	home := readFile(t, filepath.Join(out, "index.html"))
	assertMarkers(t, "index.html", home, []marker{
		{`class="landing-top"`, "the full-screen opening panel"},
		{`class="hero-title"`, "the site name set as display type"},
		// The landing layout differs below the header, not in place of it: the
		// site's usual navigation is still the way off this page.
		{`class="site-header"`, "the standard header bar, kept"},
		{`href="/about.html"`, "a nav link, so a landing page can still be left"},
		{`class="site-footer"`, "the shared footer"},
		// The arrow says the first screen is not the whole page. It is a link, so
		// it needs somewhere to land and a name to be announced by.
		{`href="#content"`, "the arrow down to the content"},
		{`aria-label="Skip to the content"`, "the arrow's accessible name"},
		{`id="content"`, "the target the arrow points at"},
		// The scaffold sets a logo, so the panel is marked with it. Only the
		// template knows the URL, so it reaches the stylesheet as a property.
		{`class="hero hero-logo"`, "the panel marked with the site's logo"},
		{`--site-logo: url('/logo.svg')`, "the logo URL handed to the stylesheet"},
	})
	// The page's own H1 comes from its Markdown. A landing page that also set one
	// in the panel would have two, so the panel sets none.
	if got := strings.Count(home, "<h1"); got != 1 {
		t.Errorf("index.html has %d <h1> elements, want exactly 1", got)
	}

	// Every other scaffold page names no layout and is untouched by this.
	about := readFile(t, filepath.Join(out, "about.html"))
	assertMarkers(t, "about.html", about, []marker{
		{`class="site-header"`, "the standard header bar"},
	})
	if strings.Contains(about, `class="hero"`) {
		t.Error("a page naming no layout must not get the landing panel")
	}
}

// An arrow down to the content is a promise that there is some. A landing page
// carrying nothing below the panel gets no arrow rather than one pointing at an
// empty main.
func TestBuild_landingWithoutBody_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\n")
	writeSiteFile(t, filepath.Join(root, "content", "index.md"), "---\ntitle: Home\nlayout: landing\n---\n")

	res, err := build.Build(build.Options{Root: root})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(res.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", res.Warnings)
	}

	home := readFile(t, filepath.Join(root, config.OutputDir, "index.html"))
	assertMarkers(t, "index.html", home, []marker{
		{`class="hero-title"`, "the panel, which does not depend on there being a body"},
	})
	if strings.Contains(home, "hero-more") {
		t.Errorf("an empty body must not get an arrow pointing at it:\n%s", home)
	}
}

// The panel's highlight is drawn from the site's logo, so a site without one
// falls back to a wash the stylesheet mixes from the accent. Either way the
// panel has something behind its type.
func TestBuild_landingWithoutLogo_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	root := t.TempDir()
	writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\n")
	writeSiteFile(t, filepath.Join(root, "content", "index.md"), "---\ntitle: Home\nlayout: landing\n---\n\n# Home\n\nbody\n")

	if _, err := build.Build(build.Options{Root: root}); err != nil {
		t.Fatalf("build: %v", err)
	}

	home := readFile(t, filepath.Join(root, config.OutputDir, "index.html"))
	assertMarkers(t, "index.html", home, []marker{
		{`class="hero hero-wash"`, "the accent wash, standing in for an absent logo"},
	})
	// Naming a property with no URL behind it would leave the panel asking the
	// browser for `url()` of nothing.
	if strings.Contains(home, "--site-logo") {
		t.Errorf("a site with no logo must not emit the logo property:\n%s", home)
	}
}

// A theme's own metadata reaches the person building the site, and nothing in
// it can stop them: theme.toml is optional provenance, so every problem it
// carries is a warning. The one exception is a file that is not TOML at all,
// where there is nothing to recover and nothing to warn about.
func TestBuild_themeMeta_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	newSite := func(t *testing.T, meta string) string {
		t.Helper()
		root := t.TempDir()
		dir := filepath.Join(root, "themes", "mine")
		writeSiteFile(t, filepath.Join(dir, "templates", "page.html"), `PAGE:{{ .Page.Title }}`)
		writeSiteFile(t, filepath.Join(dir, theme.MetaFile), meta)
		writeSiteFile(t, filepath.Join(root, "cress.toml"), "[site]\ntitle = \"S\"\ntheme = \"mine\"\n")
		writeSiteFile(t, filepath.Join(root, "content", "index.md"), "---\ntitle: Home\n---\n\nbody\n")
		return root
	}

	t.Run("a theme declaring the running contract does not warn", func(t *testing.T) {
		root := newSite(t, "name = \"Mine\"\ncress = \"1.1\"\n")
		stubVersion(t, "1.1.0")

		res, err := build.Build(build.Options{Root: root})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		if len(res.Warnings) != 0 {
			t.Errorf("unexpected warnings: %v", res.Warnings)
		}
	})

	// The contract a theme names is checked against the cress running it, in
	// both directions, and either way the site still builds.
	t.Run("a mismatched contract warns and still builds", func(t *testing.T) {
		for _, tc := range []struct {
			name      string
			declared  string
			runningIs string
			want      string
		}{
			{
				name:     "cress is older than the theme asks for",
				declared: "9.0", runningIs: "1.1.0",
				want: "may read fields this version does not provide",
			},
			{
				name: "cress is a major version newer", declared: "1.1", runningIs: "2.0.0",
				want: "not promised across a major version",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				root := newSite(t, "cress = \""+tc.declared+"\"\n")
				stubVersion(t, tc.runningIs)

				res, err := build.Build(build.Options{Root: root})
				if err != nil {
					t.Fatalf("build: %v", err)
				}
				assertWarns(t, res.Warnings, tc.want, `theme "mine"`)
				if got := readFile(t, filepath.Join(root, config.OutputDir, "index.html")); got != "PAGE:Home" {
					t.Errorf("index.html = %q, want the page to have been built anyway", got)
				}
			})
		}
	})

	// An unknown key is also what a theme looks like when it declares something
	// a later cress added, so it must never be worse than a warning.
	t.Run("an unknown key warns and still builds", func(t *testing.T) {
		root := newSite(t, "name = \"Mine\"\nscreenshot = \"preview.png\"\n")

		res, err := build.Build(build.Options{Root: root})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		assertWarns(t, res.Warnings, "screenshot", `theme "mine"`)
		if got := readFile(t, filepath.Join(root, config.OutputDir, "index.html")); got != "PAGE:Home" {
			t.Errorf("index.html = %q, want the page to have been built anyway", got)
		}
	})

	t.Run("a theme.toml that is not TOML fails the build", func(t *testing.T) {
		root := newSite(t, "name = \nnot toml")

		if _, err := build.Build(build.Options{Root: root}); err == nil {
			t.Fatal("expected an error for a theme.toml that does not parse")
		}
	})
}

// stubVersion sets the build metadata a theme's declared contract is compared
// against, restoring it afterwards. The linker owns this variable in a release,
// and an untagged local build leaves it unparseable, so a test that wants the
// comparison to happen at all has to say what is running.
func stubVersion(t *testing.T, v string) {
	t.Helper()
	previous := version.Version
	t.Cleanup(func() { version.Version = previous })
	version.Version = v
}

func TestBuild_refusesSiteRootOutput(t *testing.T) {
	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	if _, err := build.Build(build.Options{Root: root, Output: "."}); err == nil {
		t.Fatal("expected an error when output is the site root")
	}
}

// marker is one substring a rendered document must contain, with a
// human-readable name for the failure message.
type marker struct{ substr, desc string }

// assertMarkers checks every marker against doc, reporting each miss
// separately so one failure does not mask the others. The document is dumped
// once at the end when anything is missing, since the site it was built from
// lives in a t.TempDir() that is gone by the time the failure is read.
func assertMarkers(t *testing.T, name, doc string, markers []marker) {
	t.Helper()
	missing := false
	for _, m := range markers {
		if !strings.Contains(doc, m.substr) {
			t.Errorf("%s is missing %s (%q)", name, m.desc, m.substr)
			missing = true
		}
	}
	if missing {
		t.Logf("%s was:\n%s", name, doc)
	}
}

// assertWarns checks that each substring is named by at least one warning,
// reporting every miss separately and dumping the full list once, since a
// missing warning reads best next to the ones that did fire.
func assertWarns(t *testing.T, warnings []string, substrs ...string) {
	t.Helper()
	missing := false
	for _, want := range substrs {
		if !slices.ContainsFunc(warnings, func(w string) bool { return strings.Contains(w, want) }) {
			t.Errorf("no warning mentions %q", want)
			missing = true
		}
	}
	if missing {
		t.Logf("warnings were: %v", warnings)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path) // #nosec G304 -- test-controlled path
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

func writeSiteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// A symlink under static/ is not copied. Nothing narrows what static/ holds, so
// following one would publish any readable file under whatever name the link
// was given, which is worse than the same trick in content/.
func TestBuild_staticSkipsSymlinks_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("SECRET"), 0o600); err != nil {
		t.Fatalf("writing secret: %v", err)
	}

	root := t.TempDir()
	if _, err := scaffold.Create(root, false); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	writeSiteFile(t, filepath.Join(root, "static", "real.txt"), "real\n")
	if err := os.Symlink(outside, filepath.Join(root, "static", "leak.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if _, err := build.Build(build.Options{Root: root}); err != nil {
		t.Fatalf("build: %v", err)
	}

	out := filepath.Join(root, config.OutputDir)
	if _, err := os.Stat(filepath.Join(out, "real.txt")); err != nil {
		t.Errorf("a regular static file should still be copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "leak.txt")); !os.IsNotExist(err) {
		t.Errorf("a symlinked static file should not be published, got err=%v", err)
	}
}
