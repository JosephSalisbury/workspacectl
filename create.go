package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	createOrg    string
	createRepo   string
	createBranch string
	createType   string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new workspace",
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringVar(&createOrg, "org", "", "GitHub organization or user")
	createCmd.Flags().StringVar(&createRepo, "repo", "", "repository name")
	createCmd.Flags().StringVar(&createBranch, "branch", "", "branch name (auto-generated if not specified)")
	createCmd.Flags().StringVar(&createType, "type", "worktree", "workspace type (worktree or temporary)")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	executor := &DefaultExecutor{}

	switch createType {
	case "worktree":
		return runCreateWorktree(ctx, executor)
	case "temporary":
		return runCreateTemporary()
	default:
		return fmt.Errorf("unknown workspace type: %s", createType)
	}
}

func runCreateWorktree(ctx context.Context, executor Executor) error {
	if createOrg == "" {
		return fmt.Errorf("--org is required for worktree workspaces")
	}
	if createRepo == "" {
		return fmt.Errorf("--repo is required for worktree workspaces")
	}

	if createBranch == "" {
		createBranch = GenerateName()
	}

	cfg, err := EnsureConfig(baseDir)
	if err != nil {
		return err
	}
	_ = cfg

	repoDir := filepath.Join(baseDir, "repositories", createOrg, createRepo)

	// Bare clone if repo doesn't exist.
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(repoDir), 0o755); err != nil {
			return fmt.Errorf("creating repo parent dir: %w", err)
		}
		url := fmt.Sprintf("git@github.com:%s/%s.git", createOrg, createRepo)
		if err := bareClone(ctx, executor, url, repoDir); err != nil {
			return err
		}
	}

	// Check if branch exists on remote.
	exists, err := branchExistsOnRemote(ctx, executor, repoDir, createBranch)
	if err != nil {
		return err
	}

	// Create branch if it doesn't exist on remote.
	if !exists {
		defBranch, err := defaultBranch(ctx, executor, repoDir)
		if err != nil {
			return err
		}
		if err := gitCreateBranch(ctx, executor, repoDir, createBranch, defBranch); err != nil {
			return err
		}
	}

	// Create worktree.
	worktreeDir := filepath.Join(repoDir, createBranch)
	if err := createWorktree(ctx, executor, repoDir, createBranch, worktreeDir); err != nil {
		return err
	}

	name := WorkspaceName(createOrg, createRepo, createBranch)
	fmt.Println(name)
	return nil
}

func runCreateTemporary() error {
	name := createBranch
	if name == "" {
		name = GenerateName()
	}

	if _, err := EnsureConfig(baseDir); err != nil {
		return err
	}

	tempDir := filepath.Join(baseDir, "temporary", name)
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return fmt.Errorf("creating temporary workspace: %w", err)
	}

	fmt.Println(name)
	return nil
}
