package content_test

import (
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/content"
)

func TestExplicitDraftVisibility(t *testing.T) {
	for _, included := range []bool{false, true} {
		t.Run(map[bool]string{false: "excluded", true: "included"}[included], func(t *testing.T) {
			config, _, err := content.ParseConfig([]byte("[site]\nhost='https://example.org'\npath='/company/'"))
			if err != nil {
				t.Fatal(err)
			}
			config.IncludeDrafts = included
			prepared, _, err := content.Prepare("private.md", []byte("+++\ntitle='Private'\ndraft=true\n+++\nPrivate body"), config)
			if err != nil || !prepared.Draft || strings.Contains(string(prepared.Data.Body), "Private body") != included {
				t.Fatalf("prepared = %+v, %v", prepared, err)
			}
			if included && prepared.Data.Canonical != "https://example.org/company/private/" {
				t.Fatalf("draft origin was rewritten: %s", prepared.Data.Canonical)
			}
			_, _, err = content.Prepare("private.md", []byte("+++\ndraft=true\n+++\nUntitled"), config)
			if (err != nil) != included {
				t.Fatalf("untitled draft validation = %v, included %v", err, included)
			}
		})
	}
}
