package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var popupCmd = &cobra.Command{
	Use:   "popup",
	Short: "Show workspace quick-switcher menu via tmux display-menu",
	RunE:  runPopup,
}

func init() {
	rootCmd.AddCommand(popupCmd)
}

func runPopup(cmd *cobra.Command, _ []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return runPopupWith(cmd.Context(), &DefaultExecutor{}, exe)
}

func runPopupWith(ctx context.Context, executor Executor, exe string) error {
	workspaces, err := Discover(baseDir)
	if err != nil {
		return err
	}

	createWorktreePipeline := `printf "org: " && read org && printf "repo: " && read repo && name=$(` + exe + ` create --type worktree --org "$org" --repo "$repo") && ` + exe + ` attach "$name"`
	createTemporaryPipeline := `name=$(` + exe + ` create --type temporary) && ` + exe + ` attach "$name"`
	deletePipeline := exe + ` list | fzf | xargs ` + exe + ` delete --force`

	args := []string{
		"display-menu", "-T", "Workspaces",
		"Create workspace (worktree)", "c", "display-popup -E '" + createWorktreePipeline + "'",
		"Create workspace (temporary)", "t", "display-popup -E '" + createTemporaryPipeline + "'",
		"",
	}

	for _, ws := range workspaces {
		args = append(args, ws.Name, "", "display-popup -E '"+exe+" attach "+ws.Name+"'")
	}

	args = append(args, "")
	args = append(args, "Delete workspace", "d", "display-popup -E '"+deletePipeline+"'")

	_, err = executor.Run(ctx, "tmux", args...)
	return err
}
