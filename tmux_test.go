package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestTmuxSessionName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"org-repo-branch", "workspace-org-repo-branch"},
		{"org-repo-feat.1", "workspace-org-repo-feat-1"},
		{"org-repo-fix/thing", "workspace-org-repo-fix-thing"},
	}
	for _, tt := range tests {
		got := TmuxSessionName(tt.input)
		if got != tt.want {
			t.Errorf("TmuxSessionName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSessionExists(t *testing.T) {
	fake := newFakeExecutor()
	// has-session succeeds means session exists.
	exists, err := sessionExists(context.Background(), fake, "test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("expected session to exist")
	}
}

func TestSessionDoesNotExist(t *testing.T) {
	fake := newFakeExecutor()
	fake.errors["has-session"] = fmt.Errorf("exit status 1")
	exists, err := sessionExists(context.Background(), fake, "test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatal("expected session not to exist")
	}
}

func TestCreateSession(t *testing.T) {
	fake := newFakeExecutor()
	err := createSession(context.Background(), fake, "test-session", "/tmp/work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fake.calls))
	}
	if !strings.Contains(fake.calls[0], "new-session") {
		t.Fatalf("expected new-session, got %q", fake.calls[0])
	}
}

func TestAttachSession(t *testing.T) {
	fake := newFakeExecutor()
	err := attachSession(context.Background(), fake, "test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should try switch-client first.
	if !strings.Contains(fake.calls[0], "switch-client") {
		t.Fatalf("expected switch-client, got %q", fake.calls[0])
	}
}

func TestAttachSessionFallback(t *testing.T) {
	fake := newFakeExecutor()
	fake.errors["switch-client"] = fmt.Errorf("exit status 1")
	err := attachSession(context.Background(), fake, "test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(fake.calls))
	}
	if !strings.Contains(fake.calls[1], "attach-session") {
		t.Fatalf("expected attach-session, got %q", fake.calls[1])
	}
}

func TestCreateWindow(t *testing.T) {
	fake := newFakeExecutor()
	err := createWindow(context.Background(), fake, "test-session", "editor", "vim")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(fake.calls[0], "new-window") {
		t.Fatalf("expected new-window, got %q", fake.calls[0])
	}
}
