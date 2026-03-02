package main

import (
	"context"
	"fmt"
	"strings"
)

func bareClone(ctx context.Context, executor Executor, url, dest string) error {
	_, err := executor.Run(ctx, "git", "clone", "--bare", url, dest)
	if err != nil {
		return fmt.Errorf("bare cloning %s: %w", url, err)
	}
	return nil
}

func defaultBranch(ctx context.Context, executor Executor, bareCloneDir string) (string, error) {
	// In a bare clone, HEAD points to the default branch directly.
	out, err := executor.Run(ctx, "git", "-C", bareCloneDir, "symbolic-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("getting default branch: %w", err)
	}
	// Output is like "refs/heads/main"
	parts := strings.Split(out, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("unexpected symbolic-ref output: %s", out)
	}
	return parts[len(parts)-1], nil
}

func createWorktree(ctx context.Context, executor Executor, bareCloneDir, branch, dest string) error {
	_, err := executor.Run(ctx, "git", "-C", bareCloneDir, "worktree", "add", dest, branch)
	if err != nil {
		return fmt.Errorf("creating worktree for %s: %w", branch, err)
	}
	return nil
}

func branchExistsOnRemote(ctx context.Context, executor Executor, bareCloneDir, branch string) (bool, error) {
	out, err := executor.Run(ctx, "git", "-C", bareCloneDir, "ls-remote", "--heads", "origin", branch)
	if err != nil {
		return false, fmt.Errorf("checking remote branch %s: %w", branch, err)
	}
	return strings.TrimSpace(out) != "", nil
}

func gitCreateBranch(ctx context.Context, executor Executor, bareCloneDir, branch, startPoint string) error {
	_, err := executor.Run(ctx, "git", "-C", bareCloneDir, "branch", branch, startPoint)
	if err != nil {
		return fmt.Errorf("creating branch %s: %w", branch, err)
	}
	return nil
}

func hasUncommittedChanges(ctx context.Context, executor Executor, worktreeDir string) (bool, error) {
	out, err := executor.Run(ctx, "git", "-C", worktreeDir, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("checking uncommitted changes: %w", err)
	}
	return strings.TrimSpace(out) != "", nil
}

func hasUnpushedCommits(ctx context.Context, executor Executor, bareCloneDir, branch, worktreeDir string) (bool, error) {
	onRemote, err := branchExistsOnRemote(ctx, executor, bareCloneDir, branch)
	if err != nil {
		return false, err
	}
	if !onRemote {
		return true, nil
	}
	out, err := executor.Run(ctx, "git", "-C", worktreeDir, "log", "origin/"+branch+"..HEAD", "--oneline")
	if err != nil {
		return false, fmt.Errorf("checking unpushed commits: %w", err)
	}
	return strings.TrimSpace(out) != "", nil
}

func removeWorktree(ctx context.Context, executor Executor, bareCloneDir, worktreePath string, force bool) error {
	args := []string{"-C", bareCloneDir, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)
	_, err := executor.Run(ctx, "git", args...)
	if err != nil {
		return fmt.Errorf("removing worktree %s: %w", worktreePath, err)
	}
	return nil
}
