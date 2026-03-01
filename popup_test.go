package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupPopupWorkspaces(t *testing.T) string {
	t.Helper()
	base := t.TempDir()

	// Create a worktree workspace.
	wsDir := filepath.Join(base, "repositories", "acme", "anvil", "main")
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write a .git file (not dir) to mark it as a worktree.
	if err := os.WriteFile(filepath.Join(wsDir, ".git"), []byte("gitdir: ../.."), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a temporary workspace.
	tmpDir := filepath.Join(base, "temporary", "swift-owlbear")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatal(err)
	}

	return base
}

func TestPopupCallsDisplayMenu(t *testing.T) {
	origBaseDir := baseDir
	defer func() { baseDir = origBaseDir }()
	baseDir = setupPopupWorkspaces(t)

	fake := newFakeExecutor()
	err := runPopupWith(context.Background(), fake, "/usr/local/bin/workspacectl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.calls) != 1 {
		t.Fatalf("expected 1 call, got %d: %v", len(fake.calls), fake.calls)
	}
	if !strings.Contains(fake.calls[0], "display-menu") {
		t.Fatalf("expected display-menu call, got %q", fake.calls[0])
	}
}

func TestPopupContainsWorkspaceNames(t *testing.T) {
	origBaseDir := baseDir
	defer func() { baseDir = origBaseDir }()
	baseDir = setupPopupWorkspaces(t)

	fake := newFakeExecutor()
	err := runPopupWith(context.Background(), fake, "/usr/local/bin/workspacectl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	call := fake.calls[0]
	if !strings.Contains(call, "acme-anvil-main") {
		t.Fatalf("expected workspace acme-anvil-main in menu, got %q", call)
	}
	if !strings.Contains(call, "swift-owlbear") {
		t.Fatalf("expected workspace swift-owlbear in menu, got %q", call)
	}
}

func TestPopupContainsCreateWorktreePipeline(t *testing.T) {
	origBaseDir := baseDir
	defer func() { baseDir = origBaseDir }()
	baseDir = setupPopupWorkspaces(t)

	fake := newFakeExecutor()
	err := runPopupWith(context.Background(), fake, "/usr/local/bin/workspacectl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	call := fake.calls[0]
	if !strings.Contains(call, `read org`) {
		t.Fatalf("expected read org in create worktree pipeline, got %q", call)
	}
	if !strings.Contains(call, "create --type worktree") {
		t.Fatalf("expected create --type worktree in pipeline, got %q", call)
	}
}

func TestPopupContainsCreateTemporaryPipeline(t *testing.T) {
	origBaseDir := baseDir
	defer func() { baseDir = origBaseDir }()
	baseDir = setupPopupWorkspaces(t)

	fake := newFakeExecutor()
	err := runPopupWith(context.Background(), fake, "/usr/local/bin/workspacectl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	call := fake.calls[0]
	if !strings.Contains(call, "create --type temporary") {
		t.Fatalf("expected create --type temporary in pipeline, got %q", call)
	}
}

func TestPopupContainsDeletePipeline(t *testing.T) {
	origBaseDir := baseDir
	defer func() { baseDir = origBaseDir }()
	baseDir = setupPopupWorkspaces(t)

	fake := newFakeExecutor()
	err := runPopupWith(context.Background(), fake, "/usr/local/bin/workspacectl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	call := fake.calls[0]
	if !strings.Contains(call, "fzf") {
		t.Fatalf("expected fzf in delete pipeline, got %q", call)
	}
	if !strings.Contains(call, "delete --force") {
		t.Fatalf("expected delete --force in pipeline, got %q", call)
	}
}

func TestPopupUsesInjectedExePath(t *testing.T) {
	origBaseDir := baseDir
	defer func() { baseDir = origBaseDir }()
	baseDir = setupPopupWorkspaces(t)

	fake := newFakeExecutor()
	exe := "/custom/path/workspacectl"
	err := runPopupWith(context.Background(), fake, exe)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	call := fake.calls[0]
	if !strings.Contains(call, "/custom/path/workspacectl create") {
		t.Fatalf("expected injected exe path in create pipeline, got %q", call)
	}
	if !strings.Contains(call, "/custom/path/workspacectl attach") {
		t.Fatalf("expected injected exe path in attach pipeline, got %q", call)
	}
	if !strings.Contains(call, "/custom/path/workspacectl delete") {
		t.Fatalf("expected injected exe path in delete pipeline, got %q", call)
	}
	if !strings.Contains(call, "/custom/path/workspacectl list") {
		t.Fatalf("expected injected exe path in list pipeline, got %q", call)
	}
}

func TestPopupEmptyWorkspaceList(t *testing.T) {
	origBaseDir := baseDir
	defer func() { baseDir = origBaseDir }()
	baseDir = t.TempDir()

	fake := newFakeExecutor()
	err := runPopupWith(context.Background(), fake, "/usr/local/bin/workspacectl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fake.calls))
	}
	call := fake.calls[0]
	if !strings.Contains(call, "display-menu") {
		t.Fatalf("expected display-menu, got %q", call)
	}
	if !strings.Contains(call, "Create workspace") {
		t.Fatalf("expected Create workspace in menu, got %q", call)
	}
	if !strings.Contains(call, "Delete workspace") {
		t.Fatalf("expected Delete workspace in menu, got %q", call)
	}
}
