package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	toolshared "github.com/sipeed/picoclaw/pkg/tools/shared"
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
type DelegateRuntimeTool struct {
	OriginRuntime, OriginAgentID, Repository string
	Delegator                                HarnessDelegator
}

func NewDelegateRuntimeTool(originRuntime, originAgentID, repository string, delegator HarnessDelegator) *DelegateRuntimeTool {
	return &DelegateRuntimeTool{OriginRuntime: originRuntime, OriginAgentID: originAgentID, Repository: repository, Delegator: delegator}
}
func (t *DelegateRuntimeTool) Name() string { return DelegateRuntimeToolName }
func (t *DelegateRuntimeTool) Description() string {
	return "Delegate a task to another registered runtime through agent-harness. Use only for cross-runtime work; internal subtasks use the runtime's native subtask mechanism."
}
func (t *DelegateRuntimeTool) Parameters() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{"target_runtime": map[string]any{"type": "string"}, "task": map[string]any{"type": "string"}}, "required": []string{"target_runtime", "task"}}
}
func (t *DelegateRuntimeTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	target, ok1 := args["target_runtime"].(string)
	task, ok2 := args["task"].(string)
	if !ok1 || !ok2 || strings.TrimSpace(target) == "" || strings.TrimSpace(task) == "" {
		return ErrorResult("target_runtime and task are required")
	}
	if t.Delegator == nil {
		return ErrorResult("delegate_runtime is not configured")
	}
	session := toolshared.ToolSessionKey(ctx)
	if session == "" {
		return ErrorResult("delegate_runtime requires an active session")
	}
	id := session + "-" + strings.ToLower(strings.ReplaceAll(strings.TrimSpace(target), " ", "-"))
	req := HarnessDelegationRequest{OriginRuntime: t.OriginRuntime, OriginAgentID: t.OriginAgentID, SessionID: session, TargetRuntime: strings.TrimSpace(target), TargetSessionID: id, ParentWorkID: session, ParentTraceID: session + "-trace", DelegationID: id, Task: task, Repository: t.Repository}
	text, err := t.Delegator.DelegateRuntime(ctx, req)
	if err != nil {
		return ErrorResult(fmt.Sprintf("delegate_runtime failed: %v", err))
	}
	return toolshared.NewToolResult(text)
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
