package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mpartipilo/autovendor/internal/hooks"
)

// Install handles the `autovendor install [path]` command.
func Install(args []string) error {
	repoDir := "."
	if len(args) > 0 {
		repoDir = args[0]
	}

	absDir, err := filepath.Abs(repoDir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	hooksDir, err := hooks.DetectHooksDir(absDir)
	if err != nil {
		return err
	}

	fmt.Printf("Installing autovendor hooks into %s\n", hooksDir)
	if err := hooks.Install(hooksDir, Version); err != nil {
		return err
	}

	if Version != "" && Version != "dev" {
		fmt.Printf("\nDone! Vendor will auto-sync after pull, checkout, and rebase. (pinned to v%s)\n", Version)
	} else {
		fmt.Println("\nDone! Vendor will auto-sync after pull, checkout, and rebase.")
	}
	return nil
}
