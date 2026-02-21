package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupWorktreeWorkspace(t *testing.T, base, branch string) {
	t.Helper()
	wsDir := filepath.Join(base, "repositories", "org", "repo", branch)
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsDir, ".git"), []byte("gitdir: ../\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newCleanFakeExecutor() *fakeExecutor {
	fake := newFakeExecutor()
	// Branch exists on remote, no unpushed commits.
	fake.results["ls-remote --heads"] = "abc123\trefs/heads/feature"
	fake.results["origin/feature..HEAD"] = ""
	// No session by default.
	fake.errors["has-session"] = fmt.Errorf("exit status 1")
	return fake
}

func TestDeleteWorktreeClean(t *testing.T) {
	base := t.TempDir()
	baseDir = base
	setupWorktreeWorkspace(t, base, "feature")

	fake := newCleanFakeExecutor()
	// status --porcelain returns empty → clean
	fake.results["status --porcelain"] = ""

	err := runDeleteWith(context.Background(), fake, strings.NewReader(""), "org-repo-feature", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundRemove := false
	for _, call := range fake.calls {
		if strings.Contains(call, "worktree remove") {
			foundRemove = true
		}
	}
	if !foundRemove {
		t.Error("expected worktree remove call")
	}
}

func TestDeleteWorktreeDirtyConfirmed(t *testing.T) {
	base := t.TempDir()
	baseDir = base
	setupWorktreeWorkspace(t, base, "feature")

	fake := newCleanFakeExecutor()
	// Dirty: uncommitted changes.
	fake.results["status --porcelain"] = "M  file.go"

	err := runDeleteWith(context.Background(), fake, strings.NewReader("y\n"), "org-repo-feature", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundRemove := false
	for _, call := range fake.calls {
		if strings.Contains(call, "worktree remove") {
			foundRemove = true
		}
	}
	if !foundRemove {
		t.Error("expected worktree remove call after confirmation")
	}
}

func TestDeleteWorktreeDirtyAborted(t *testing.T) {
	base := t.TempDir()
	baseDir = base
	setupWorktreeWorkspace(t, base, "feature")

	fake := newCleanFakeExecutor()
	// Dirty: uncommitted changes.
	fake.results["status --porcelain"] = "M  file.go"

	err := runDeleteWith(context.Background(), fake, strings.NewReader("n\n"), "org-repo-feature", false)
	if err == nil {
		t.Fatal("expected error when aborted")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("expected 'aborted' in error, got: %v", err)
	}

	for _, call := range fake.calls {
		if strings.Contains(call, "worktree remove") {
			t.Error("worktree remove should not be called when aborted")
		}
	}
}

func TestDeleteWorktreeForce(t *testing.T) {
	base := t.TempDir()
	baseDir = base
	setupWorktreeWorkspace(t, base, "feature")

	fake := newCleanFakeExecutor()
	// Dirty: unpushed commits.
	fake.results["status --porcelain"] = "M  file.go"

	err := runDeleteWith(context.Background(), fake, strings.NewReader(""), "org-repo-feature", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundRemove := false
	for _, call := range fake.calls {
		if strings.Contains(call, "worktree remove") {
			foundRemove = true
		}
	}
	if !foundRemove {
		t.Error("expected worktree remove call with --force")
	}
}

func TestDeleteWorktreeLastInBareClone(t *testing.T) {
	base := t.TempDir()
	baseDir = base
	setupWorktreeWorkspace(t, base, "feature")

	repoDir := filepath.Join(base, "repositories", "org", "repo")
	orgDir := filepath.Join(base, "repositories", "org")

	fake := newCleanFakeExecutor()
	fake.results["status --porcelain"] = ""

	err := runDeleteWith(context.Background(), fake, strings.NewReader(""), "org-repo-feature", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// repoDir (bare clone) should be removed — it was the last worktree.
	if _, err := os.Stat(repoDir); !os.IsNotExist(err) {
		t.Error("expected repoDir to be removed after last worktree deleted")
	}
	// orgDir should also be removed — it became empty.
	if _, err := os.Stat(orgDir); !os.IsNotExist(err) {
		t.Error("expected orgDir to be removed after it became empty")
	}
}

func TestDeleteWorktreeNotLastInBareClone(t *testing.T) {
	base := t.TempDir()
	baseDir = base
	setupWorktreeWorkspace(t, base, "feature")
	setupWorktreeWorkspace(t, base, "another")

	repoDir := filepath.Join(base, "repositories", "org", "repo")

	fake := newCleanFakeExecutor()
	fake.results["status --porcelain"] = ""

	err := runDeleteWith(context.Background(), fake, strings.NewReader(""), "org-repo-feature", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// repoDir should remain — another worktree still exists.
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		t.Error("expected repoDir to remain when another worktree still exists")
	}
}

func TestDeleteWorktreeKillsSession(t *testing.T) {
	base := t.TempDir()
	baseDir = base
	setupWorktreeWorkspace(t, base, "feature")

	fake := newCleanFakeExecutor()
	fake.results["status --porcelain"] = ""
	// Override: session exists (has-session succeeds).
	delete(fake.errors, "has-session")

	err := runDeleteWith(context.Background(), fake, strings.NewReader(""), "org-repo-feature", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundKill := false
	for _, call := range fake.calls {
		if strings.Contains(call, "kill-session") {
			foundKill = true
		}
	}
	if !foundKill {
		t.Error("expected kill-session call when tmux session exists")
	}
}

func TestDeleteTemporary(t *testing.T) {
	base := t.TempDir()
	baseDir = base

	wsDir := filepath.Join(base, "temporary", "myws")
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	fake := newFakeExecutor()
	fake.errors["has-session"] = fmt.Errorf("exit status 1")

	err := runDeleteWith(context.Background(), fake, strings.NewReader(""), "myws", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(wsDir); !os.IsNotExist(err) {
		t.Error("expected temporary workspace directory to be removed")
	}
}

func TestDeleteNotFound(t *testing.T) {
	base := t.TempDir()
	baseDir = base

	fake := newFakeExecutor()
	err := runDeleteWith(context.Background(), fake, strings.NewReader(""), "nonexistent", false)
	if err == nil {
		t.Fatal("expected error for unknown workspace name")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}
