package release_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleaseArchives(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping archive integration test")
	}
	directory := os.Getenv("CRESS_RELEASE_DIR")
	if directory == "" {
		t.Skip("run mise run test:release against packaged artifacts")
	}
	metadata, err := os.ReadFile(filepath.Join(directory, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	var release struct {
		Project string `json:"project_name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(metadata, &release); err != nil {
		t.Fatal(err)
	}
	if release.Project != "cress" || release.Version == "" {
		t.Fatalf("invalid release metadata: %s", metadata)
	}
	if tag := os.Getenv("RELEASE_TAG"); tag != "" && release.Version != strings.TrimPrefix(tag, "v") {
		t.Fatalf("packaged version %q does not match tag %q", release.Version, tag)
	}

	targets := []string{"darwin_arm64", "linux_amd64", "linux_arm64"}
	expected := make(map[string]bool)
	for _, target := range targets {
		prefix := "cress_" + release.Version + "_" + target
		expected[prefix+".tar.gz"] = true
		expected[prefix+".licenses.csv"] = true
	}
	checkReleaseChecksums(t, directory, expected)

	host := runtime.GOOS + "_" + runtime.GOARCH
	executed := false
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			prefix := filepath.Join(directory, "cress_"+release.Version+"_"+target)
			checkLicenseReport(t, prefix+".licenses.csv")
			binary := readArchiveBinary(t, prefix+".tar.gz")
			info, err := buildinfo.Read(bytes.NewReader(binary))
			if err != nil {
				t.Fatal(err)
			}
			settings := make(map[string]string)
			for _, setting := range info.Settings {
				settings[setting.Key] = setting.Value
			}
			if settings["GOOS"]+"_"+settings["GOARCH"] != target {
				t.Fatalf("binary target does not match archive: %v", settings)
			}
			if target == host {
				executed = true
				checkArchivedCommand(t, binary, release.Version)
			}
		})
	}
	if !executed {
		t.Fatalf("no archive executable on host %s", host)
	}
}

func checkReleaseChecksums(t *testing.T, directory string, expected map[string]bool) {
	t.Helper()
	manifest, err := os.ReadFile(filepath.Join(directory, "checksums.txt"))
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(manifest)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || !expected[fields[1]] || seen[fields[1]] {
			t.Fatalf("unexpected or duplicate checksum entry: %q", line)
		}
		name := fields[1]
		file, err := os.Open(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.New()
		_, readErr := io.Copy(hash, file)
		closeErr := file.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read %s: %v; close: %v", name, readErr, closeErr)
		}
		if fmt.Sprintf("%x", hash.Sum(nil)) != fields[0] {
			t.Fatalf("checksum mismatch for %s", name)
		}
		seen[name] = true
	}
	for name := range expected {
		if !seen[name] {
			t.Fatalf("missing checksum for %s", name)
		}
	}
	for _, pattern := range []string{"*.tar.gz", "*.licenses.csv"} {
		files, err := filepath.Glob(filepath.Join(directory, pattern))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			if !expected[filepath.Base(file)] {
				t.Fatalf("unexpected release artifact: %s", file)
			}
		}
	}
}

func checkLicenseReport(t *testing.T, name string) {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = 3
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 {
		t.Fatalf("empty license report: %s", name)
	}
	for _, record := range records {
		for _, field := range record {
			if strings.TrimSpace(field) == "" {
				t.Fatalf("empty license field in %s: %v", name, record)
			}
		}
	}
}

func readArchiveBinary(t *testing.T, name string) []byte {
	t.Helper()
	file, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Error(err)
		}
	})
	compressed, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := compressed.Close(); err != nil {
			t.Error(err)
		}
	})
	reader := tar.NewReader(compressed)
	required := map[string]bool{
		"cress": false, "LICENSE": false, "README.md": false,
		"THIRD-PARTY-NOTICES.txt": false, "docs/INSTALL.md": false,
	}
	var binary []byte
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		found, needed := required[header.Name]
		if !needed {
			continue
		}
		if found || header.Typeflag != tar.TypeReg || header.Size == 0 {
			t.Fatalf("duplicate, nonregular, or empty archive member: %s", header.Name)
		}
		required[header.Name] = true
		if header.Name == "cress" {
			if header.Mode&0o111 == 0 {
				t.Fatal("archived cress is not executable")
			}
			binary, err = io.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	for name, found := range required {
		if !found {
			t.Fatalf("missing archive member: %s", name)
		}
	}
	if _, err := io.Copy(io.Discard, compressed); err != nil {
		t.Fatal(err)
	}
	return binary
}

func checkArchivedCommand(t *testing.T, binary []byte, version string) {
	t.Helper()
	directory := t.TempDir()
	executable := filepath.Join(directory, "cress")
	if err := os.WriteFile(executable, binary, 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(t *testing.T, args ...string) string {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), executable, args...)
		cmd.Dir = directory
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("cress %v: %v\n%s", args, err, output)
		}
		return strings.TrimSpace(string(output))
	}
	if output := run(t, "--version"); output != "cress version "+version {
		t.Fatalf("unexpected executable version: %q", output)
	}
	for _, theme := range []string{"poppy", "clover", "anemone"} {
		t.Run(theme, func(t *testing.T) {
			run(t, "site", "init", "--theme", theme, theme)
			run(t, "site", "build", theme)
			for _, name := range []string{
				"index.html", "about/index.html", "journal/list/index.html",
				"tour/index.html", "themes/" + theme + "/site.css",
			} {
				file, err := os.Stat(filepath.Join(directory, theme, "public", filepath.FromSlash(name)))
				if err != nil {
					t.Fatal(err)
				}
				if !file.Mode().IsRegular() || file.Size() == 0 {
					t.Fatalf("missing or empty generated file: %s", name)
				}
			}
		})
	}
}
