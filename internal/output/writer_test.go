package output

import (
	"strings"
	"testing"
)

func TestOwnershipTokens(t *testing.T) {
	for _, test := range []struct {
		token string
		valid bool
	}{
		{strings.Repeat("a", tokenBytes*2), true},
		{strings.Repeat("0", tokenBytes*2), true},
		{"", false},
		{"../public", false},
		{strings.Repeat("a", tokenBytes*2-1), false},
		{strings.Repeat("a", tokenBytes*2+2), false},
		{strings.Repeat("A", tokenBytes*2), false},
		{strings.Repeat("z", tokenBytes*2), false},
	} {
		t.Run(test.token, func(t *testing.T) {
			if got := validToken(test.token); got != test.valid {
				t.Fatalf("validToken = %v, want %v", got, test.valid)
			}
		})
	}
}

func TestStateValidation(t *testing.T) {
	oldToken, newToken := strings.Repeat("a", tokenBytes*2), strings.Repeat("b", tokenBytes*2)
	tests := []struct {
		name    string
		record  state
		invalid bool
	}{
		{name: "new", record: state{Version: 1}},
		{name: "owned", record: state{Version: 1, Current: oldToken}},
		{name: "missing version", record: state{}, invalid: true},
		{name: "future version", record: state{Version: 2}, invalid: true},
		{name: "invalid owner", record: state{Version: 1, Current: "toilet"}, invalid: true},
	}
	for _, phase := range []string{"staging", "ready", "published", "unknown"} {
		record := state{Version: 1, Current: oldToken, Pending: &transaction{
			Token: newToken, Previous: oldToken, HadPublic: true, Phase: phase,
		}}
		if phase == "published" {
			record.Current = newToken
		}
		tests = append(tests, struct {
			name    string
			record  state
			invalid bool
		}{phase, record, phase == "unknown"})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.record.validate(); (err != nil) != test.invalid {
				t.Fatalf("error = %v, want invalid %v", err, test.invalid)
			}
		})
	}
	for _, mutate := range []func(*state){
		func(record *state) { record.Pending.Token = "../public" },
		func(record *state) { record.Pending.Previous = "bad" },
		func(record *state) { record.Pending.Token = oldToken },
		func(record *state) { record.Pending.HadPublic = false },
		func(record *state) { record.Pending.Phase = "published" },
		func(record *state) { record.Current = newToken },
	} {
		record := state{Version: 1, Current: oldToken, Pending: &transaction{
			Token: newToken, Previous: oldToken, HadPublic: true, Phase: "ready",
		}}
		mutate(&record)
		if err := record.validate(); err == nil {
			t.Fatalf("invalid transaction accepted: %+v", record)
		}
	}
}
