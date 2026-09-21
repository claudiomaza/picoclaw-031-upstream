package metadata

import (
	"context"
	"testing"
)

func TestRequestIDRoundTrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), "run-123")
	if got := RequestID(ctx); got != "run-123" {
		t.Fatalf("got %q", got)
	}
}

func TestProviderMetadataIsCopied(t *testing.T) {
	original := map[string]string{"X-RoundRobin-Model": "gemini"}
	copied := New(original)
	original["X-RoundRobin-Model"] = "changed"
	if copied.Get("X-RoundRobin-Model") != "gemini" {
		t.Fatal("metadata was not copied")
	}
}
