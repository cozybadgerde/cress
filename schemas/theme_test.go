package schemas_test

import (
	"encoding/json"
	"testing"

	"github.com/cozybadgerde/cress/internal/theme"
)

func TestThemeSchemaMatchesRuntime(t *testing.T) {
	schema := compileSchema(t, "theme.schema.json")
	for _, test := range []struct {
		source string
		valid  bool
	}{
		{`{"name":"demo","version":"1.0.0","cress":{"minimum":"1.0.0","major":1}}`, true},
		{`{"name":"demo","version":"1.0.0-rc.1+build","extra":true,"cress":{"minimum":"1.0.0-rc.1","major":1,"extra":true}}`, true},
		{`{"name":"demo","version":"1.0","cress":{"minimum":"1.0.0","major":1}}`, false},
		{`{"name":"demo","version":"v1.0.0","cress":{"minimum":"1.0.0","major":1}}`, false},
		{`{"name":"demo","version":"1.0.0-01","cress":{"minimum":"1.0.0","major":1}}`, false},
		{`{"name":"../demo","version":"1.0.0","cress":{"minimum":"1.0.0","major":1}}`, false},
		{`{"name":"demo","version":"1.0.0","cress":{"minimum":"1.0.0"}}`, false},
		{`{"name":"demo","version":"1.0.0","cress":{"minimum":"1.0.0","major":null}}`, false},
		{`{"name":"demo","version":"1.0.0","cress":[]}`, false},
		{`null`, false},
		{`[]`, false},
	} {
		t.Run(test.source, func(t *testing.T) {
			var instance any
			if err := json.Unmarshal([]byte(test.source), &instance); err != nil {
				t.Fatal(err)
			}
			schemaErr := schema.Validate(instance)
			_, _, runtimeErr := theme.ParseManifest([]byte(test.source))
			if (schemaErr == nil) != test.valid || (runtimeErr == nil) != test.valid {
				t.Fatalf("want valid=%v; schema=%v runtime=%v", test.valid, schemaErr, runtimeErr)
			}
		})
	}
}
