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
	RuntimeRoot   string         `json:"runtime_root"`
	Targets       []TargetReport `json:"targets"`
}

var allowedNames = map[string]bool{"picoclaw": true, "picoclaw-a2a": true, "picoclaw-edit": true, "picoclaw-service": true}

func Discover(runtimeRoot string) []Target {
	runtimeRoot = filepath.Clean(runtimeRoot)
	source := strings.TrimSuffix(runtimeRoot, "-edit") + "-edit"
	return []Target{
		{Name: "picoclaw", Path: filepath.Join(runtimeRoot, "bin", "picoclaw"), Kind: "binary"},
		{Name: "picoclaw-a2a", Path: filepath.Join(runtimeRoot, "picoclaw-a2a-prod"), Kind: "binary"},
		{Name: "picoclaw-edit", Path: source, Kind: "source"},
		{Name: "picoclaw-service", Path: "/etc/systemd/system/picoclaw.service", Kind: "unit"},
		{Name: "picoclaw-user-service", Path: "/home/ubuntu/.config/systemd/user/picoclaw.service", Kind: "unit"},
	}
}

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

func Inspect(runtimeRoot string) (Report, error) {
	if strings.TrimSpace(runtimeRoot) == "" {
		return Report{}, fmt.Errorf("runtime root is required")
	}
	runtimeRoot = filepath.Clean(runtimeRoot)
	r := Report{SchemaVersion: "1", RuntimeRoot: runtimeRoot}
	for _, target := range Discover(runtimeRoot) {
		r.Targets = append(r.Targets, inspectTarget(target))
	}
	return r, nil
}

func inspectTarget(target Target) TargetReport {
	r := TargetReport{Target: target}
	info, err := os.Lstat(target.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return r
		}
		r.Error = err.Error()
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
	if b, err := os.ReadFile(target.Path); err == nil {
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
