package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <workspace-name>",
	Short: "Delete a workspace",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "skip confirmation even if there are uncommitted changes or unpushed commits")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	return runDeleteWith(cmd.Context(), &DefaultExecutor{}, os.Stdin, args[0], deleteForce)
}

func runDeleteWith(ctx context.Context, executor Executor, stdin io.Reader, name string, force bool) error {
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

	sessionName := TmuxSessionName(ws.Name)
	logVerbose("tmux session name: %s", sessionName)

	switch ws.Type {
	case WorkspaceTypeWorktree:
		repoDir := filepath.Join(baseDir, "repositories", ws.Org, ws.Repo)
		orgDir := filepath.Join(baseDir, "repositories", ws.Org)

		uncommitted, err := hasUncommittedChanges(ctx, executor, ws.Path)
		if err != nil {
			return err
		}
		unpushed, err := hasUnpushedCommits(ctx, executor, repoDir, ws.Branch, ws.Path)
		if err != nil {
			return err
		}

		if (uncommitted || unpushed) && !force {
			if uncommitted {
				fmt.Fprintln(os.Stderr, "Warning: workspace has uncommitted changes")
			}
			if unpushed {
				fmt.Fprintln(os.Stderr, "Warning: workspace has unpushed commits")
			}
			fmt.Fprintf(os.Stderr, "Delete workspace %q? [y/N]: ", name)
			scanner := bufio.NewScanner(stdin)
			scanner.Scan()
			response := strings.TrimSpace(scanner.Text())
			if response != "y" {
				return fmt.Errorf("aborted")
			}
		}

		isLast, err := isLastWorktree(repoDir)
		if err != nil {
			return err
		}

		if err := killSession(ctx, executor, sessionName); err != nil {
			return err
		}

		if err := removeWorktree(ctx, executor, repoDir, ws.Path, uncommitted); err != nil {
			return err
		}

		if isLast {
			logVerbose("last worktree in bare clone, removing %s", repoDir)
			if err := os.RemoveAll(repoDir); err != nil {
				return fmt.Errorf("removing bare clone: %w", err)
			}
			entries, err := os.ReadDir(orgDir)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("reading org dir: %w", err)
			}
			if len(entries) == 0 {
				logVerbose("org dir is empty, removing %s", orgDir)
				if err := os.RemoveAll(orgDir); err != nil {
					return fmt.Errorf("removing org dir: %w", err)
				}
			}
		}

	case WorkspaceTypeTemporary:
		if err := killSession(ctx, executor, sessionName); err != nil {
			return err
		}
		if err := os.RemoveAll(ws.Path); err != nil {
			return fmt.Errorf("removing temporary workspace: %w", err)
		}
	}

	return nil
}

// isLastWorktree returns true if there is exactly one worktree in repoDir.
func isLastWorktree(repoDir string) (bool, error) {
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return false, fmt.Errorf("reading repo dir: %w", err)
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		gitPath := filepath.Join(repoDir, entry.Name(), ".git")
		info, err := os.Stat(gitPath)
		if err != nil || info.IsDir() {
			continue
		}
		count++
	}
	return count == 1, nil
}
