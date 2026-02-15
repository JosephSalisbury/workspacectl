package main

import (
	"context"
	"fmt"
	"os"
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
	if os.Getenv("TMUX") != "" {
		if _, err := executor.Run(ctx, "tmux", "switch-client", "-t", name); err != nil {
			return fmt.Errorf("switching to tmux session %s: %w", name, err)
		}
		return nil
	}
	return executor.RunAttached(ctx, "tmux", "attach-session", "-t", name)
}

func createPane(ctx context.Context, executor Executor, sessionName, workDir, command string) error {
	paneID, err := executor.Run(ctx, "tmux", "split-window", "-h", "-d", "-t", sessionName, "-c", workDir, "-P", "-F", "#{pane_id}")
	if err != nil {
		return fmt.Errorf("creating tmux pane in %s: %w", sessionName, err)
	}
	if _, err := executor.Run(ctx, "tmux", "send-keys", "-t", strings.TrimSpace(paneID), command, "Enter"); err != nil {
		return fmt.Errorf("sending command to pane in %s: %w", sessionName, err)
	}
	return nil
}
