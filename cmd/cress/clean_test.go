package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Only an explicit yes deletes. Everything else, including nonsense and an
// empty line, has to fall through to the answer that removes nothing.
func TestAffirmative(t *testing.T) {
	tests := []struct {
		answer string
		want   bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"  yes  \n", true},

		{"", false},
		{"\n", false},
		{"n\n", false},
		{"no\n", false},
		{"nope\n", false},
		{"ye\n", false},
		{"yeah\n", false},
		{"yesterday\n", false},
		{"1\n", false},
		{"lizard\n", false},
	}

	for _, tc := range tests {
		t.Run(strings.TrimSpace(tc.answer), func(t *testing.T) {
			if got := affirmative(tc.answer); got != tc.want {
				t.Errorf("affirmative(%q) = %v, want %v", tc.answer, got, tc.want)
			}
		})
	}
}

// Anything that is not a terminal has to read as non-interactive, or a
// redirected or piped stdin would be prompted at and answered by end-of-input.
func TestInteractive(t *testing.T) {
	regular, err := os.Create(filepath.Join(t.TempDir(), "stdin"))
	if err != nil {
		t.Fatalf("creating file: %v", err)
	}
	defer func() { _ = regular.Close() }()

	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	defer func() { _ = pipeRead.Close(); _ = pipeWrite.Close() }()

	tests := []struct {
		name string
		in   interface{ Read([]byte) (int, error) }
	}{
		{"a string reader", strings.NewReader("y\n")},
		{"a redirected file", regular},
		{"a pipe", pipeRead},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if interactive(tc.in) {
				t.Errorf("interactive(%s) = true, want false", tc.name)
			}
		})
	}
}
