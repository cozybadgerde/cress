package preview

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestPublishedFilesExcludeMarkersAndListings(t *testing.T) {
	files := publishedFS{source: fstest.MapFS{
		"index.html":                  {Data: []byte("home")},
		".cress-output.json":          {Data: []byte("token")},
		"assets/file.txt":             {Data: []byte("asset")},
		"about/index.html":            {Data: []byte("about")},
		"listing/index.html/file.txt": {Data: []byte("not an index page")},
	}}
	for _, name := range []string{".cress-output.json", ".CRESS-OUTPUT.JSON", "assets", "listing"} {
		if file, err := files.Open(name); err == nil {
			t.Errorf("private marker or directory listing available: %s", name)
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, name := range []string{".", "index.html", "about", "assets/file.txt"} {
		file, err := files.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestUnavailableServer(t *testing.T) {
	response := httptest.NewRecorder()
	(&pageServer{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatal(response.Code)
	}
}

func TestSourceSelection(t *testing.T) {
	for _, name := range []string{"cress.toml", "content", "content/a.md", "content/a.assets/a", "static/file", "themes/demo/templates/page.gohtml"} {
		if !watchesSource(name) {
			t.Errorf("source excluded: %s", name)
		}
	}
	for _, name := range []string{".", ".cress/stage-a", "public/index.html", "content-other/file", "cress.toml.bak", "../content/a"} {
		if watchesSource(name) {
			t.Errorf("nonsource included: %s", name)
		}
	}
}

func TestWarningReportingFailsExplicitly(t *testing.T) {
	if err := reportWarnings(io.Discard, []string{"warning"}); err != nil {
		t.Fatal(err)
	}
	if err := reportWarnings(failedWriter{}, []string{"warning"}); err == nil {
		t.Fatal("reporting error hidden")
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
