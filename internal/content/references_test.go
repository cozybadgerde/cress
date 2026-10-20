package content

import (
	"strings"
	"testing"
)

func TestResolveSourceReferences(t *testing.T) {
	references := references{
		prefix: "/company/", source: "content/blog/post.md",
		files: map[string]reference{
			"content/about.md":                  {url: "/company/about/", page: true},
			"content/about/index.md":            {url: "/company/about/", page: true},
			"content/blog/post.assets/team.jpg": {url: "/company/blog/post/team.jpg"},
			"content/blog/post.assets/index.md": {url: "/company/blog/post/index.md"},
			"content/Our team.md":               {url: "/company/Our%20team/", page: true},
			"static/images/logo.svg":            {url: "/company/images/logo.svg"},
		},
	}
	for _, test := range []struct{ source, want string }{
		{"../about.md", "/company/about/"},
		{"../about/index.md?view=full#team", "/company/about/?view=full#team"},
		{"../Our%20team.md", "/company/Our%20team/"},
		{"post.assets/team.jpg?size=2#image", "/company/blog/post/team.jpg?size=2#image"},
		{"post.assets/index.md", "/company/blog/post/index.md"},
		{"../../static/images/logo.svg", "/company/images/logo.svg"},
		{"/about/", "/company/about/"},
		{"/company/about/", "/company/company/about/"},
		{"?q=1", "?q=1"},
		{"#section", "#section"},
		{"https://example.org/page.md", "https://example.org/page.md"},
		{"mailto:hello@example.org", "mailto:hello@example.org"},
		{"//example.org/image.jpg", "//example.org/image.jpg"},
	} {
		t.Run(test.source, func(t *testing.T) {
			resolved, err := references.resolve(test.source)
			if err != nil || resolved != test.want {
				t.Fatalf("resolved = %q, error = %v; want %q", resolved, err, test.want)
			}
		})
	}
	for _, source := range []string{
		"missing.md", "../draft.md", "post.assets/missing.jpg", "../../../outside.jpg",
		"../private.jpg", "post.assets/%5Csecret.jpg", "%00", "broken%escape",
		"/../outside", "/%5Csecret",
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := references.resolve(source); err == nil {
				t.Fatal("invalid reference accepted")
			}
		})
	}
	if _, err := references.resolveImage("../about.md"); err == nil || !strings.Contains(err.Error(), "asset") {
		t.Fatalf("page accepted as image: %v", err)
	}
	if _, err := references.resolveImage("post.assets/team.jpg"); err != nil {
		t.Fatal(err)
	}
	if resolved, err := references.resolveImage("#fragment"); err != nil || resolved != "#fragment" {
		t.Fatalf("fragment-only image changed: %q, %v", resolved, err)
	}
}
