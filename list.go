package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(_ *cobra.Command, _ []string) error {
	return runListWith(baseDir)
}

func runListWith(base string) error {
	workspaces, err := Discover(base)
	if err != nil {
		return fmt.Errorf("discovering workspaces: %w", err)
	}

	for _, ws := range workspaces {
		fmt.Println(ws.Name)
	}
	return nil
}
