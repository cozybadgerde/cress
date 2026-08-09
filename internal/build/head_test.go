package build

import (
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/config"
)

// siteWith builds the minimal site config absoluteURL reads.
func siteWith(baseURL string) config.Site {
	return config.Site{BaseURL: baseURL}
}

func TestRenderHead(t *testing.T) {
	tests := []struct {
		name string
		data headData
		want []string
		omit []string
	}{
		{
			name: "everything a normal page has",
			data: headData{Description: "A cozy site", Canonical: "https://example.com/about.html"},
			want: []string{
				`<meta name="description" content="A cozy site" />`,
				`<link rel="canonical" href="https://example.com/about.html" />`,
			},
			omit: []string{"robots"},
		},
		{
			// Without a base_url there is no absolute URL to point at, so the tag
			// that needs one is left out rather than emitted empty.
			name: "no canonical without an absolute URL",
			data: headData{Description: "A cozy site"},
			want: []string{`<meta name="description"`},
			omit: []string{"canonical"},
		},
		{
			name: "no description when neither page nor site has one",
			data: headData{Canonical: "https://example.com/"},
			want: []string{`<link rel="canonical"`},
			omit: []string{"description"},
		},
		{
			name: "the synthesized 404 is kept out of search results",
			data: headData{Description: "A cozy site", NoIndex: true},
			want: []string{`<meta name="robots" content="noindex" />`},
			omit: []string{"canonical"},
		},
		{
			name: "nothing to say renders nothing",
			data: headData{},
			omit: []string{"<meta", "<link"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := renderHead(tc.data)
			if err != nil {
				t.Fatalf("renderHead: %v", err)
			}
			assertHead(t, string(got), tc.want, tc.omit)
			if strings.HasSuffix(string(got), "\n") {
				t.Errorf("head should carry no trailing newline, got %q", got)
			}
		})
	}
}

// assertHead checks a rendered head for the elements a case expects and against
// the ones it expects to be absent.
//
// Absence carries most of the meaning here: what this template does is leave
// out the tag it has no value for, so a case that could only assert presence
// would not be testing the behaviour at all.
//
// It is a near twin of assertDoc in the integration tests, and stays separate
// because those are an external test package. Sharing one would mean exporting
// a test helper from the package under test.
func assertHead(t *testing.T, head string, want, omit []string) {
	t.Helper()
	wrong := false
	for _, text := range want {
		if !strings.Contains(head, text) {
			t.Errorf("head is missing %q", text)
			wrong = true
		}
	}
	for _, text := range omit {
		if strings.Contains(head, text) {
			t.Errorf("head should not contain %q", text)
			wrong = true
		}
	}
	if wrong {
		t.Logf("head was:\n%s", head)
	}
}

// The description is the author's prose landing inside an attribute. Escaping is
// what keeps a quotation mark in a sentence from closing that attribute and
// spilling the rest of it into the markup.
func TestRenderHeadEscapesDescription(t *testing.T) {
	got, err := renderHead(headData{
		Description: `He said "hello" & <script>alert(1)</script>`,
		Canonical:   "https://example.com/a.html?x=1&y=2",
	})
	if err != nil {
		t.Fatalf("renderHead: %v", err)
	}

	for _, raw := range []string{`"hello"`, "<script>"} {
		if strings.Contains(string(got), raw) {
			t.Errorf("head contains unescaped %q\ngot: %s", raw, got)
		}
	}
	if !strings.Contains(string(got), "&amp;") {
		t.Errorf("head should escape the ampersand\ngot: %s", got)
	}
	// One opening and one closing quote per attribute, so nothing broke out.
	if n := strings.Count(string(got), `content="`); n != 1 {
		t.Errorf("expected exactly one content attribute, got %d\n%s", n, got)
	}
}

func TestFallback(t *testing.T) {
	tests := []struct {
		name, value, siteWide, want string
	}{
		{"the page wins", "page", "site", "page"},
		{"the site stands in", "", "site", "site"},
		{"neither leaves nothing", "", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := fallback(tc.value, tc.siteWide); got != tc.want {
				t.Errorf("fallback(%q, %q) = %q, want %q", tc.value, tc.siteWide, got, tc.want)
			}
		})
	}
}

func TestAbsoluteURL(t *testing.T) {
	tests := []struct {
		name, baseURL, basePath, pageURL, want string
	}{
		{"domain root", "https://example.com", "", "/about.html", "https://example.com/about.html"},
		{"home page", "https://example.com", "", "/", "https://example.com/"},
		// The page URL already carries the base path, and so does BaseURL, so
		// joining them naively would name the subdirectory twice.
		{
			name: "subpath is not doubled", baseURL: "https://user.github.io/cress", basePath: "/cress",
			pageURL: "/cress/about.html", want: "https://user.github.io/cress/about.html",
		},
		{
			name: "subpath home", baseURL: "https://user.github.io/cress", basePath: "/cress",
			pageURL: "/cress/", want: "https://user.github.io/cress/",
		},
		{"no base_url, no absolute form", "", "", "/about.html", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := &pageWriter{
				site:     siteWith(tc.baseURL),
				basePath: tc.basePath,
			}
			if got := w.absoluteURL(tc.pageURL); got != tc.want {
				t.Errorf("absoluteURL(%q) = %q, want %q", tc.pageURL, got, tc.want)
			}
		})
	}
}
