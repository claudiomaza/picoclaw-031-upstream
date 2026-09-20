package tools

import (
	"context"
	"fmt"
	"strings"

	harness "github.com/sipeed/picoclaw/pkg/extensibility/harness"
	toolshared "github.com/sipeed/picoclaw/pkg/tools/shared"
)

const DelegateRuntimeToolName = harness.DelegateRuntimeToolName

type HarnessDelegationRequest = harness.HarnessDelegationRequest
type HarnessDelegator = harness.HarnessDelegator
type HarnessCLI = harness.HarnessCLI

func EncodeHarnessDelegationRequest(req HarnessDelegationRequest) ([]byte, error) {
	return harness.EncodeHarnessDelegationRequest(req)
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
