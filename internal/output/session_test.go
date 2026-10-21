package output

import (
	"strings"
	"testing"
)

func TestSessionIdentity(t *testing.T) {
	token := strings.Repeat("a", tokenBytes*2)
	for _, test := range []struct {
		name     string
		identity sessionIdentity
		valid    bool
	}{
		{"valid", sessionIdentity{Version: stateVersion, Site: "/site", Token: token}, true},
		{"missing", sessionIdentity{}, false},
		{"other site", sessionIdentity{Version: stateVersion, Site: "/other", Token: token}, false},
		{"other token", sessionIdentity{Version: stateVersion, Site: "/site", Token: strings.Repeat("b", tokenBytes*2)}, false},
		{"future format", sessionIdentity{Version: stateVersion + 1, Site: "/site", Token: token}, false},
		{"invalid token", sessionIdentity{Version: stateVersion, Site: "/site", Token: "../public"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.identity.matches("/site", token); got != test.valid {
				t.Fatalf("matches = %v, want %v", got, test.valid)
			}
		})
	}
}
