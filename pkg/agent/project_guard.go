package agent

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sipeed/picoclaw/pkg/logger"
)

// projectGuard records the working tree at turn start and verifies it again
// when the turn finishes. It also bootstraps a local repository for a new
// workspace without adding workspace files to the first commit.
type projectGuard struct {
	root   string
	before string
}

func beginProjectGuard(workspace string) (*projectGuard, error) {
	root := strings.TrimSpace(workspace)
	if root == "" {
		return nil, nil
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("project guard: resolve workspace: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("project guard: create workspace: %w", err)
	}
	if _, err := os.Stat(filepath.Join(abs, ".git")); os.IsNotExist(err) {
		if err := runGit(abs, "init"); err != nil {
			return nil, fmt.Errorf("project guard: git init: %w", err)
		}
		// Keep credentials and runtime state out of the project baseline.
		ignore := "# PicoClaw project guard\n.env\n.env.*\n!.env.example\n*.key\n*.pem\nsecrets/\nsessions/\nlogs/\ncache/\n.tmp/\n"
		if err := os.WriteFile(filepath.Join(abs, ".gitignore"), []byte(ignore), 0o600); err != nil {
			return nil, fmt.Errorf("project guard: write .gitignore: %w", err)
		}
		if err := runGit(abs, "add", ".gitignore"); err != nil {
			return nil, fmt.Errorf("project guard: stage baseline: %w", err)
		}
		if err := runGit(abs, "-c", "user.name=PicoClaw", "-c", "user.email=picoclaw@localhost", "commit", "-m", "chore: initialize project guard"); err != nil {
			return nil, fmt.Errorf("project guard: baseline commit: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("project guard: inspect git: %w", err)
	}
	status, err := gitOutput(abs, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return nil, fmt.Errorf("project guard: initial status: %w", err)
	}
	return &projectGuard{root: abs, before: status}, nil
}

func (g *projectGuard) Finish() {
	if g == nil {
		return
	}
	after, err := gitOutput(g.root, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		logger.WarnCF("project-guard", "Unable to verify project changes", map[string]any{"workspace": g.root, "error": err.Error()})
		return
	}
	if after == g.before {
		logger.InfoCF("project-guard", "No project changes detected", map[string]any{"workspace": g.root})
		return
	}
	logger.InfoCF("project-guard", "Verified project changes", map[string]any{"workspace": g.root, "before_status": strings.TrimSpace(g.before), "after_status": strings.TrimSpace(after)})
}

func runGit(root string, args ...string) error { _, err := gitOutput(root, args...); return err }
func gitOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), err
	}
	return out.String(), nil
}
