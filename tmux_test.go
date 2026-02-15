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

func TestAttachSessionOutsideTmux(t *testing.T) {
	t.Setenv("TMUX", "")
	fake := newFakeExecutor()
	err := attachSession(context.Background(), fake, "test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fake.calls))
	}
	if !strings.Contains(fake.calls[0], "attach-session") {
		t.Fatalf("expected attach-session, got %q", fake.calls[0])
	}
}

func TestAttachSessionInsideTmux(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	fake := newFakeExecutor()
	err := attachSession(context.Background(), fake, "test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fake.calls))
	}
	if !strings.Contains(fake.calls[0], "switch-client") {
		t.Fatalf("expected switch-client, got %q", fake.calls[0])
	}
}

func TestCreatePane(t *testing.T) {
	fake := newFakeExecutor()
	err := createPane(context.Background(), fake, "test-session", "/tmp/work", "vim")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(fake.calls))
	}
	if !strings.Contains(fake.calls[0], "split-window") {
		t.Fatalf("expected split-window, got %q", fake.calls[0])
	}
	if !strings.Contains(fake.calls[1], "send-keys") || !strings.Contains(fake.calls[1], "vim") {
		t.Fatalf("expected send-keys with vim, got %q", fake.calls[1])
	}
}
