package a2a

import (
	"context"
	"fmt"
	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
	"net/url"
	"os"
)

type TurnRequest struct{ RunID, TraceID, SessionID, Agent, Message string }
type Metrics struct {
	Profile, Model, Provider, Node                              string
	Attempts, Failovers, InputTokens, OutputTokens, TotalTokens int
	Tools                                                       []string
}
type TurnResult struct {
	SessionID, Text string
	Metrics         Metrics
}
type Runner struct {
	loop    *agent.AgentLoop
	apiNode string
}

func NewRunner(path string) (*Runner, error) {
	cfg, e := config.LoadConfig(path)
	if e != nil {
		return nil, e
	}
	p, _, e := providers.CreateProvider(cfg)
	if e != nil {
		return nil, e
	}
	node := ""
	if len(cfg.ModelList) > 0 {
		if u, err := url.Parse(cfg.ModelList[0].APIBase); err == nil {
			node = u.Host
		}
	}
	return &Runner{loop: agent.NewAgentLoop(cfg, bus.NewMessageBus(), p), apiNode: node}, nil
}
func (r *Runner) Execute(ctx context.Context, q TurnRequest) (TurnResult, error) {
	if r == nil || r.loop == nil {
		return TurnResult{}, fmt.Errorf("a2a runner is not initialized")
	}
	if q.Message == "" {
		return TurnResult{}, fmt.Errorf("message is required")
	}
	s := q.SessionID
	if s == "" {
		s = "a2a-" + q.RunID
	}
	t, e := r.loop.ProcessA2AMessage(ctx, bus.InboundMessage{Content: q.Message, Channel: "a2a", ChatID: s, SenderID: q.Agent, MessageID: q.RunID, SessionKey: s})
	m := r.loop.LastA2AMetrics()
	if m.Provider == "openai" {
		m.Provider = "local-router (OpenAI-compatible)"
	}
	if m.Node == "" {
		m.Node = r.apiNode
	}
	if m.Node == "" {
		m.Node, _ = os.Hostname()
	}
	return TurnResult{SessionID: s, Text: t, Metrics: Metrics{m.Profile, m.Model, m.Provider, m.Node, m.Attempts, m.Failovers, m.InputTokens, m.OutputTokens, m.TotalTokens, m.Tools}}, e
}
