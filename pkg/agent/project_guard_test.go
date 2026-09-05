package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectGuardInitializesAndDetectsChanges(t *testing.T) {
	root := t.TempDir()
	g, err := beginProjectGuard(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(g.before) != "" {
		t.Fatalf("new baseline should be clean, got %q", g.before)
	}
	if err := os.WriteFile(filepath.Join(root, "changed.txt"), []byte("changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	after, err := gitOutput(root, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(after, "?? changed.txt") {
		t.Fatalf("change not detected: %q", after)
	}
}

func TestProjectGuardPreservesExistingGitState(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := beginProjectGuard(root); err == nil {
		t.Fatal("expected invalid git directory to fail")
	}
}
