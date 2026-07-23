package content

import (
	"testing"
)

func TestSplitFrontMatter(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		wantTitle any
		wantBody  string
	}{
		{
			name:      "no front matter",
			src:       "# Hello\n\nbody\n",
			wantTitle: nil,
			wantBody:  "# Hello\n\nbody\n",
		},
		{
			name:      "with front matter",
			src:       "---\ntitle: Hi\n---\n# Hello\n",
			wantTitle: "Hi",
			wantBody:  "# Hello\n",
		},
		{
			name:      "empty front matter block",
			src:       "---\n---\nbody\n",
			wantTitle: nil,
			wantBody:  "body\n",
		},
		{
			name:      "unterminated block is not front matter",
			src:       "---\ntitle: Hi\nbody without close\n",
			wantTitle: nil,
			wantBody:  "---\ntitle: Hi\nbody without close\n",
		},
		{
			name:      "leading dashes that are not a fence",
			src:       "text --- more\n",
			wantTitle: nil,
			wantBody:  "text --- more\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meta, body, err := splitFrontMatter([]byte(tc.src))
			if err != nil {
				t.Fatalf("splitFrontMatter: %v", err)
			}
			if got := meta["title"]; got != tc.wantTitle {
				t.Errorf("title = %v, want %v", got, tc.wantTitle)
			}
			if string(body) != tc.wantBody {
				t.Errorf("body = %q, want %q", body, tc.wantBody)
			}
		})
	}
}

func TestSplitFrontMatterInvalidYAML(t *testing.T) {
	_, _, err := splitFrontMatter([]byte("---\ntitle: [unclosed\n---\nbody\n"))
	if err == nil {
		t.Fatal("expected an error for malformed YAML front matter")
	}
}

func TestURLFor(t *testing.T) {
	tests := []struct {
		out  string
		want string
	}{
		{"index.html", "/"},
		{"about.html", "/about.html"},
		{"guide/setup.html", "/guide/setup.html"},
		{"guide/index.html", "/guide/"},
	}
	for _, tc := range tests {
		t.Run(tc.out, func(t *testing.T) {
			if got := urlFor(tc.out); got != tc.want {
				t.Errorf("urlFor(%q) = %q, want %q", tc.out, got, tc.want)
			}
		})
	}
}

func TestTitleFor(t *testing.T) {
	tests := []struct {
		name string
		meta map[string]any
		body string
		rel  string
		want string
	}{
		{
			name: "front matter title wins",
			meta: map[string]any{"title": "From Meta"},
			body: "# From Heading\n",
			rel:  "page.md",
			want: "From Meta",
		},
		{
			name: "first heading when no title",
			meta: map[string]any{},
			body: "intro\n\n# From Heading\n",
			rel:  "page.md",
			want: "From Heading",
		},
		{
			name: "file name when nothing else",
			meta: map[string]any{},
			body: "just prose\n",
			rel:  "guide/setup.md",
			want: "setup",
		},
		{
			name: "blank front matter title falls through",
			meta: map[string]any{"title": "   "},
			body: "# Real\n",
			rel:  "page.md",
			want: "Real",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := titleFor(tc.meta, []byte(tc.body), tc.rel); got != tc.want {
				t.Errorf("titleFor = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFirstHeading(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"plain h1", "# Title\nbody\n", "Title"},
		{"not the first line", "intro\n# Title\n", "Title"},
		{"ignores fenced hash", "```\n# not a heading\n```\n# Real\n", "Real"},
		{"h2 is not h1", "## Sub\n", ""},
		{"none", "just text\n", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := firstHeading([]byte(tc.body)); got != tc.want {
				t.Errorf("firstHeading = %q, want %q", got, tc.want)
			}
		})
	}
}
