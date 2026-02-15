package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Executor runs external commands and returns their combined stdout.
type Executor interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
	RunAttached(ctx context.Context, name string, args ...string) error
}

// DefaultExecutor shells out using os/exec.
type DefaultExecutor struct {
	Dir string
}

// Run executes a command and returns trimmed stdout.
func (e *DefaultExecutor) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if e.Dir != "" {
		cmd.Dir = e.Dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running %s %s: %w: %s", name, strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// RunAttached replaces the current process with the given command.
func (e *DefaultExecutor) RunAttached(_ context.Context, name string, args ...string) error {
	binary, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("finding %s: %w", name, err)
	}
	return syscall.Exec(binary, append([]string{name}, args...), os.Environ())
}
