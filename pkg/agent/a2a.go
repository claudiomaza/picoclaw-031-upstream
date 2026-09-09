package agent

import (
	"context"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/providers/protocoltypes"
)

// ProcessA2AMessage is the local cm2labs execution seam for callers that
// already own transport, sessions, persistence, and delivery. It deliberately
// does not create an A2A server or persist metrics.
func (al *AgentLoop) ProcessA2AMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
	if msg.Channel == "" {
		msg.Channel = "a2a"
	}
	return al.processMessage(protocoltypes.WithRequestID(ctx, msg.MessageID), msg)
}
