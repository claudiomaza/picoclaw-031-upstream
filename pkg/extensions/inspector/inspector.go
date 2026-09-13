package inspector

import (
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Target struct{ Name, Path, Kind string }
type TargetReport struct {
	Target          Target   `json:"target"`
	Exists          bool     `json:"exists"`
	FileType        string   `json:"file_type,omitempty"`
	Size            int64    `json:"size,omitempty"`
	Mode            string   `json:"mode,omitempty"`
	SHA256          string   `json:"sha256,omitempty"`
	GoVersion       string   `json:"go_version,omitempty"`
	VCSRevision     string   `json:"vcs_revision,omitempty"`
	WhatsAppStrings []string `json:"whatsapp_strings,omitempty"`
	Error           string   `json:"error,omitempty"`
}
type Report struct {
	SchemaVersion string         `json:"schema_version"`
	RuntimeRoot   string         `json:"runtime_root,omitempty"`
	Targets       []TargetReport `json:"targets"`
	Warnings      []string       `json:"warnings,omitempty"`
}

var allowedNames = map[string]bool{"picoclaw": true, "picoclaw-a2a": true, "picoclaw-edit": true, "picoclaw-service": true}

func DefaultRuntimeRoot() string {
	if root := strings.TrimSpace(os.Getenv("PICOCLAW_RUNTIME_ROOT")); root != "" {
		return filepath.Clean(root)
	}
	exe, err := os.Executable()
	if err == nil {
		return filepath.Dir(filepath.Dir(exe))
	}
	return ""
}

// Discover scans only approved local roots and returns evidence candidates. It
// never executes candidates and never reads secret-looking files.
func DiscoverInstalled() []Target {
	seen := map[string]bool{}
	var out []Target
	add := func(name, path, kind string) {
		path = filepath.Clean(path)
		if !seen[path] {
			seen[path] = true
			out = append(out, Target{Name: name, Path: path, Kind: kind})
		}
	}
	if root := DefaultRuntimeRoot(); root != "" {
		add("runtime", root, "directory")
	}
	home, _ := os.UserHomeDir()
	roots := []string{home, "/etc/systemd/system"}
	if home != "" {
		roots = append(roots, filepath.Join(home, ".config", "systemd", "user"))
	}
	for _, base := range roots {
		scan(base, 4, func(path string, info os.FileInfo) {
			n := filepath.Base(path)
			switch {
			case n == "picoclaw" && info.Mode().IsRegular():
				add("picoclaw", path, "binary")
			case n == "picoclaw-a2a-prod" && info.Mode().IsRegular():
				add("picoclaw-a2a", path, "binary")
			case n == "picoclaw-edit" && info.IsDir():
				add("picoclaw-edit", path, "source")
			case n == "picoclaw.service" && info.Mode().IsRegular():
				add("picoclaw-service", path, "unit")
			}
		})
	}
	return out
}

func scan(root string, maxDepth int, visit func(string, os.FileInfo)) {
	if root == "" {
		return
	}
	baseDepth := strings.Count(filepath.Clean(root), string(os.PathSeparator))
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if strings.Count(path, string(os.PathSeparator))-baseDepth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		visit(path, info)
		return nil
	})
}

func Discover(runtimeRoot string) []Target {
	if strings.TrimSpace(runtimeRoot) == "" {
		return DiscoverInstalled()
	}
	runtimeRoot = filepath.Clean(runtimeRoot)
	source := strings.TrimSuffix(runtimeRoot, "-edit") + "-edit"
	return []Target{{Name: "picoclaw", Path: filepath.Join(runtimeRoot, "bin", "picoclaw"), Kind: "binary"}, {Name: "picoclaw-a2a", Path: filepath.Join(runtimeRoot, "picoclaw-a2a-prod"), Kind: "binary"}, {Name: "picoclaw-edit", Path: source, Kind: "source"}, {Name: "picoclaw-service", Path: "/etc/systemd/system/picoclaw.service", Kind: "unit"}}
}

func Inspect(runtimeRoot string) (Report, error) {
	if strings.TrimSpace(runtimeRoot) == "" {
		return Report{}, fmt.Errorf("runtime root is required")
	}
	r := Report{SchemaVersion: "1", RuntimeRoot: filepath.Clean(runtimeRoot)}
	for _, target := range Discover(runtimeRoot) {
		r.Targets = append(r.Targets, inspectTarget(target))
	}
	return r, nil
}

func InspectInstalled() Report {
	r := Report{SchemaVersion: "1"}
	targets := DiscoverInstalled()
	for _, target := range targets {
		r.Targets = append(r.Targets, inspectTarget(target))
	}
	if len(targets) == 0 {
		r.Warnings = append(r.Warnings, "no approved runtime evidence found on local host")
	}
	return r
}

func inspectTarget(target Target) TargetReport {
	r := TargetReport{Target: target}
	info, err := os.Lstat(target.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			r.Error = err.Error()
		}
		return r
	}
	r.Exists = true
	r.Size = info.Size()
	r.Mode = info.Mode().String()
	if info.Mode()&os.ModeSymlink != 0 {
		r.Error = "symlink target rejected"
		return r
	}
	if info.IsDir() {
		r.FileType = "directory"
		return r
	}
	r.FileType = "file"
	b, err := os.ReadFile(target.Path)
	if err == nil {
		sum := sha256.Sum256(b)
		r.SHA256 = hex.EncodeToString(sum[:])
		if target.Kind == "binary" {
			r.WhatsAppStrings = whatsappStrings(b)
		}
	}
	if target.Kind == "binary" {
		if bi, err := buildinfo.ReadFile(target.Path); err == nil {
			r.GoVersion = bi.GoVersion
			for _, s := range bi.Settings {
				if s.Key == "vcs.revision" {
					r.VCSRevision = s.Value
				}
			}
		}
	}
	return r
}

func whatsappStrings(b []byte) []string {
	var out []string
	for i := 0; i < len(b); {
		for i < len(b) && (b[i] < 32 || b[i] > 126) {
			i++
		}
		start := i
		for i < len(b) && b[i] >= 32 && b[i] <= 126 {
			i++
		}
		if i-start >= 4 {
			s := string(b[start:i])
			if strings.Contains(strings.ToLower(s), "whatsapp") {
				out = append(out, s)
				if len(out) >= 32 {
					break
				}
			}
		}
	}
	return out
}
