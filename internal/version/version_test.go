package version_test

import (
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/version"
)

func TestInfo(t *testing.T) {
	got := version.Info()

	if !strings.HasPrefix(got, version.Name+" ") {
		t.Errorf("Info() = %q, want it to start with %q", got, version.Name+" ")
	}
	if !strings.Contains(got, version.Version) {
		t.Errorf("Info() = %q, want it to contain version %q", got, version.Version)
	}
}
