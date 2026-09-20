package tools

import (
	"context"
	"strings"
	"testing"
)

func TestHarnessCLIDelegatorUsesFixedConfiguration(t *testing.T) {
	var got []string
	h := HarnessCLI{Binary: "/usr/local/bin/agent-harness", ProfileDir: "/profiles", Routes: "hermes=hermes-a2a", PicoBaseURL: "http://pico", HermesBaseURL: "http://hermes", Run: func(_ context.Context, b string, args []string) ([]byte, error) {
		got = append([]string{b}, args...)
		return []byte(`{"status":"COMPLETED","runtime":{"text":"ok"}}`), nil
	}}
	text, err := h.DelegateRuntime(context.Background(), HarnessDelegationRequest{TargetRuntime: "hermes", Task: "inspect"})
	if err != nil || text != "ok" {
		t.Fatalf("text=%q err=%v", text, err)
	}
	joined := strings.Join(got, " ")
	if strings.Contains(joined, "token") || strings.Contains(joined, "8645") {
		t.Fatalf("unsafe args=%q", joined)
	}
}
