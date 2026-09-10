package profile

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler exposes the profile contract for a PicoClaw HTTP host.
// The host owns authentication and mounting; this handler owns only profile semantics.
type Handler struct{ Service *Service }

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.Service == nil {
		http.Error(w, `{"error":"profile service is not initialized"}`, http.StatusServiceUnavailable)
		return
	}
	switch {
	case r.Method == http.MethodGet && (r.URL.Path == "/health" || r.URL.Path == "/healthz"):
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case r.Method == http.MethodGet && r.URL.Path == "/v1/profile/agents":
		writeJSON(w, http.StatusOK, map[string]any{"agents": h.Service.ListAgents()})
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/profile/agents/"):
		id := strings.TrimPrefix(r.URL.Path, "/v1/profile/agents/")
		for _, a := range h.Service.ListAgents() {
			if a.ID == id {
				writeJSON(w, http.StatusOK, a)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "agent_id is not configured"})
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/turn"):
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/profile/agents/"), "/turn")
		var q Request
		if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		q.AgentID = id
		out, err := h.Service.Execute(r.Context(), q)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, out)
			return
		}
		writeJSON(w, http.StatusOK, out)
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "profile route not found"})
	}
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
