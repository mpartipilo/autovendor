package vendor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func init() {
	if !isTTY() {
		colorReset = ""
		colorGreen = ""
		colorYellow = ""
		colorCyan = ""
	}
}

func isTTY() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

// ChangedGoMods returns the list of module directories whose go.mod or go.sum
// changed between two git refs. It runs git diff in the given repoDir.
func ChangedGoMods(repoDir, oldRef, newRef string) ([]string, error) {
	out, err := gitExec(repoDir, "diff", "--name-only", oldRef, newRef)
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	return ParseChangedFiles(out), nil
}

// ParseChangedFiles extracts unique module directories from git diff --name-only output.
// It returns directories containing changed go.mod or go.sum files.
// Root module is represented as empty string "".
func ParseChangedFiles(diffOutput string) []string {
	var mods []string
	seen := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(diffOutput), "\n") {
		if line == "" {
			continue
		}
		base := filepath.Base(line)
		if base != "go.mod" && base != "go.sum" {
			continue
		}
		modDir := filepath.Dir(line)
		if modDir == "." {
			modDir = ""
		}
		if !seen[modDir] {
			seen[modDir] = true
			mods = append(mods, modDir)
		}
	}
	return mods
}

// SyncModules runs `go mod vendor` for each module directory that has a vendor/ dir.
// modDirs are relative to repoDir. Empty string means the repo root.
func SyncModules(repoDir string, modDirs []string) error {
	for _, modDir := range modDirs {
		absDir := repoDir
		displayDir := "."
		if modDir != "" {
			absDir = filepath.Join(repoDir, modDir)
			displayDir = "./" + modDir
		}

		// Check that a vendor/ directory exists for this module
		vendorDir := filepath.Join(absDir, "vendor")
		info, err := os.Stat(vendorDir)
		if err != nil || !info.IsDir() {
			continue
		}

		fmt.Printf("%sautovendor:%s go.mod changed in %s%s%s, running go mod vendor...\n",
			colorCyan, colorReset, colorYellow, displayDir, colorReset)

		cmd := exec.Command("go", "mod", "vendor")
		cmd.Dir = absDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go mod vendor in %s: %w", displayDir, err)
		}

		fmt.Printf("%sautovendor:%s vendor synced in %s %s✓%s\n",
			colorCyan, colorReset, displayDir, colorGreen, colorReset)
	}
	return nil
}

func gitExec(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s: %s", err, string(exitErr.Stderr))
		}
		return "", err
	}
	return string(out), nil
}
