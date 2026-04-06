package cmd

import (
	"fmt"
	"os"

	"github.com/mpartipilo/autovendor/internal/vendorsync"
)

// Run handles the `autovendor run <hook-type> [args]` command.
// This is called by the git hooks.
func Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: autovendor run <post-merge|post-checkout|post-rewrite> [args]")
	}

	hookType := args[0]
	hookArgs := args[1:]

	repoDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	oldRef, newRef, err := resolveRefs(hookType, hookArgs)
	if err != nil {
		return err
	}

	// Empty refs means we should skip (e.g., file checkout, not branch checkout)
	if oldRef == "" || newRef == "" {
		return nil
	}

	mods, err := vendorsync.ChangedGoMods(repoDir, oldRef, newRef)
	if err != nil {
		return err
	}

	if len(mods) == 0 {
		return nil
	}

	return vendorsync.SyncModules(repoDir, mods)
}

// resolveRefs determines the old and new git refs to compare based on the hook type.
// Returns empty strings if the hook invocation should be skipped.
func resolveRefs(hookType string, args []string) (oldRef, newRef string, err error) {
	switch hookType {
	case "post-merge":
		// post-merge receives one arg: is-squash-merge (0 or 1)
		// Compare ORIG_HEAD (before merge) with HEAD (after merge)
		return "ORIG_HEAD", "HEAD", nil

	case "post-checkout":
		// post-checkout receives: <prev-HEAD> <new-HEAD> <branch-flag>
		// branch-flag: 1 = branch checkout, 0 = file checkout
		if len(args) < 3 {
			return "", "", fmt.Errorf("post-checkout: expected 3 args, got %d", len(args))
		}
		// Skip file checkouts — only branch switches matter
		if args[2] != "1" {
			return "", "", nil
		}
		return args[0], args[1], nil

	case "post-rewrite":
		// post-rewrite receives: <command> (rebase or amend)
		// Skip amend — it doesn't change working tree deps relative to what's vendored
		if len(args) > 0 && args[0] == "amend" {
			return "", "", nil
		}
		// For rebase, compare previous HEAD with current HEAD
		return "HEAD@{1}", "HEAD", nil

	default:
		return "", "", fmt.Errorf("unknown hook type: %s", hookType)
	}
}
