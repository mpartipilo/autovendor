package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {
	for _, name := range HookNames {
		t.Run(name, func(t *testing.T) {
			data, err := Template(name)
			if err != nil {
				t.Fatalf("Template(%q) error: %v", name, err)
			}
			content := string(data)
			if !strings.HasPrefix(content, "#!/bin/sh") {
				t.Error("template should start with shebang")
			}
			if !strings.Contains(content, beginMarker) {
				t.Error("template should contain begin marker")
			}
			if !strings.Contains(content, endMarker) {
				t.Error("template should contain end marker")
			}
			if !strings.Contains(content, "autovendor run "+name) {
				t.Errorf("template should call 'autovendor run %s'", name)
			}
		})
	}
}

func TestInstallFreshDirectory(t *testing.T) {
	dir := t.TempDir()

	if err := Install(dir, "dev"); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	for _, name := range HookNames {
		hookPath := filepath.Join(dir, name)
		info, err := os.Stat(hookPath)
		if err != nil {
			t.Fatalf("hook %s not found: %v", name, err)
		}
		// Check executable
		if info.Mode()&0o111 == 0 {
			t.Errorf("hook %s is not executable", name)
		}
		content, _ := os.ReadFile(hookPath)
		if !strings.Contains(string(content), beginMarker) {
			t.Errorf("hook %s missing begin marker", name)
		}
	}
}

func TestInstallIdempotent(t *testing.T) {
	dir := t.TempDir()

	if err := Install(dir, "dev"); err != nil {
		t.Fatalf("first Install() error: %v", err)
	}
	if err := Install(dir, "dev"); err != nil {
		t.Fatalf("second Install() error: %v", err)
	}

	// Should only have one autovendor block
	content, _ := os.ReadFile(filepath.Join(dir, "post-merge"))
	count := strings.Count(string(content), beginMarker)
	if count != 1 {
		t.Errorf("expected 1 begin marker, got %d", count)
	}
}

func TestInstallAppendsToExistingHook(t *testing.T) {
	dir := t.TempDir()
	hookPath := filepath.Join(dir, "post-merge")

	existing := "#!/bin/sh\necho 'existing hook'\n"
	if err := os.WriteFile(hookPath, []byte(existing), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Install(dir, "dev"); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	content, _ := os.ReadFile(hookPath)
	s := string(content)
	if !strings.Contains(s, "existing hook") {
		t.Error("existing hook content was lost")
	}
	if !strings.Contains(s, beginMarker) {
		t.Error("autovendor block not appended")
	}
}

func TestUninstallRemovesFile(t *testing.T) {
	dir := t.TempDir()

	if err := Install(dir, "dev"); err != nil {
		t.Fatal(err)
	}
	if err := Uninstall(dir); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	for _, name := range HookNames {
		hookPath := filepath.Join(dir, name)
		if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
			t.Errorf("hook %s should have been removed", name)
		}
	}
}

func TestUninstallPreservesExistingHook(t *testing.T) {
	dir := t.TempDir()
	hookPath := filepath.Join(dir, "post-merge")

	existing := "#!/bin/sh\necho 'existing hook'\n"
	if err := os.WriteFile(hookPath, []byte(existing), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Install(dir, "dev"); err != nil {
		t.Fatal(err)
	}
	if err := Uninstall(dir); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook file should still exist: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "existing hook") {
		t.Error("existing hook content was lost after uninstall")
	}
	if strings.Contains(s, beginMarker) {
		t.Error("autovendor block should have been removed")
	}
}

func TestExtractBlock(t *testing.T) {
	content := "before\n# autovendor:begin\nstuff\n# autovendor:end\nafter\n"
	block := extractBlock(content)
	if !strings.HasPrefix(block, beginMarker) {
		t.Errorf("block should start with begin marker, got: %q", block)
	}
	if !strings.HasSuffix(block, endMarker) {
		t.Errorf("block should end with end marker, got: %q", block)
	}
}

func TestRemoveBlock(t *testing.T) {
	content := "#!/bin/sh\necho 'keep'\n\n# autovendor:begin\nstuff\n# autovendor:end\n\necho 'also keep'\n"
	result := removeBlock(content)
	if strings.Contains(result, beginMarker) {
		t.Error("should not contain begin marker")
	}
	if !strings.Contains(result, "keep") {
		t.Error("should preserve non-autovendor content")
	}
	if !strings.Contains(result, "also keep") {
		t.Error("should preserve content after block")
	}
}

func TestUninstallRemovesBashShebangOnly(t *testing.T) {
	dir := t.TempDir()
	hookPath := filepath.Join(dir, "post-merge")

	// Write a hook with #!/bin/bash shebang + autovendor block
	content := "#!/bin/bash\n# autovendor:begin\nautovendor run post-merge \"$@\"\n# autovendor:end\n"
	if err := os.WriteFile(hookPath, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Uninstall(dir); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	// File should be fully removed (only a #!/bin/bash shebang would remain)
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		remaining, _ := os.ReadFile(hookPath)
		t.Errorf("hook should have been removed, but contains: %q", string(remaining))
	}
}

func TestUninstallRemovesEnvShebangOnly(t *testing.T) {
	dir := t.TempDir()
	hookPath := filepath.Join(dir, "post-merge")

	content := "#!/usr/bin/env sh\n# autovendor:begin\nautovendor run post-merge \"$@\"\n# autovendor:end\n"
	if err := os.WriteFile(hookPath, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Uninstall(dir); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		remaining, _ := os.ReadFile(hookPath)
		t.Errorf("hook should have been removed, but contains: %q", string(remaining))
	}
}

func TestUninstallNoHookFile(t *testing.T) {
	dir := t.TempDir()
	// Should not error when hooks don't exist
	if err := Uninstall(dir); err != nil {
		t.Fatalf("Uninstall() on empty dir should not error: %v", err)
	}
}

func TestRemoveBlockNoMarkers(t *testing.T) {
	content := "#!/bin/sh\necho 'hello'\n"
	result := removeBlock(content)
	if result != content {
		t.Errorf("removeBlock with no markers should return content unchanged")
	}
}

func TestExtractBlockNoMarkers(t *testing.T) {
	content := "no markers here"
	result := extractBlock(content)
	if result != content {
		t.Errorf("extractBlock with no markers should return content unchanged")
	}
}

func TestInstallPinsVersion(t *testing.T) {
	dir := t.TempDir()

	if err := Install(dir, "1.2.3"); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "post-merge"))
	s := string(content)
	if !strings.Contains(s, "@v1.2.3") {
		t.Errorf("hook should contain pinned version @v1.2.3, got:\n%s", s)
	}
	if strings.Contains(s, "@latest") {
		t.Error("hook should not contain @latest when version is pinned")
	}
}

func TestInstallDevVersionUsesLatest(t *testing.T) {
	dir := t.TempDir()

	if err := Install(dir, "dev"); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "post-merge"))
	s := string(content)
	if !strings.Contains(s, "@latest") {
		t.Errorf("hook should contain @latest for dev version, got:\n%s", s)
	}
}

func TestInstallVersionWithVPrefix(t *testing.T) {
	dir := t.TempDir()

	if err := Install(dir, "v2.0.0"); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "post-merge"))
	s := string(content)
	if !strings.Contains(s, "@v2.0.0") {
		t.Errorf("hook should contain @v2.0.0, got:\n%s", s)
	}
	// Should not double the v prefix
	if strings.Contains(s, "@vv2.0.0") {
		t.Error("hook should not have double v prefix")
	}
}
