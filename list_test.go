package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunListWith(t *testing.T) {
	base := t.TempDir()

	dirs := []string{
		filepath.Join(base, "repositories", "org1", "repo1", "main"),
		filepath.Join(base, "repositories", "org1", "repo1", "feature"),
		filepath.Join(base, "temporary", "quick-session"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// runListWith prints to stdout; just verify it doesn't error.
	err := runListWith(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunListWithEmpty(t *testing.T) {
	base := t.TempDir()

	err := runListWith(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
