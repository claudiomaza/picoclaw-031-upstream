package agent

import (
	"context"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/providers/protocoltypes"
)

// ProcessA2AMessage runs one message without starting channel transports or notifications.
func (al *AgentLoop) ProcessA2AMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return al.processMessage(protocoltypes.WithRequestID(ctx, msg.MessageID), msg)
}
