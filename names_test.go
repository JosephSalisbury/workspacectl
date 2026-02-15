package main

import (
	"strings"
	"testing"
)

func TestGenerateNameFormat(t *testing.T) {
	name := GenerateName()
	parts := strings.SplitN(name, "-", 2)
	if len(parts) < 2 {
		t.Fatalf("expected adjective-monster format, got %q", name)
	}
	if parts[0] == "" || parts[1] == "" {
		t.Fatalf("expected non-empty parts, got %q", name)
	}
}

func TestGenerateNameVariety(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		seen[GenerateName()] = true
	}
	// With 25x25=625 combos, 100 calls should give at least 50 unique names.
	if len(seen) < 50 {
		t.Fatalf("expected at least 50 unique names from 100 calls, got %d", len(seen))
	}
}
