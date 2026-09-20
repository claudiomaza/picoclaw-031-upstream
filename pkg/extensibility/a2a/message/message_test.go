package message

import (
	"context"
	"testing"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/providers/protocoltypes"
)

func TestPrepareDefaultsChannelAndPropagatesRequestID(t *testing.T) {
	ctx, msg := Prepare(context.Background(), bus.InboundMessage{MessageID: "req-123"})
	if msg.Channel != "a2a" {
		t.Fatalf("channel = %q, want a2a", msg.Channel)
	}
	if got := protocoltypes.RequestID(ctx); got != "req-123" {
		t.Fatalf("request id = %q, want req-123", got)
	}
}

func TestPreparePreservesExplicitChannel(t *testing.T) {
	_, msg := Prepare(context.Background(), bus.InboundMessage{Channel: "custom", MessageID: "req-456"})
	if msg.Channel != "custom" {
		t.Fatalf("channel = %q, want custom", msg.Channel)
	}
}
