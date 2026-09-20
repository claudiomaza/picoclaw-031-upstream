package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const DelegateRuntimeToolName = "delegate_runtime"

type HarnessDelegationRequest struct {
	OriginRuntime   string `json:"origin_runtime"`
	OriginAgentID   string `json:"origin_agent_id"`
	SessionID       string `json:"session_id"`
	TargetRuntime   string `json:"target_runtime"`
	TargetSessionID string `json:"target_session_id"`
	ParentWorkID    string `json:"parent_work_id"`
	ParentTraceID   string `json:"parent_trace_id"`
	DelegationID    string `json:"delegation_id"`
	Task            string `json:"task"`
	Repository      string `json:"repository"`
}

type HarnessDelegator interface {
	DelegateRuntime(context.Context, HarnessDelegationRequest) (string, error)
}

func EncodeHarnessDelegationRequest(req HarnessDelegationRequest) ([]byte, error) {
	return json.Marshal(req)
}

// HarnessCLI delegates through the cm2labs agent-harness CLI. Configuration is
// fixed at construction; request data cannot select endpoints or credentials.
type HarnessCLI struct {
	Binary, ProfileDir, Routes, PicoBaseURL, HermesBaseURL string
	Run                                                    func(context.Context, string, []string) ([]byte, error)
}

func (h HarnessCLI) DelegateRuntime(ctx context.Context, req HarnessDelegationRequest) (string, error) {
	if strings.TrimSpace(h.Binary) == "" || strings.TrimSpace(h.ProfileDir) == "" || strings.TrimSpace(h.Routes) == "" {
		return "", fmt.Errorf("harness CLI configuration is incomplete")
	}
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", "cm2labs-delegation-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "request.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", err
	}
	args := []string{"-delegate-request", path, "-profile-dir", h.ProfileDir, "-delegation-routes", h.Routes, "-pico-base-url", h.PicoBaseURL, "-hermes-base-url", h.HermesBaseURL}
	var out []byte
	if h.Run != nil {
		out, err = h.Run(ctx, h.Binary, args)
	} else {
		out, err = exec.CommandContext(ctx, h.Binary, args...).CombinedOutput()
	}
	if err != nil {
		return "", fmt.Errorf("harness delegation failed: %w", err)
	}
	var report struct {
		Status  string `json:"status"`
		Error   string `json:"error"`
		Runtime struct {
			Text string `json:"text"`
		} `json:"runtime"`
	}
	if err := json.Unmarshal(out, &report); err != nil {
		return "", fmt.Errorf("harness report: %w", err)
	}
	if report.Status != "COMPLETED" {
		return "", fmt.Errorf("harness delegation status %s: %s", report.Status, report.Error)
	}
	return report.Runtime.Text, nil
}
