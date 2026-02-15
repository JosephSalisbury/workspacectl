package main

import (
	"context"
	"testing"
)

func TestDefaultExecutorRun(t *testing.T) {
	e := &DefaultExecutor{}
	out, err := e.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello" {
		t.Fatalf("got %q, want %q", out, "hello")
	}
}

func TestDefaultExecutorRunError(t *testing.T) {
	e := &DefaultExecutor{}
	_, err := e.Run(context.Background(), "false")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
