package release_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseGate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping release gate integration test")
	}
	for _, tool := range []string{"jq", "sha256sum"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("release gate requires %s on the GitHub runner", tool)
		}
	}
	script, err := filepath.Abs("../../.github/scripts/release-gate.sh")
	if err != nil {
		t.Fatal(err)
	}

	type asset struct {
		Name   string `json:"name"`
		Digest string `json:"digest"`
		State  string `json:"state"`
	}
	type release struct {
		ID         int     `json:"id"`
		Tag        string  `json:"tag_name"`
		Draft      bool    `json:"draft"`
		Prerelease bool    `json:"prerelease"`
		Assets     []asset `json:"assets"`
	}
	const tag = "v0.0.0-gate.1"
	cases := []struct {
		name        string
		action      string
		matches     int
		tag         string
		failList    bool
		mutate      func(*release)
		wantSuccess bool
		wantPatch   bool
	}{
		{name: "new tag", action: "check", wantSuccess: true},
		{name: "existing draft", action: "check", matches: 1},
		{name: "release listing failure", action: "check", failList: true},
		{name: "duplicate drafts", action: "publish", matches: 2},
		{name: "matching draft", action: "publish", matches: 1, wantSuccess: true, wantPatch: true},
		{name: "stable with build metadata", action: "publish", matches: 1, tag: "v1.0.0+build-1",
			mutate: func(r *release) { r.Prerelease = false }, wantSuccess: true, wantPatch: true},
		{name: "wrong asset digest", action: "publish", matches: 1, mutate: func(r *release) {
			r.Assets[0].Digest = "sha256:incorrect"
		}},
		{name: "wrong prerelease flag", action: "publish", matches: 1, mutate: func(r *release) {
			r.Prerelease = false
		}},
		{name: "already published", action: "publish", matches: 1, mutate: func(r *release) {
			r.Draft = false
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			dist := filepath.Join(dir, "dist")
			if err := os.Mkdir(dist, 0o700); err != nil {
				t.Fatal(err)
			}
			releaseTag := tc.tag
			if releaseTag == "" {
				releaseTag = tag
			}
			r := release{ID: 42, Tag: releaseTag, Draft: true, Prerelease: true}
			for _, name := range []string{
				"darwin.tar.gz", "linux-amd64.tar.gz", "linux-arm64.tar.gz",
				"darwin.licenses.csv", "linux-amd64.licenses.csv", "linux-arm64.licenses.csv",
				"checksums.txt", "checksums.txt.sigstore.json",
			} {
				data := []byte(name)
				if err := os.WriteFile(filepath.Join(dist, name), data, 0o600); err != nil {
					t.Fatal(err)
				}
				r.Assets = append(r.Assets, asset{Name: name, Digest: fmt.Sprintf("sha256:%x", sha256.Sum256(data)), State: "uploaded"})
			}
			if tc.mutate != nil {
				tc.mutate(&r)
			}
			matches := make([]release, tc.matches)
			for i := range matches {
				matches[i] = r
			}
			writeJSON := func(name string, value any) string {
				t.Helper()
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dir, name)
				if err := os.WriteFile(path, data, 0o600); err != nil {
					t.Fatal(err)
				}
				return path
			}
			list := writeJSON("releases.json", matches)
			detail := writeJSON("release.json", r)
			r.Draft = false
			published := writeJSON("published.json", r)
			calls := filepath.Join(dir, "calls")
			gh := `#!/usr/bin/env bash
printf '%s\n' "$*" >>"$MOCK_CALLS"
case "$*" in
  *"--method PATCH"*) cat "$MOCK_PUBLISHED" ;;
  *"/git/ref/tags/"*) printf '{}\n' ;;
  *"/releases/42"*) cat "$MOCK_DETAIL" ;;
  *"/releases?per_page=100"*)
    if [[ $MOCK_FAIL_LIST == true ]]; then
      echo "release listing failed" >&2
      exit 1
    fi
    cat "$MOCK_LIST"
    ;;
  *) exit 1 ;;
esac
`
			if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(gh), 0o700); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", script, tc.action)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(),
				"PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"GITHUB_REPOSITORY=example/cress", "RELEASE_TAG="+releaseTag,
				"MOCK_LIST="+list, "MOCK_DETAIL="+detail, "MOCK_PUBLISHED="+published, "MOCK_CALLS="+calls,
				fmt.Sprintf("MOCK_FAIL_LIST=%t", tc.failList),
			)
			output, err := cmd.CombinedOutput()
			if (err == nil) != tc.wantSuccess {
				t.Fatalf("unexpected result: %v\n%s", err, output)
			}
			if tc.failList && !strings.Contains(string(output), "release listing failed") {
				t.Fatalf("missing listing error: %s", output)
			}
			log, err := os.ReadFile(calls)
			if err != nil {
				t.Fatal(err)
			}
			if patched := strings.Contains(string(log), "--method PATCH"); patched != tc.wantPatch {
				t.Fatalf("unexpected publication call: %s", log)
			}
		})
	}
}
