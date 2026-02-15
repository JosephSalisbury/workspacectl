package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceName(t *testing.T) {
	got := WorkspaceName("JosephSalisbury", "workspacectl", "main")
	want := "JosephSalisbury-workspacectl-main"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDiscoverWorktreeWorkspaces(t *testing.T) {
	base := t.TempDir()

	// Create a worktree workspace layout.
	branchDir := filepath.Join(base, "repositories", "org1", "repo1", "feature-1")
	if err := os.MkdirAll(branchDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Worktrees have a .git file (not directory).
	if err := os.WriteFile(filepath.Join(branchDir, ".git"), []byte("gitdir: ../\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	workspaces, err := Discover(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(workspaces))
	}

	ws := workspaces[0]
	if ws.Name != "org1-repo1-feature-1" {
		t.Errorf("name: got %q, want %q", ws.Name, "org1-repo1-feature-1")
	}
	if ws.Type != WorkspaceTypeWorktree {
		t.Errorf("type: got %q, want %q", ws.Type, WorkspaceTypeWorktree)
	}
	if ws.Org != "org1" {
		t.Errorf("org: got %q, want %q", ws.Org, "org1")
	}
	if ws.Repo != "repo1" {
		t.Errorf("repo: got %q, want %q", ws.Repo, "repo1")
	}
	if ws.Branch != "feature-1" {
		t.Errorf("branch: got %q, want %q", ws.Branch, "feature-1")
	}
	if ws.Path != branchDir {
		t.Errorf("path: got %q, want %q", ws.Path, branchDir)
	}
}

func TestDiscoverTemporaryWorkspaces(t *testing.T) {
	base := t.TempDir()

	tempDir := filepath.Join(base, "temporary", "swift-owlbear")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		t.Fatal(err)
	}

	workspaces, err := Discover(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(workspaces))
	}

	ws := workspaces[0]
	if ws.Name != "swift-owlbear" {
		t.Errorf("name: got %q, want %q", ws.Name, "swift-owlbear")
	}
	if ws.Type != WorkspaceTypeTemporary {
		t.Errorf("type: got %q, want %q", ws.Type, WorkspaceTypeTemporary)
	}
}

func TestDiscoverMultipleWorkspaces(t *testing.T) {
	base := t.TempDir()

	worktrees := []string{
		filepath.Join(base, "repositories", "org1", "repo1", "main"),
		filepath.Join(base, "repositories", "org1", "repo1", "feature"),
		filepath.Join(base, "repositories", "org2", "repo2", "fix"),
	}
	for _, d := range worktrees {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, ".git"), []byte("gitdir: ../\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Temporary workspace (no .git file needed).
	if err := os.MkdirAll(filepath.Join(base, "temporary", "quick-session"), 0o755); err != nil {
		t.Fatal(err)
	}

	workspaces, err := Discover(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workspaces) != 4 {
		t.Fatalf("expected 4 workspaces, got %d", len(workspaces))
	}
}

func TestDiscoverIgnoresBareCloneInternals(t *testing.T) {
	base := t.TempDir()

	// Simulate a bare clone with its internal directories.
	repoDir := filepath.Join(base, "repositories", "org", "repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create HEAD file (marks this as a bare git repo).
	if err := os.WriteFile(filepath.Join(repoDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create bare clone internal directories.
	for _, dir := range []string{"hooks", "info", "objects", "refs"} {
		if err := os.MkdirAll(filepath.Join(repoDir, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Create a real worktree directory (has a .git file, not directory).
	worktreeDir := filepath.Join(repoDir, "my-feature")
	if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktreeDir, ".git"), []byte("gitdir: ../../../.bare\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	workspaces, err := Discover(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only find the worktree, not hooks/info/objects/refs.
	if len(workspaces) != 1 {
		names := make([]string, len(workspaces))
		for i, ws := range workspaces {
			names[i] = ws.Name
		}
		t.Fatalf("expected 1 workspace, got %d: %v", len(workspaces), names)
	}
	if workspaces[0].Branch != "my-feature" {
		t.Errorf("branch: got %q, want %q", workspaces[0].Branch, "my-feature")
	}
}

func TestDiscoverEmptyBaseDir(t *testing.T) {
	base := t.TempDir()

	workspaces, err := Discover(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workspaces) != 0 {
		t.Fatalf("expected 0 workspaces, got %d", len(workspaces))
	}
}
