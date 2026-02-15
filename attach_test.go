package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAttachCreatesSession(t *testing.T) {
	base := t.TempDir()
	baseDir = base

	// Create a workspace on disk (worktrees have a .git file).
	wsDir := filepath.Join(base, "repositories", "org", "repo", "feature")
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsDir, ".git"), []byte("gitdir: ../\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Ensure config exists.
	if _, err := EnsureConfig(base); err != nil {
		t.Fatal(err)
	}

	fake := newFakeExecutor()
	// Session does not exist.
	fake.errors["has-session"] = fmt.Errorf("exit status 1")

	err := runAttachWith(context.Background(), fake, "org-repo-feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create session, rename window, send-keys, create pane, then attach.
	foundNewSession := false
	foundRenameWindow := false
	foundSendKeysClaude := false
	foundSplitWindow := false
	foundSendKeysDiff := false
	foundAttach := false
	for _, call := range fake.calls {
		if strings.Contains(call, "new-session") {
			foundNewSession = true
		}
		if strings.Contains(call, "rename-window") {
			foundRenameWindow = true
		}
		if strings.Contains(call, "send-keys") && strings.Contains(call, "claude") {
			foundSendKeysClaude = true
		}
		if strings.Contains(call, "split-window") {
			foundSplitWindow = true
		}
		if strings.Contains(call, "send-keys") && strings.Contains(call, "diff") {
			foundSendKeysDiff = true
		}
		if strings.Contains(call, "attach-session") || strings.Contains(call, "switch-client") {
			foundAttach = true
		}
	}

	if !foundNewSession {
		t.Error("expected new-session call")
	}
	if !foundRenameWindow {
		t.Error("expected rename-window call")
	}
	if !foundSendKeysClaude {
		t.Error("expected send-keys call with claude command")
	}
	if !foundSplitWindow {
		t.Error("expected split-window call for diff pane")
	}
	if !foundSendKeysDiff {
		t.Error("expected send-keys call with diff command")
	}
	if !foundAttach {
		t.Error("expected attach/switch call")
	}
}

func TestRunAttachExistingSession(t *testing.T) {
	base := t.TempDir()
	baseDir = base

	wsDir := filepath.Join(base, "repositories", "org", "repo", "feature")
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsDir, ".git"), []byte("gitdir: ../\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureConfig(base); err != nil {
		t.Fatal(err)
	}

	fake := newFakeExecutor()
	// Session already exists (has-session succeeds).

	err := runAttachWith(context.Background(), fake, "org-repo-feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT create session, just attach.
	for _, call := range fake.calls {
		if strings.Contains(call, "new-session") {
			t.Error("should not create new session when one exists")
		}
	}

	foundAttach := false
	for _, call := range fake.calls {
		if strings.Contains(call, "switch-client") || strings.Contains(call, "attach-session") {
			foundAttach = true
		}
	}
	if !foundAttach {
		t.Error("expected attach call")
	}
}

func TestRunAttachWorkspaceNotFound(t *testing.T) {
	base := t.TempDir()
	baseDir = base

	fake := newFakeExecutor()
	err := runAttachWith(context.Background(), fake, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing workspace")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}
