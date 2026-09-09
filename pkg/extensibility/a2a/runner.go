package a2a

import (
	"context"
	"fmt"

	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
)

type TurnRequest struct{ RunID, TraceID, SessionID, Agent, Message string }
type TurnResult struct {
	Status, RunID, TraceID, SessionID, Text string
	Error                                   string
}
type Runner struct{ loop *agent.AgentLoop }

func NewRunner(configPath string) (*Runner, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	provider, _, err := providers.CreateProvider(cfg)
	if err != nil {
		return nil, err
	}
	return &Runner{loop: agent.NewAgentLoop(cfg, bus.NewMessageBus(), provider)}, nil
}

func (r *Runner) Execute(ctx context.Context, q TurnRequest) (TurnResult, error) {
	if r == nil || r.loop == nil {
		return TurnResult{}, fmt.Errorf("a2a runner is not initialized")
	}
	if q.Message == "" {
		return TurnResult{}, fmt.Errorf("message is required")
	}
	session := q.SessionID
	if session == "" {
		session = "a2a-" + q.RunID
	}
	text, err := r.loop.ProcessA2AMessage(ctx, bus.InboundMessage{Content: q.Message, Channel: "a2a", ChatID: session, SenderID: q.Agent, MessageID: q.TraceID, SessionKey: session})
	result := TurnResult{Status: "COMPLETED", RunID: q.RunID, TraceID: q.TraceID, SessionID: session, Text: text}
	if err != nil {
		result.Status = "FAILED"
		result.Error = err.Error()
	}
	return result, err
}
