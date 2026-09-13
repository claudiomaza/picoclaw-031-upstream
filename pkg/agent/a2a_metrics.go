package agent

type A2AMetrics struct {
	Profile, Model, Provider, Node         string
	Attempts, Failovers                    int
	InputTokens, OutputTokens, TotalTokens int
	Tools                                  []string
}

func (al *AgentLoop) setLastA2AMetrics(m A2AMetrics) {
	al.a2aMetricsMu.Lock()
	al.lastA2AMetrics = m
	al.a2aMetricsMu.Unlock()
}
func (al *AgentLoop) LastA2AMetrics() A2AMetrics {
	al.a2aMetricsMu.RLock()
	defer al.a2aMetricsMu.RUnlock()
	m := al.lastA2AMetrics
	m.Tools = append([]string(nil), m.Tools...)
	return m
}
