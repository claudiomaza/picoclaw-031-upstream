package agent

import (
	"context"

	"github.com/sipeed/picoclaw/pkg/bus"
	a2amessage "github.com/sipeed/picoclaw/pkg/extensibility/a2a/message"
)

// ProcessA2AMessage is the local cm2labs execution seam for callers that
// already own transport, sessions, persistence, and delivery. It deliberately
// does not create an A2A server or persist metrics.
func (al *AgentLoop) ProcessA2AMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
	ctx, msg = a2amessage.Prepare(ctx, msg)
	return al.processMessage(ctx, msg)
}
