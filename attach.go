package main

import (
	"context"
	"fmt"
	"os"

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

func logVerbose(format string, args ...any) {
	if verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func runAttachWith(ctx context.Context, executor Executor, name string) error {
	logVerbose("discovering workspaces in %s", baseDir)
	workspaces, err := Discover(baseDir)
	if err != nil {
		return fmt.Errorf("discovering workspaces: %w", err)
	}
	logVerbose("found %d workspace(s)", len(workspaces))

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
	logVerbose("workspace: name=%s type=%s path=%s", ws.Name, ws.Type, ws.Path)

	cfg, err := EnsureConfig(baseDir)
	if err != nil {
		return err
	}

	sessionName := TmuxSessionName(ws.Name)
	logVerbose("tmux session name: %s", sessionName)

	exists, err := sessionExists(ctx, executor, sessionName)
	if err != nil {
		return err
	}
	logVerbose("session exists: %v", exists)

	if !exists {
		logVerbose("creating session %s in %s", sessionName, ws.Path)
		if err := createSession(ctx, executor, sessionName, ws.Path); err != nil {
			return err
		}

		// Create layout windows.
		layoutKey := string(ws.Type)
		windows, ok := cfg.Layouts[layoutKey]
		logVerbose("layout %q: found=%v windows=%d", layoutKey, ok, len(windows))
		if ok {
			// The first window is the default window created with the session.
			// Rename it and run the command.
			if len(windows) > 0 {
				logVerbose("renaming first window to %q, command=%q", windows[0].Name, windows[0].Command)
				if _, err := executor.Run(ctx, "tmux", "rename-window", "-t", sessionName, windows[0].Name); err != nil {
					return fmt.Errorf("renaming first window: %w", err)
				}
				if _, err := executor.Run(ctx, "tmux", "send-keys", "-t", sessionName, windows[0].Command, "Enter"); err != nil {
					return fmt.Errorf("sending command to first window: %w", err)
				}
			}
			// Create additional panes.
			for _, w := range windows[1:] {
				logVerbose("creating pane: command=%q", w.Command)
				if err := createPane(ctx, executor, sessionName, ws.Path, w.Command); err != nil {
					return err
				}
			}
		}
	}

	logVerbose("TMUX env: %q", os.Getenv("TMUX"))
	logVerbose("attaching to session %s", sessionName)
	return attachSession(ctx, executor, sessionName)
}
