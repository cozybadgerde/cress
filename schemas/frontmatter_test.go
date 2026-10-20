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

//go:embed site.schema.json
var siteSchema string

func compileSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	for name, source := range map[string]string{
		"frontmatter.schema.json": frontmatterSchema,
		"site.schema.json":        siteSchema,
	} {
		document, err := jsonschema.UnmarshalJSON(strings.NewReader(source))
		if err != nil {
			t.Fatal(err)
		}
		if err := compiler.AddResource("https://example.invalid/"+name, document); err != nil {
			t.Fatal(err)
		}
	}
	schema, err := compiler.Compile("https://example.invalid/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func TestFrontmatterSchemaMatchesRuntime(t *testing.T) {
	schema := compileSchema(t, "frontmatter.schema.json")
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
		{"description", "title = 'About'\ndescription = ''", true},
		{"invalid description", "title = 'About'\ndescription = 1", false},
		{"ordinary", "title = 'About'\ntype = 'ordinary'", true},
		{"landing", "title = 'About'\ntype = 'landing'", true},
		{"excluded list", "type = 'list'\ndraft = true", true},
		{"invalid type", "title = 'About'\ntype = 'article'", false},
		{"invalid type kind", "title = 'About'\ntype = false", false},
		{"image", "title = 'About'\n[image]\nsrc = '/photo.jpg'\nalt = 'Photo'", true},
		{"decorative image", "title = 'About'\n[image]\nsrc = '/photo.jpg'\nalt = ''", true},
		{"cleared image", "title = 'About'\n[image]\nsrc = ''", true},
		{"missing alternative", "title = 'About'\n[image]\nsrc = '/photo.jpg'", false},
		{"missing source", "title = 'About'\n[image]\nalt = ''", false},
		{"relative image", "title = 'About'\n[image]\nsrc = 'photo.jpg'\nalt = ''", false},
		{"invalid draft image", "draft = true\nimage = false", false},
		{"invalid alternative", "title = 'About'\n[image]\nsrc = ''\nalt = 1", false},
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
			config, _, err := content.ParseConfig(nil)
			if err != nil {
				t.Fatal(err)
			}
			_, _, runtimeErr := content.Prepare("page.md", []byte("+++\n"+test.fields+"\n+++\nBody"), config)
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
