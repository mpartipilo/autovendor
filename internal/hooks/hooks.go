package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	beginMarker = "# autovendor:begin"
	endMarker   = "# autovendor:end"
)

// DetectHooksDir returns the absolute path to the hooks directory for the given repo.
// It checks `git config core.hooksPath` first, falling back to `.git/hooks`.
func DetectHooksDir(repoDir string) (string, error) {
	cmd := exec.Command("git", "config", "core.hooksPath")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err == nil {
		hooksPath := strings.TrimSpace(string(out))
		if hooksPath != "" {
			if filepath.IsAbs(hooksPath) {
				return hooksPath, nil
			}
			return filepath.Join(repoDir, hooksPath), nil
		}
	}

	// Find the .git directory (handles worktrees too)
	cmd = exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoDir
	out, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	gitDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoDir, gitDir)
	}
	return filepath.Join(gitDir, "hooks"), nil
}

// Install installs all autovendor hook scripts into the given hooks directory.
// It creates the directory if it doesn't exist.
// If a hook file already exists, the autovendor block is appended (unless already present).
func Install(hooksDir string) error {
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}

	for _, name := range HookNames {
		hookPath := filepath.Join(hooksDir, name)
		if err := installHook(hookPath, name); err != nil {
			return fmt.Errorf("install %s: %w", name, err)
		}
	}
	return nil
}

// Uninstall removes autovendor blocks from all hooks in the given directory.
// If a hook file contains only the autovendor block, the file is removed entirely.
func Uninstall(hooksDir string) error {
	for _, name := range HookNames {
		hookPath := filepath.Join(hooksDir, name)
		if err := uninstallHook(hookPath); err != nil {
			return fmt.Errorf("uninstall %s: %w", name, err)
		}
	}
	return nil
}

func installHook(hookPath, name string) error {
	tmpl, err := Template(name)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	existing, err := os.ReadFile(hookPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing hook: %w", err)
	}

	content := string(existing)

	// Already installed — skip
	if strings.Contains(content, beginMarker) {
		fmt.Printf("  %s: already installed, skipping\n", name)
		return nil
	}

	if len(existing) == 0 {
		// No existing hook — write the full template
		if err := os.WriteFile(hookPath, tmpl, 0o755); err != nil {
			return err
		}
		fmt.Printf("  %s: installed\n", name)
		return nil
	}

	// Existing hook — extract only the autovendor block from the template and append it
	block := extractBlock(string(tmpl))
	content = strings.TrimRight(content, "\n") + "\n\n" + block + "\n"
	if err := os.WriteFile(hookPath, []byte(content), 0o755); err != nil {
		return err
	}
	fmt.Printf("  %s: appended to existing hook\n", name)
	return nil
}

func uninstallHook(hookPath string) error {
	existing, err := os.ReadFile(hookPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read hook: %w", err)
	}

	content := string(existing)
	if !strings.Contains(content, beginMarker) {
		return nil // Nothing to remove
	}

	cleaned := removeBlock(content)
	cleaned = strings.TrimSpace(cleaned)

	// If only a shebang (or nothing) remains, remove the file
	lines := strings.Split(cleaned, "\n")
	if cleaned == "" || (len(lines) == 1 && strings.HasPrefix(lines[0], "#!")) {
		if err := os.Remove(hookPath); err != nil {
			return err
		}
		fmt.Printf("  %s: removed\n", filepath.Base(hookPath))
		return nil
	}

	if err := os.WriteFile(hookPath, []byte(cleaned+"\n"), 0o755); err != nil {
		return err
	}
	fmt.Printf("  %s: autovendor block removed\n", filepath.Base(hookPath))
	return nil
}

// extractBlock returns the text between (and including) the begin/end markers.
func extractBlock(content string) string {
	startIdx := strings.Index(content, beginMarker)
	endIdx := strings.Index(content, endMarker)
	if startIdx < 0 || endIdx < 0 {
		return content
	}
	return content[startIdx : endIdx+len(endMarker)]
}

// removeBlock removes the text between (and including) the begin/end markers,
// plus any surrounding blank lines.
func removeBlock(content string) string {
	startIdx := strings.Index(content, beginMarker)
	endIdx := strings.Index(content, endMarker)
	if startIdx < 0 || endIdx < 0 {
		return content
	}
	before := content[:startIdx]
	after := content[endIdx+len(endMarker):]

	// Clean up extra blank lines
	before = strings.TrimRight(before, "\n")
	after = strings.TrimLeft(after, "\n")

	if before == "" {
		return after
	}
	if after == "" {
		return before
	}
	return before + "\n" + after
}
