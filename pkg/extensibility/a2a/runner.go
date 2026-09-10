package a2a

import (
	"context"
	"fmt"
	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/extensibility/profile"
	"github.com/sipeed/picoclaw/pkg/providers"
	"net/http"
)

type TurnRequest struct{ Operation, RunID, TraceID, SessionID, Agent, AgentID, Message string }
type TurnResult struct {
	Status, RunID, TraceID, SessionID, AgentID, Text string
	Error                                            string
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
	out, err := (&profile.Service{Loop: r.loop}).Execute(ctx, profile.Request{RunID: q.RunID, TraceID: q.TraceID, SessionID: q.SessionID, AgentID: q.AgentID, Message: q.Message})
	return TurnResult{Status: out.Status, RunID: out.RunID, TraceID: out.TraceID, SessionID: out.SessionID, AgentID: out.AgentID, Text: out.Text, Error: out.Error}, err
}

func (r *Runner) ListAgents() []profile.Descriptor {
	if r == nil || r.loop == nil {
		return nil
	}
	return (&profile.Service{Loop: r.loop}).ListAgents()
}

func (r *Runner) ProfileHandler() http.Handler {
	if r == nil || r.loop == nil {
		return profile.Handler{}
	}
	return profile.Handler{Service: &profile.Service{Loop: r.loop}}
}
