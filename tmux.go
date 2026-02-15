package main

import (
	"context"
	"fmt"
	"strings"
)

// TmuxSessionName derives a tmux session name from a workspace name.
func TmuxSessionName(workspaceName string) string {
	replacer := strings.NewReplacer(
		".", "-",
		":", "-",
		"/", "-",
	)
	return "workspace-" + replacer.Replace(workspaceName)
}

func sessionExists(ctx context.Context, executor Executor, name string) (bool, error) {
	_, err := executor.Run(ctx, "tmux", "has-session", "-t", name)
	if err != nil {
		// tmux has-session exits non-zero if session doesn't exist.
		if strings.Contains(err.Error(), "exit status") {
			return false, nil
		}
		return false, fmt.Errorf("checking tmux session %s: %w", name, err)
	}
	return true, nil
}

func createSession(ctx context.Context, executor Executor, name, workDir string) error {
	_, err := executor.Run(ctx, "tmux", "new-session", "-d", "-s", name, "-c", workDir)
	if err != nil {
		return fmt.Errorf("creating tmux session %s: %w", name, err)
	}
	return nil
}

func attachSession(ctx context.Context, executor Executor, name string) error {
	// If we're inside tmux, switch; otherwise attach.
	_, err := executor.Run(ctx, "tmux", "switch-client", "-t", name)
	if err != nil {
		// Not inside tmux, try attach.
		_, err = executor.Run(ctx, "tmux", "attach-session", "-t", name)
		if err != nil {
			return fmt.Errorf("attaching to tmux session %s: %w", name, err)
		}
	}
	return nil
}

func createWindow(ctx context.Context, executor Executor, sessionName, windowName, command string) error {
	_, err := executor.Run(ctx, "tmux", "new-window", "-t", sessionName, "-n", windowName, command)
	if err != nil {
		return fmt.Errorf("creating tmux window %s in %s: %w", windowName, sessionName, err)
	}
	return nil
}
