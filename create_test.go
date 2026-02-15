package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCreateWorktree(t *testing.T) {
	base := t.TempDir()
	baseDir = base

	// Set up a fake bare repo dir so bare clone is skipped.
	repoDir := filepath.Join(base, "repositories", "testorg", "testrepo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}

	createOrg = "testorg"
	createRepo = "testrepo"
	createBranch = "test-branch"

	fake := newFakeExecutor()
	// Branch does not exist on remote.
	fake.results["ls-remote"] = ""
	// Default branch is main.
	fake.results["symbolic-ref"] = "refs/heads/main"

	err := runCreateWorktree(context.Background(), fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: ls-remote, symbolic-ref, branch create, worktree add.
	if len(fake.calls) != 4 {
		t.Fatalf("expected 4 git calls, got %d: %v", len(fake.calls), fake.calls)
	}

	if !strings.Contains(fake.calls[0], "ls-remote") {
		t.Errorf("call 0: expected ls-remote, got %q", fake.calls[0])
	}
	if !strings.Contains(fake.calls[1], "symbolic-ref") {
		t.Errorf("call 1: expected symbolic-ref, got %q", fake.calls[1])
	}
	if !strings.Contains(fake.calls[2], "branch test-branch main") {
		t.Errorf("call 2: expected branch create, got %q", fake.calls[2])
	}
	if !strings.Contains(fake.calls[3], "worktree add") {
		t.Errorf("call 3: expected worktree add, got %q", fake.calls[3])
	}
}

func TestRunCreateWorktreeBranchExists(t *testing.T) {
	base := t.TempDir()
	baseDir = base

	repoDir := filepath.Join(base, "repositories", "testorg", "testrepo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}

	createOrg = "testorg"
	createRepo = "testrepo"
	createBranch = "existing-branch"

	fake := newFakeExecutor()
	// Branch exists on remote.
	fake.results["ls-remote"] = "abc123\trefs/heads/existing-branch"

	err := runCreateWorktree(context.Background(), fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have ls-remote and worktree add (no symbolic-ref or branch create).
	if len(fake.calls) != 2 {
		t.Fatalf("expected 2 git calls, got %d: %v", len(fake.calls), fake.calls)
	}
	if !strings.Contains(fake.calls[0], "ls-remote") {
		t.Errorf("call 0: expected ls-remote, got %q", fake.calls[0])
	}
	if !strings.Contains(fake.calls[1], "worktree add") {
		t.Errorf("call 1: expected worktree add, got %q", fake.calls[1])
	}
}

func TestRunCreateWorktreeMissingOrg(t *testing.T) {
	createOrg = ""
	createRepo = "repo"
	createBranch = "branch"

	err := runCreateWorktree(context.Background(), newFakeExecutor())
	if err == nil {
		t.Fatal("expected error for missing org")
	}
}

func TestRunCreateWorktreeMissingRepo(t *testing.T) {
	createOrg = "org"
	createRepo = ""
	createBranch = "branch"

	err := runCreateWorktree(context.Background(), newFakeExecutor())
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
}

func TestRunCreateTemporary(t *testing.T) {
	base := t.TempDir()
	baseDir = base
	createBranch = "test-temp"

	err := runCreateTemporary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tempDir := filepath.Join(base, "temporary", "test-temp")
	if _, err := os.Stat(tempDir); err != nil {
		t.Fatalf("temporary directory not created: %v", err)
	}
}
