package profile

import (
	"context"
	"fmt"

	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/bus"
)

type Service struct{ Loop *agent.AgentLoop }
type Descriptor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
}
type Request struct {
	RunID     string `json:"run_id,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
	Message   string `json:"message"`
}
type Result struct {
	Status    string `json:"status"`
	RunID     string `json:"run_id,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
	Text      string `json:"text,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (s *Service) ListAgents() []Descriptor {
	if s == nil || s.Loop == nil {
		return nil
	}
	ds := s.Loop.GetRegistry().ListAgents("")
	out := make([]Descriptor, 0, len(ds))
	for _, d := range ds {
		out = append(out, Descriptor{ID: d.ID, Name: d.Name, Description: d.Description, Enabled: true})
	}
	return out
}
func (s *Service) Execute(ctx context.Context, q Request) (Result, error) {
	if s == nil || s.Loop == nil {
		return Result{Status: "FAILED", Error: "profile service is not initialized"}, fmt.Errorf("profile service is not initialized")
	}
	if q.Message == "" {
		return Result{Status: "FAILED", Error: "message is required"}, fmt.Errorf("message is required")
	}
	agentID := q.AgentID
	if agentID != "" {
		if _, ok := s.Loop.GetRegistry().GetAgent(agentID); !ok {
			return Result{Status: "FAILED", RunID: q.RunID, TraceID: q.TraceID, AgentID: agentID, Error: "agent_id is not configured"}, fmt.Errorf("agent_id %q is not configured", agentID)
		}
	}
	session := q.SessionID
	if session == "" {
		session = "a2a-" + q.RunID
	}
	raw := map[string]string{}
	if agentID != "" {
		raw["agent_id"] = agentID
	}
	text, err := s.Loop.ProcessA2AMessage(ctx, bus.InboundMessage{Content: q.Message, Channel: "a2a", ChatID: session, SenderID: "a2a", MessageID: q.TraceID, SessionKey: session, Context: bus.InboundContext{Raw: raw}})
	out := Result{Status: "COMPLETED", RunID: q.RunID, TraceID: q.TraceID, SessionID: session, AgentID: agentID, Text: text}
	if err != nil {
		out.Status = "FAILED"
		out.Error = err.Error()
	}
	return out, err
}
