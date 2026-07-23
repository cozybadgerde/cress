// Package version is the single source of truth for cress's build metadata.
//
// Version is overridden at build time via the linker (goreleaser), e.g.
//
//	go build -ldflags "-X github.com/cozybadgerde/cress/internal/version.Version=1.0.0"
package version

// Name is the program name printed alongside the version.
const Name = "cress"

// Version is the release version. It defaults to "dev" for untagged local
// builds and is set by the linker for releases.
var Version = "dev"

// Info returns the one-line version string, e.g. "cress 1.0.0".
func Info() string {
	return Name + " " + Version
}
