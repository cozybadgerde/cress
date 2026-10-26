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

func TestTransactionEvidenceChecks(t *testing.T) {
	token, previous := strings.Repeat("b", tokenBytes*2), strings.Repeat("a", tokenBytes*2)
	pending := func(phase string, hadPublic bool) transaction {
		transaction := transaction{Token: token, Phase: phase, HadPublic: hadPublic}
		if hadPublic {
			transaction.Previous = previous
		}
		return transaction
	}
	ownership := []struct {
		name          string
		pending       transaction
		backup, stage string
		invalid       bool
	}{
		{"nothing", pending("ready", true), "absent", "absent", false},
		{"own stage", pending("ready", true), "absent", token, false},
		{"empty stage while staging", pending("staging", true), "absent", "empty", false},
		{"empty stage when ready", pending("ready", true), "absent", "empty", true},
		{"foreign stage", pending("ready", true), "absent", previous, true},
		{"previous backup", pending("ready", true), previous, "absent", false},
		{"foreign backup", pending("ready", true), token, "absent", true},
		{"backup without previous public", pending("ready", false), "empty", "absent", true},
		{"empty backup of empty previous", transaction{Token: token, Phase: "ready", HadPublic: true}, "empty", "absent", false},
	}
	for _, test := range ownership {
		t.Run("ownership/"+test.name, func(t *testing.T) {
			if err := test.pending.checkOwnership(test.backup, test.stage); (err != nil) != test.invalid {
				t.Fatalf("error = %v, want invalid %v", err, test.invalid)
			}
		})
	}
	published := []struct {
		name          string
		pending       transaction
		backup, stage string
		invalid       bool
	}{
		{"staging", pending("staging", true), previous, "absent", true},
		{"stage remains", pending("ready", true), previous, token, true},
		{"ready without backup", pending("ready", true), "absent", "absent", true},
		{"ready without previous", pending("ready", false), "absent", "absent", false},
		{"ready with backup", pending("ready", true), previous, "absent", false},
		{"published without backup", pending("published", true), "absent", "absent", false},
	}
	for _, test := range published {
		t.Run("published/"+test.name, func(t *testing.T) {
			if err := test.pending.checkPublished(test.backup, test.stage); (err != nil) != test.invalid {
				t.Fatalf("error = %v, want invalid %v", err, test.invalid)
			}
		})
	}
	previousOutput := []struct {
		name           string
		pending        transaction
		public, backup string
		restore        bool
		invalid        bool
	}{
		{"restore backup", pending("ready", true), "absent", previous, true, false},
		{"backup not restorable while staging", pending("staging", true), "absent", previous, false, true},
		{"previous intact", pending("staging", true), previous, "absent", false, false},
		{"previously absent", pending("staging", false), "absent", "absent", false, false},
		{"unverifiable public", pending("ready", true), "unknown", "absent", false, true},
		{"previous and backup both present", pending("ready", true), previous, previous, false, true},
		{"confirmed publication unverifiable", pending("published", true), "absent", "absent", false, true},
	}
	for _, test := range previousOutput {
		t.Run("previous/"+test.name, func(t *testing.T) {
			restore, err := test.pending.checkPrevious(test.public, test.backup)
			if (err != nil) != test.invalid || restore != test.restore {
				t.Fatalf("restore = %v, error = %v; want restore %v, invalid %v", restore, err, test.restore, test.invalid)
			}
		})
	}
}
