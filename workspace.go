package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// WorkspaceType distinguishes between workspace types.
type WorkspaceType string

const (
	// WorkspaceTypeWorktree is a git worktree workspace.
	WorkspaceTypeWorktree WorkspaceType = "worktree"
	// WorkspaceTypeTemporary is a temporary workspace.
	WorkspaceTypeTemporary WorkspaceType = "temporary"
)

// Workspace represents a discovered workspace on disk.
type Workspace struct {
	Name   string
	Type   WorkspaceType
	Path   string
	Org    string
	Repo   string
	Branch string
}

// WorkspaceName computes the workspace name for a worktree workspace.
func WorkspaceName(org, repo, branch string) string {
	return fmt.Sprintf("%s-%s-%s", org, repo, branch)
}

// Discover walks the base directory and returns all discovered workspaces.
func Discover(baseDir string) ([]Workspace, error) {
	var workspaces []Workspace

	// Discover worktree workspaces: repositories/<org>/<repo>/<branch>/
	reposDir := filepath.Join(baseDir, "repositories")
	if info, err := os.Stat(reposDir); err == nil && info.IsDir() {
		orgs, err := os.ReadDir(reposDir)
		if err != nil {
			return nil, fmt.Errorf("reading repositories dir: %w", err)
		}
		for _, org := range orgs {
			if !org.IsDir() {
				continue
			}
			orgDir := filepath.Join(reposDir, org.Name())
			repos, err := os.ReadDir(orgDir)
			if err != nil {
				return nil, fmt.Errorf("reading org dir %s: %w", org.Name(), err)
			}
			for _, repo := range repos {
				if !repo.IsDir() {
					continue
				}
				repoDir := filepath.Join(orgDir, repo.Name())
				branches, err := os.ReadDir(repoDir)
				if err != nil {
					return nil, fmt.Errorf("reading repo dir %s/%s: %w", org.Name(), repo.Name(), err)
				}
				for _, branch := range branches {
					if !branch.IsDir() {
						continue
					}
					// A git worktree has a .git file (not directory) in it.
					// Bare clone internal dirs (hooks, info, objects, refs) don't.
					gitPath := filepath.Join(repoDir, branch.Name(), ".git")
					info, err := os.Stat(gitPath)
					if err != nil || info.IsDir() {
						continue
					}
					workspaces = append(workspaces, Workspace{
						Name:   WorkspaceName(org.Name(), repo.Name(), branch.Name()),
						Type:   WorkspaceTypeWorktree,
						Path:   filepath.Join(repoDir, branch.Name()),
						Org:    org.Name(),
						Repo:   repo.Name(),
						Branch: branch.Name(),
					})
				}
			}
		}
	}

	// Discover temporary workspaces: temporary/<name>/
	tempDir := filepath.Join(baseDir, "temporary")
	if info, err := os.Stat(tempDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(tempDir)
		if err != nil {
			return nil, fmt.Errorf("reading temporary dir: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			workspaces = append(workspaces, Workspace{
				Name: entry.Name(),
				Type: WorkspaceTypeTemporary,
				Path: filepath.Join(tempDir, entry.Name()),
			})
		}
	}

	return workspaces, nil
}
