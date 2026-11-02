package theme

import (
	"strings"
	"testing"
)

func TestInsertStylesheet(t *testing.T) {
	const anchor = `{{block "theme-styles" .}}{{end}}`
	for _, test := range []struct {
		name   string
		source string
		want   string
		err    string
	}{
		{"single anchor", "<head>\n" + anchor + "\n</head>",
			"<head>\n" + `<link rel="stylesheet" href="{{asset "site.css"}}">` + "\n</head>", ""},
		{"missing anchor", "<head></head>", "", "found 0"},
		{"duplicate anchor", anchor + "\n" + anchor, "", "found 2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := insertStylesheet([]byte(test.source))
			if test.err != "" {
				if err == nil || !strings.Contains(err.Error(), "embedded templates/shared.gohtml") ||
					!strings.Contains(err.Error(), test.err) {
					t.Fatalf("expected contextual anchor error, got %v", err)
				}
				return
			}
			if err != nil || string(got) != test.want {
				t.Fatalf("source = %q, error = %v; want %q", got, err, test.want)
			}
		})
	}
}
