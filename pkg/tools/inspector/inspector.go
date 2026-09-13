package inspector

import (
	"context"
	"encoding/json"

	ext "github.com/sipeed/picoclaw/pkg/extensions/inspector"
	toolshared "github.com/sipeed/picoclaw/pkg/tools/shared"
)

type Tool struct{ runtimeRoot string }

func New(runtimeRoot string) *Tool { return &Tool{runtimeRoot: runtimeRoot} }
func (t *Tool) Name() string       { return "infraestructura_inspector" }
func (t *Tool) Description() string {
	return "Read-only inspection of the discovered PicoClaw runtime using a fixed allowlist; never executes inspected binaries or reads secrets."
}
func (t *Tool) Parameters() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}
}
func (t *Tool) Execute(_ context.Context, _ map[string]any) *toolshared.ToolResult {
	root := t.runtimeRoot
	if root == "" {
		root = ext.DefaultRuntimeRoot()
	}
	r, err := ext.Inspect(root)
	if err != nil {
		return toolshared.ErrorResult(err.Error())
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return toolshared.NewToolResult(string(b))
}
