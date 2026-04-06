package vendorsync

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseChangedFiles(t *testing.T) {
	tests := []struct {
		name       string
		diffOutput string
		want       []string
	}{
		{
			name:       "empty output",
			diffOutput: "",
			want:       nil,
		},
		{
			name:       "no go.mod files",
			diffOutput: "main.go\nREADME.md\npkg/util.go\n",
			want:       nil,
		},
		{
			name:       "root go.mod only",
			diffOutput: "go.mod\n",
			want:       []string{""},
		},
		{
			name:       "root go.sum only",
			diffOutput: "go.sum\n",
			want:       []string{""},
		},
		{
			name:       "root go.mod and go.sum deduplicates",
			diffOutput: "go.mod\ngo.sum\n",
			want:       []string{""},
		},
		{
			name:       "nested module",
			diffOutput: "services/auth/go.mod\n",
			want:       []string{"services/auth"},
		},
		{
			name:       "multiple modules (monorepo)",
			diffOutput: "go.mod\ngo.sum\nservices/auth/go.mod\ntools/lint/go.sum\nmain.go\n",
			want:       []string{"", "services/auth", "tools/lint"},
		},
		{
			name:       "deeply nested module",
			diffOutput: "a/b/c/d/go.mod\n",
			want:       []string{"a/b/c/d"},
		},
		{
			name:       "mixed with unrelated files",
			diffOutput: "README.md\ngo.mod\npkg/handler.go\nservices/api/go.sum\n.gitignore\n",
			want:       []string{"", "services/api"},
		},
		{
			name:       "file named gomod is ignored",
			diffOutput: "gomod\ngo.modifier\nvendor/go.mod\n",
			want:       []string{"vendor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseChangedFiles(tt.diffOutput)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseChangedFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncModules_SkipsNonVendored(t *testing.T) {
	dir := t.TempDir()

	// Create a module dir without vendor/ — should be skipped
	modDir := filepath.Join(dir, "novendor")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// This should not error — it should skip the module
	err := SyncModules(dir, []string{"novendor"})
	if err != nil {
		t.Fatalf("SyncModules() should skip non-vendored modules, got error: %v", err)
	}
}

func TestSyncModules_SkipsVendorFile(t *testing.T) {
	dir := t.TempDir()

	// Create a "vendor" that's a file, not a directory
	modDir := filepath.Join(dir, "hasfile")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modDir, "vendor"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := SyncModules(dir, []string{"hasfile"})
	if err != nil {
		t.Fatalf("SyncModules() should skip when vendor is a file, got error: %v", err)
	}
}
