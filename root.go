package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	baseDir string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:           "workspacectl",
	Short:         "Manage developer workspaces",
	Long:          "A CLI tool for managing developer workspaces. Automates git worktrees, tmux sessions, and Claude Code environments.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	defaultBaseDir := filepath.Join(home, ".workspacectl")
	rootCmd.PersistentFlags().StringVar(&baseDir, "base-dir", defaultBaseDir, "base directory for workspaces")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
