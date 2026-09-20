package tools

import (
	"context"
	toolshared "github.com/sipeed/picoclaw/pkg/tools/shared"
	"testing"
)

type fakeDelegator struct{ got HarnessDelegationRequest }

func (f *fakeDelegator) DelegateRuntime(_ context.Context, req HarnessDelegationRequest) (string, error) {
	f.got = req
	return "delegated-result", nil
}
func TestDelegateRuntimeToolUsesHarnessAndSessionContext(t *testing.T) {
	f := &fakeDelegator{}
	tool := NewDelegateRuntimeTool("picoclaw", "default", "/workspace/project", f)
	ctx := toolshared.WithToolSessionContext(context.Background(), "default", "parent-session", nil)
	out := tool.Execute(ctx, map[string]any{"target_runtime": "hermes", "task": "inspect tests"})
	if out.IsError || out.ForLLM != "delegated-result" {
		t.Fatalf("out=%#v", out)
	}
	if f.got.TargetRuntime != "hermes" || f.got.OriginRuntime != "picoclaw" || f.got.SessionID != "parent-session" {
		t.Fatalf("req=%#v", f.got)
	}
	if f.got.Repository != "/workspace/project" {
		t.Fatalf("repository=%q", f.got.Repository)
	}
}
func TestDelegateRuntimeToolRejectsMissingSession(t *testing.T) {
	tool := NewDelegateRuntimeTool("picoclaw", "default", "/repo", &fakeDelegator{})
	out := tool.Execute(context.Background(), map[string]any{"target_runtime": "hermes", "task": "inspect"})
	if !out.IsError {
		t.Fatal("expected missing session error")
	}
}
