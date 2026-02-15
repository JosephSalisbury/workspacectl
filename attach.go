package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach <workspace-name>",
	Short: "Attach to a workspace's tmux session",
	Args:  cobra.ExactArgs(1),
	RunE:  runAttach,
}

func init() {
	rootCmd.AddCommand(attachCmd)
}

func runAttach(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	executor := &DefaultExecutor{}
	return runAttachWith(ctx, executor, args[0])
}

func runAttachWith(ctx context.Context, executor Executor, name string) error {
	workspaces, err := Discover(baseDir)
	if err != nil {
		return fmt.Errorf("discovering workspaces: %w", err)
	}

	var ws *Workspace
	for i := range workspaces {
		if workspaces[i].Name == name {
			ws = &workspaces[i]
			break
		}
	}
	if ws == nil {
		return fmt.Errorf("workspace %q not found", name)
	}

	cfg, err := EnsureConfig(baseDir)
	if err != nil {
		return err
	}

	sessionName := TmuxSessionName(ws.Name)

	exists, err := sessionExists(ctx, executor, sessionName)
	if err != nil {
		return err
	}

	if !exists {
		if err := createSession(ctx, executor, sessionName, ws.Path); err != nil {
			return err
		}

		// Create layout windows.
		layoutKey := string(ws.Type)
		windows, ok := cfg.Layouts[layoutKey]
		if ok {
			// The first window is the default window created with the session.
			// Rename it and run the command.
			if len(windows) > 0 {
				if _, err := executor.Run(ctx, "tmux", "rename-window", "-t", sessionName, windows[0].Name); err != nil {
					return fmt.Errorf("renaming first window: %w", err)
				}
				if _, err := executor.Run(ctx, "tmux", "send-keys", "-t", sessionName, windows[0].Command, "Enter"); err != nil {
					return fmt.Errorf("sending command to first window: %w", err)
				}
			}
			// Create additional windows.
			for _, w := range windows[1:] {
				if err := createWindow(ctx, executor, sessionName, w.Name, w.Command); err != nil {
					return err
				}
			}
		}
	}

	return attachSession(ctx, executor, sessionName)
}
