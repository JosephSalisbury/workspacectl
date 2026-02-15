package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

type fakeExecutor struct {
	calls   []string
	results map[string]string
	errors  map[string]error
}

func newFakeExecutor() *fakeExecutor {
	return &fakeExecutor{
		results: make(map[string]string),
		errors:  make(map[string]error),
	}
}

func (f *fakeExecutor) Run(_ context.Context, name string, args ...string) (string, error) {
	call := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, call)
	for pattern, err := range f.errors {
		if strings.Contains(call, pattern) {
			return "", err
		}
	}
	for pattern, result := range f.results {
		if strings.Contains(call, pattern) {
			return result, nil
		}
	}
	return "", nil
}

func (f *fakeExecutor) RunAttached(_ context.Context, name string, args ...string) error {
	call := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, call)
	for pattern, err := range f.errors {
		if strings.Contains(call, pattern) {
			return err
		}
	}
	return nil
}

func TestBareClone(t *testing.T) {
	fake := newFakeExecutor()
	err := bareClone(context.Background(), fake, "git@github.com:org/repo.git", "/tmp/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fake.calls))
	}
	if !strings.Contains(fake.calls[0], "clone --bare") {
		t.Fatalf("expected bare clone, got %q", fake.calls[0])
	}
}

func TestBareCloneError(t *testing.T) {
	fake := newFakeExecutor()
	fake.errors["clone"] = fmt.Errorf("clone failed")
	err := bareClone(context.Background(), fake, "git@github.com:org/repo.git", "/tmp/repo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDefaultBranch(t *testing.T) {
	fake := newFakeExecutor()
	fake.results["symbolic-ref"] = "refs/heads/main"
	branch, err := defaultBranch(context.Background(), fake, "/tmp/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Fatalf("got %q, want %q", branch, "main")
	}
}

func TestCreateWorktree(t *testing.T) {
	fake := newFakeExecutor()
	err := createWorktree(context.Background(), fake, "/tmp/repo", "feature", "/tmp/repo/feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(fake.calls[0], "worktree add") {
		t.Fatalf("expected worktree add, got %q", fake.calls[0])
	}
}

func TestBranchExistsOnRemote(t *testing.T) {
	fake := newFakeExecutor()
	fake.results["ls-remote"] = "abc123\trefs/heads/main"
	exists, err := branchExistsOnRemote(context.Background(), fake, "/tmp/repo", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("expected branch to exist")
	}
}

func TestBranchDoesNotExistOnRemote(t *testing.T) {
	fake := newFakeExecutor()
	fake.results["ls-remote"] = ""
	exists, err := branchExistsOnRemote(context.Background(), fake, "/tmp/repo", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatal("expected branch not to exist")
	}
}

func TestCreateBranch(t *testing.T) {
	fake := newFakeExecutor()
	err := gitCreateBranch(context.Background(), fake, "/tmp/repo", "feature", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(fake.calls[0], "branch feature main") {
		t.Fatalf("expected branch create, got %q", fake.calls[0])
	}
}
