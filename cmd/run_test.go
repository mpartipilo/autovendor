package cmd

import (
	"testing"
)

func TestResolveRefs_PostMerge(t *testing.T) {
	old, new, err := resolveRefs("post-merge", []string{"0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old != "ORIG_HEAD" || new != "HEAD" {
		t.Errorf("got (%q, %q), want (ORIG_HEAD, HEAD)", old, new)
	}
}

func TestResolveRefs_PostCheckout_BranchSwitch(t *testing.T) {
	old, new, err := resolveRefs("post-checkout", []string{"abc123", "def456", "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old != "abc123" || new != "def456" {
		t.Errorf("got (%q, %q), want (abc123, def456)", old, new)
	}
}

func TestResolveRefs_PostCheckout_FileCheckout(t *testing.T) {
	old, new, err := resolveRefs("post-checkout", []string{"abc123", "def456", "0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old != "" || new != "" {
		t.Errorf("file checkout should return empty refs, got (%q, %q)", old, new)
	}
}

func TestResolveRefs_PostCheckout_TooFewArgs(t *testing.T) {
	_, _, err := resolveRefs("post-checkout", []string{"abc123"})
	if err == nil {
		t.Error("expected error for too few args")
	}
}

func TestResolveRefs_PostRewrite_Rebase(t *testing.T) {
	old, new, err := resolveRefs("post-rewrite", []string{"rebase"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old != "HEAD@{1}" || new != "HEAD" {
		t.Errorf("got (%q, %q), want (HEAD@{1}, HEAD)", old, new)
	}
}

func TestResolveRefs_PostRewrite_Amend(t *testing.T) {
	old, new, err := resolveRefs("post-rewrite", []string{"amend"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old != "" || new != "" {
		t.Errorf("amend should return empty refs (skip), got (%q, %q)", old, new)
	}
}

func TestResolveRefs_UnknownHook(t *testing.T) {
	_, _, err := resolveRefs("post-receive", nil)
	if err == nil {
		t.Error("expected error for unknown hook type")
	}
}
