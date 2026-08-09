package theme

import (
	"strings"
	"testing"
	"testing/fstest"
)

// validateFS validates a theme assembled in memory, so the checks are exercised
// without a directory to lay out.
func validateFS(t *testing.T, files map[string]string) *Report {
	t.Helper()

	fsys := fstest.MapFS{}
	for path, body := range files {
		fsys[path] = &fstest.MapFile{Data: []byte(body)}
	}
	collected, findings := collectTemplates(fsys)
	report := &Report{Name: "test"}
	report.Findings = append(report.Findings, findings...)
	report.Findings = append(report.Findings, checkEntry(collected)...)
	report.Findings = append(report.Findings, checkDuplicateDefines(collected)...)
	report.Findings = append(report.Findings, checkEmptyLayouts(collected)...)
	report.Findings = append(report.Findings, checkAssetLinks(collected)...)
	report.Findings = append(report.Findings, checkFields(collected)...)
	sortFindings(report.Findings)
	return report
}

// assertOne checks that exactly one finding came back, in the named file and
// saying want.
func assertOne(t *testing.T, report *Report, file, want string) {
	t.Helper()
	if len(report.Findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(report.Findings), report.Findings)
	}
	if report.Findings[0].File != file {
		t.Errorf("File = %q, want %q", report.Findings[0].File, file)
	}
	if !strings.Contains(report.Findings[0].Message, want) {
		t.Errorf("message = %q, want it to contain %q", report.Findings[0].Message, want)
	}
}

func TestValidateAcceptsAMinimalTheme(t *testing.T) {
	report := validateFS(t, map[string]string{
		"templates/page.html": `<html lang="{{ .Page.Language }}">{{ .Page.HTML }}</html>`,
	})
	if !report.OK() {
		t.Errorf("a correct theme reported: %v", report.Findings)
	}
}

func TestValidateEntryTemplate(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		report := validateFS(t, map[string]string{"templates/landing.html": `hello`})
		assertOne(t, report, "", "no templates/page.html")
	})

	// A partials directory is not where the entry template lives, so a theme
	// with one there still has none.
	t.Run("not satisfied by a partial of the same name", func(t *testing.T) {
		report := validateFS(t, map[string]string{
			"templates/landing.html":       `hello`,
			"templates/partials/page.html": `{{ define "x" }}y{{ end }}`,
		})
		assertOne(t, report, "", "no templates/page.html")
	})
}

// A page.html with a syntax error is one problem, not two. Reporting it as
// missing as well would send the reader looking for a file that is right there.
func TestValidateParseErrors(t *testing.T) {
	report := validateFS(t, map[string]string{
		"templates/page.html": "{{ if .Page.IsHome }}unclosed",
	})
	if len(report.Findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(report.Findings), report.Findings)
	}
	if report.Findings[0].File != "templates/page.html" {
		t.Errorf("File = %q, want the file that does not parse", report.Findings[0].File)
	}
	// The message says what is wrong without repeating the package name that
	// text/template prefixes to it.
	if strings.HasPrefix(report.Findings[0].Message, "template: ") {
		t.Errorf("message = %q, want the template: prefix stripped", report.Findings[0].Message)
	}
}

func TestValidateDuplicateDefines(t *testing.T) {
	t.Run("across two partials", func(t *testing.T) {
		report := validateFS(t, map[string]string{
			"templates/page.html":          `{{ template "head" . }}`,
			"templates/partials/a.html":    `{{ define "head" }}a{{ end }}`,
			"templates/partials/head.html": `{{ define "head" }}b{{ end }}`,
		})
		assertOne(t, report, "templates/partials/head.html", `defines "head", which templates/partials/a.html also defines`)
	})

	t.Run("between a layout and a partial", func(t *testing.T) {
		report := validateFS(t, map[string]string{
			"templates/page.html":          `{{ define "aside" }}a{{ end }}body`,
			"templates/partials/side.html": `{{ define "aside" }}b{{ end }}`,
		})
		assertOne(t, report, "templates/partials/side.html", `defines "aside"`)
	})

	// The same name in two files is only a collision when both are the theme's.
	// One file defining one name is not.
	t.Run("no false positive on distinct names", func(t *testing.T) {
		report := validateFS(t, map[string]string{
			"templates/page.html":          `{{ template "head" . }}{{ template "foot" . }}`,
			"templates/partials/head.html": `{{ define "head" }}a{{ end }}`,
			"templates/partials/foot.html": `{{ define "foot" }}b{{ end }}`,
		})
		if !report.OK() {
			t.Errorf("reported distinct define names: %v", report.Findings)
		}
	})
}

func TestValidateEmptyLayouts(t *testing.T) {
	t.Run("a layout holding only defines", func(t *testing.T) {
		report := validateFS(t, map[string]string{
			"templates/page.html":  `body`,
			"templates/aside.html": "\n{{ define \"aside\" }}x{{ end }}\n",
		})
		assertOne(t, report, "templates/aside.html", "renders nothing outside its {{ define }} blocks")
	})

	t.Run("a layout that only calls a partial is not empty", func(t *testing.T) {
		report := validateFS(t, map[string]string{
			"templates/page.html":          `body`,
			"templates/wide.html":          `{{ template "head" . }}`,
			"templates/partials/head.html": `{{ define "head" }}x{{ end }}`,
		})
		if !report.OK() {
			t.Errorf("reported a layout that renders a partial: %v", report.Findings)
		}
	})

	// A partial is supposed to be nothing but defines, so the same shape one
	// directory down is correct rather than a finding.
	t.Run("a partial holding only defines is fine", func(t *testing.T) {
		report := validateFS(t, map[string]string{
			"templates/page.html":          `{{ template "head" . }}`,
			"templates/partials/head.html": `{{ define "head" }}x{{ end }}`,
		})
		if !report.OK() {
			t.Errorf("reported a normal partial: %v", report.Findings)
		}
	})
}

func TestValidateAssetLinks(t *testing.T) {
	rooted := "is not rooted under {{ .Site.BasePath }}"

	t.Run("reports", func(t *testing.T) {
		for _, tc := range []struct{ name, body, want string }{
			{"a stylesheet", `<link rel="stylesheet" href="/style.css">`, `href="/style.css"`},
			{"an image", `<img src="/logo.svg">`, `src="/logo.svg"`},
			{"a link to the home page", `<a href="/">home</a>`, `href="/"`},
			{"single quotes", `<img src='/logo.svg'>`, `src='/logo.svg'`},
			{"a value that starts rooted", `<a href="/{{ .Page.URL }}">x</a>`, `href="/{{ .Page.URL }}"`},
			{"whitespace around the equals", `<img src = "/logo.svg">`, `src="/logo.svg"`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				report := validateFS(t, map[string]string{"templates/page.html": tc.body})
				assertOne(t, report, "templates/page.html", rooted)
			})
		}
	})

	t.Run("stays silent", func(t *testing.T) {
		for _, tc := range []struct{ name, body string }{
			{"a rooted asset link", `<link rel="stylesheet" href="{{ .Site.BasePath }}/style.css">`},
			{"a URL cress already rooted", `<a href="{{ .Page.URL }}">x</a>`},
			{"an absolute URL", `<a href="https://example.com/x">x</a>`},
			{"a protocol-relative URL", `<img src="//cdn.example.com/logo.svg">`},
			{"a relative link", `<a href="about.html">x</a>`},
			{"a fragment", `<a href="#main">skip</a>`},
			{"a mailto", `<a href="mailto:x@example.com">x</a>`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				report := validateFS(t, map[string]string{"templates/page.html": tc.body})
				if !report.OK() {
					t.Errorf("reported a correct link: %v", report.Findings)
				}
			})
		}
	})

	// Two of the same link are two findings on their own lines, rather than one
	// line reported twice.
	t.Run("reports each occurrence on its own line", func(t *testing.T) {
		report := validateFS(t, map[string]string{
			"templates/page.html": "<img src=\"/logo.svg\">\n<b>x</b>\n<img src=\"/logo.svg\">\n",
		})
		if len(report.Findings) != 2 {
			t.Fatalf("got %d findings, want 2: %v", len(report.Findings), report.Findings)
		}
		if report.Findings[0].Line != 1 || report.Findings[1].Line != 3 {
			t.Errorf("lines = %d and %d, want 1 and 3", report.Findings[0].Line, report.Findings[1].Line)
		}
	})
}

// A report is read top to bottom, so it is ordered the way somebody would walk
// the theme rather than in whatever order the checks happened to run.
func TestValidateOrdersFindings(t *testing.T) {
	report := validateFS(t, map[string]string{
		"templates/page.html":          "{{ .Page.Titel }}\n{{ template \"head\" . }}\n<img src=\"/a.png\">\n",
		"templates/partials/head.html": "{{ define \"head\" }}\n{{ .Site.Titel }}\n{{ end }}",
	})
	if len(report.Findings) != 3 {
		t.Fatalf("got %d findings, want 3: %v", len(report.Findings), report.Findings)
	}
	want := []struct {
		file string
		line int
	}{
		{"templates/page.html", 1},
		{"templates/page.html", 3},
		{"templates/partials/head.html", 2},
	}
	for i, w := range want {
		if report.Findings[i].File != w.file || report.Findings[i].Line != w.line {
			t.Errorf("finding %d = %s:%d, want %s:%d", i, report.Findings[i].File, report.Findings[i].Line, w.file, w.line)
		}
	}
}

func TestLineAt(t *testing.T) {
	const text = "one\ntwo\nthree"
	for _, tc := range []struct {
		offset, want int
	}{
		{0, 1}, {3, 1}, {4, 2}, {8, 3}, {len(text), 3}, {len(text) + 99, 3},
	} {
		if got := lineAt(text, tc.offset); got != tc.want {
			t.Errorf("lineAt(offset %d) = %d, want %d", tc.offset, got, tc.want)
		}
	}
}
