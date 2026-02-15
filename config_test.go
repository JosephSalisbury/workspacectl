package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureConfigCreatesDefault(t *testing.T) {
	base := t.TempDir()

	cfg, err := EnsureConfig(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check default config has expected layouts.
	if len(cfg.Layouts) != 2 {
		t.Fatalf("expected 2 layouts, got %d", len(cfg.Layouts))
	}

	worktree, ok := cfg.Layouts["worktree"]
	if !ok {
		t.Fatal("expected worktree layout")
	}
	if len(worktree) != 2 {
		t.Fatalf("expected 2 worktree windows, got %d", len(worktree))
	}
	if worktree[0].Name != "claude" {
		t.Errorf("first window: got %q, want %q", worktree[0].Name, "claude")
	}

	// Check config file was created.
	configPath := filepath.Join(base, "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestEnsureConfigLoadsExisting(t *testing.T) {
	base := t.TempDir()

	configPath := filepath.Join(base, "config.yaml")
	content := []byte(`layouts:
  worktree:
    - name: shell
      command: bash
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := EnsureConfig(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	worktree, ok := cfg.Layouts["worktree"]
	if !ok {
		t.Fatal("expected worktree layout")
	}
	if len(worktree) != 1 {
		t.Fatalf("expected 1 worktree window, got %d", len(worktree))
	}
	if worktree[0].Name != "shell" {
		t.Errorf("got %q, want %q", worktree[0].Name, "shell")
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := []byte(`layouts:
  temporary:
    - name: claude
      command: claude
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	temp, ok := cfg.Layouts["temporary"]
	if !ok {
		t.Fatal("expected temporary layout")
	}
	if len(temp) != 1 {
		t.Fatalf("expected 1 window, got %d", len(temp))
	}
}

func TestLoadConfigMissing(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
