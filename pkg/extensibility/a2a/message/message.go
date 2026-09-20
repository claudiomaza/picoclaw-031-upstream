// Package message contains transport-neutral A2A message preparation.
package message

import (
	"context"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/providers/protocoltypes"
)

// Prepare applies the cm2labs A2A channel default and propagates the request ID.
// It does not execute, persist, or deliver the message.
func Prepare(ctx context.Context, msg bus.InboundMessage) (context.Context, bus.InboundMessage) {
	if msg.Channel == "" {
		msg.Channel = "a2a"
	}
	return protocoltypes.WithRequestID(ctx, msg.MessageID), msg
}
