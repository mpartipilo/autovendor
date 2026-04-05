package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mpartipilo/autovendor/internal/hooks"
)

// Uninstall handles the `autovendor uninstall [path]` command.
func Uninstall(args []string) error {
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

	fmt.Printf("Removing autovendor hooks from %s\n", hooksDir)
	if err := hooks.Uninstall(hooksDir); err != nil {
		return err
	}

	fmt.Println("\nDone! Autovendor hooks removed.")
	return nil
}
