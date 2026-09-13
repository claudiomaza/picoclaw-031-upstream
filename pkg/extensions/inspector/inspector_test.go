package inspector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverUsesRuntimeRoot(t *testing.T) {
	xs := Discover("/opt/picoclaw")
	if xs[0].Path != "/opt/picoclaw/bin/picoclaw" {
		t.Fatal(xs[0].Path)
	}
	if !allowedNames[xs[0].Name] {
		t.Fatal("unexpected target")
	}
}
func TestInspectDoesNotFollowSymlink(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "x"), []byte("x"), 0600)
	os.Symlink(filepath.Join(d, "x"), filepath.Join(d, "link"))
	r := inspectTarget(Target{Name: "x", Path: filepath.Join(d, "link"), Kind: "binary"})
	if r.Error == "" {
		t.Fatal("expected symlink rejection")
	}
}
