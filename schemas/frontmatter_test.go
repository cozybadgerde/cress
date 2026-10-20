package schemas_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/cozybadgerde/cress/internal/content"
)

//go:embed frontmatter.schema.json
var frontmatterSchema string

func TestFrontmatterSchemaMatchesRuntime(t *testing.T) {
	document, err := jsonschema.UnmarshalJSON(strings.NewReader(frontmatterSchema))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	const schemaURL = "https://example.invalid/frontmatter.schema.json"
	if err := compiler.AddResource(schemaURL, document); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(schemaURL)
	if err != nil {
		t.Fatal(err)
	}
	type testCase struct {
		name   string
		fields string
		valid  bool
	}
	tests := []testCase{
		{"title", "title = 'About'", true},
		{"draft omitted", "title = 'About'\nextra = 1", true},
		{"explicit published", "title = 'About'\ndraft = false", true},
		{"untitled draft", "draft = true", true},
		{"draft title", "draft = true\ntitle = 'Draft'", true},
		{"missing title", "", false},
		{"published without title", "draft = false", false},
		{"invalid draft", "title = 'About'\ndraft = 'true'", false},
		{"empty title", "title = ''", false},
		{"invalid title", "title = 1", false},
		{"invalid draft title", "draft = true\ntitle = []", false},
		{"unknown table", "title = 'About'\n[custom]\nvalue = 'Open'", true},
	}
	for _, whitespace := range []rune{'\t', '\n', '\v', '\f', '\r', ' ', '\u0085',
		'\u00a0', '\u1680', '\u2000', '\u2001', '\u2002', '\u2003', '\u2004',
		'\u2005', '\u2006', '\u2007', '\u2008', '\u2009', '\u200a', '\u2028',
		'\u2029', '\u202f', '\u205f', '\u3000'} {
		tests = append(tests, testCase{
			name:   fmt.Sprintf("whitespace U+%04X", whitespace),
			fields: fmt.Sprintf(`title = "\u%04X"`, whitespace),
			valid:  false,
		})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fields := make(map[string]any)
			if err := toml.Unmarshal([]byte(test.fields), &fields); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			instance, err := jsonschema.UnmarshalJSON(strings.NewReader(string(encoded)))
			if err != nil {
				t.Fatal(err)
			}
			schemaErr := schema.Validate(instance)
			_, _, runtimeErr := content.Prepare("page.md", []byte("+++\n"+test.fields+"\n+++\nBody"))
			if (schemaErr == nil) != test.valid || (runtimeErr == nil) != test.valid {
				t.Fatalf("want valid %v; schema error = %v; runtime error = %v",
					test.valid, schemaErr, runtimeErr)
			}
		})
	}
	for _, invalid := range []any{nil, "title", []any{"title"}} {
		if err := schema.Validate(invalid); err == nil {
			t.Errorf("schema accepted non-object %#v", invalid)
		}
	}
}
