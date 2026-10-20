package content_test

import (
	"testing"

	"github.com/cozybadgerde/cress/internal/content"
)

func TestResolveURL(t *testing.T) {
	tests := []struct{ source, prefix, want string }{
		{"/about/", "/", "/about/"},
		{"/about/", "/company/", "/company/about/"},
		{"/", "/company/", "/company/"},
		{"/about/?q=1#team", "/company/", "/company/about/?q=1#team"},
		{"/a%20b/", "/company/", "/company/a%20b/"},
		{"/company/about/", "/company/", "/company/company/about/"},
		{"https://example.net/about/?q=1#team", "/company/", "https://example.net/about/?q=1#team"},
		{"http://localhost:8080/", "/", "http://localhost:8080/"},
		{"https://[::1]:8443/image.jpg", "/", "https://[::1]:8443/image.jpg"},
		{"HTTPS://EXAMPLE.NET/about/", "/company/", "HTTPS://EXAMPLE.NET/about/"},
	}
	for _, test := range tests {
		t.Run(test.source+" under "+test.prefix, func(t *testing.T) {
			got, err := content.ResolveURL(test.source, test.prefix)
			if err != nil || got != test.want {
				t.Fatalf("resolved = %q, error = %v, want %q", got, err, test.want)
			}
		})
	}
}

func TestRejectInvalidURLs(t *testing.T) {
	for _, source := range []string{
		"", "about/", "#team", "?query=1", "//example.org/", "///example.org/",
		"javascript:alert(1)", "data:image/svg+xml,test", "mailto:hello@example.org",
		"https:/example.org", "https://", "https:opaque", "https://example.org:",
		"https://example.org:0/", "https://example.org:65536/", "https://example.org:abc/",
		"https://user@example.org/", "https://[not-ip]/", "https://[127.0.0.1]/", "https://[::1", "https://::1/",
		"/a b/", "/a\tb/", "/a\nb/", "/a\u00a0b/", `/a\b/`, "/../outside/", "/a/./b/",
		"/%2e%2e/outside/", "/%2Fexample.org/", "/a%5cb/", "/a%00b/", "/invalid%zz/",
	} {
		t.Run(source, func(t *testing.T) {
			if got, err := content.ResolveURL(source, "/company/"); err == nil {
				t.Fatalf("invalid URL accepted as %q", got)
			}
		})
	}
}
